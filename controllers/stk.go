package controllers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"gopkg.in/natefinch/lumberjack.v2"
	"runtime/debug" 
	"github.com/joho/godotenv"
    "os"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"stock/db"
	"stock/models"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// MPesaSettings represents organization-specific credentials
type MPesaSettings struct {
	ID                int64  `json:"id" gorm:"primaryKey"`
	OrganizationID    int64  `json:"organization_id"`
	ConsumerKey       string `json:"consumer_key"`
	ConsumerSecret    string `json:"consumer_secret"`
	BusinessShortCode int64  `json:"business_short_code"`
	PassKey           string `json:"pass_key"`
	Password          string `json:"password"`
	CallbackURL       string `json:"-" gorm:"-"` 
	// CallbackURL       string `json:"callback_url"`
	Environment       string `json:"environment"`
}

func (MPesaSettings) TableName() string {
	return "mpesasettings"
}

// STKPushRequest represents the incoming request
type STKPushRequest struct {
	PhoneNumber    int64   `json:"phone_number"`
	Amount         float64 `json:"amount"`
	TransactionID  string  `json:"transaction_id,omitempty"`
	OrganizationID int64
}

type STKPushResponse struct {
	MerchantRequestID   string    `json:"MerchantRequestID"`
	CheckoutRequestID   string    `json:"CheckoutRequestID"`
	ResponseCode        string    `json:"ResponseCode"`
	ResponseDescription string    `json:"ResponseDescription"`
	CustomerMessage     string    `json:"CustomerMessage"`
	TransactionID       string    `json:"TransactionID"`
	CallbackReceived    time.Time `json:"callback_received"`
}

func (STKPushResponse) TableName() string {
	return "mpesa_callbacks"
}

// TransactionRecord represents the database model for MPesa transactions
type TransactionRecord struct {
	ID                int64     `json:"id" gorm:"primaryKey"`
	TransactionID     string    `json:"transaction_id"`
	OrganizationID    int64     `json:"organization_id"`
	PhoneNumber       string    `json:"phone_number"`
	Amount            float64   `json:"amount"`
	MerchantRequestID string    `json:"merchant_request_id"`
	CheckoutRequestID string    `json:"checkout_request_id"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	MpesaReceipt      string    `json:"mpesa_receipt"`
}


func (TransactionRecord) TableName() string {
	return "mpesa_transactions"
}

type MpesaCallbackResponse struct {
	Body struct {
		StkCallback struct {
			MerchantRequestID string `json:"MerchantRequestID"`
			CheckoutRequestID string `json:"CheckoutRequestID"`
			ResultCode        int    `json:"ResultCode"`
			ResultDesc        string `json:"ResultDesc"`
			CallbackMetadata  struct {
				Item []struct {
					Name  string      `json:"Name"`
					Value interface{} `json:"Value"`
				} `json:"Item"`
			} `json:"CallbackMetadata"`
		} `json:"stkCallback"`
	} `json:"Body"`
}

func generateTransactionID() string {
	return fmt.Sprintf("MPE-%s-%d", uuid.New().String()[:8], time.Now().Unix())
}

// HandleSTKPush processes the STK push request
func HandleSTKPush(c echo.Context) error {
	var req STKPushRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("Error binding request: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}

	log.Printf("Received STK Push request: %+v", req)

	organizationID, ok := c.Get("organizationID").(uint)
	if !ok {
		log.Printf("Failed to get organizationID from context")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	// organizationID, ok := c.Get("organizationID").(uint)
	// if !ok {
	// 	log.Printf("Failed to get organizationID from context")
	// 	return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	// }
	// log.Printf("Processing request for organization ID: %d", organizationID)

	// log.Printf("Processing request for organization ID: %d", organizationID)

	// Validate request
	if req.PhoneNumber == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Phone number is required",
		})
	}
	if req.Amount <= 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Amount must be greater than 0",
		})
	}

	// Process STK Push
	response, err := InitiateSTKPush(int64(organizationID), req)
	if err != nil {
		log.Printf("Error initiating STK push: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response)
}


func loadMPesaCredentials(organizationID int64) (MPesaSettings, error) {
    var creds MPesaSettings

    // Debug: Print the SQL query
    query := db.GetDB().Where("organization_id = ?", organizationID)
    log.Printf("SQL Query: %v", query.Statement.SQL.String())

    err := query.First(&creds).Error
    if err != nil {
        log.Printf("Error fetching MPesa credentials from DB: %v", err)
        return creds, fmt.Errorf("could not find credentials for organization ID %d: %v", organizationID, err)
    }

    // Debug: Log all credential fields
    log.Printf("Retrieved credentials from database: %+v", creds)

    // Validate required fields
    if creds.ConsumerKey == "" {
        return creds, fmt.Errorf("consumer key is missing")
    }
    if creds.ConsumerSecret == "" {
        return creds, fmt.Errorf("consumer secret is missing")
    }
    if creds.BusinessShortCode == 0 {
        return creds, fmt.Errorf("business short code is missing")
    }
    if creds.PassKey == "" {
        return creds, fmt.Errorf("pass key is missing")
    }

    // Fetch CallbackURL from environment variable
    err = godotenv.Load() 
	log.Printf("MPESA_CALLBACK_URL loaded")
    if err != nil {
        log.Printf("Error loading .env file: %v", err)
        return creds, fmt.Errorf("could not load environment variables")
    }

    creds.CallbackURL = os.Getenv("MPESA_CALLBACK_URL")
    if creds.CallbackURL == "" {
		log.Printf("Mpesa callback: %v", creds)
        return creds, fmt.Errorf("callback URL is missing from environment variables")
    }

    return creds, nil
}

// getAccessToken generates an OAuth access token
func getAccessToken(creds MPesaSettings) (string, error) {
	// authURL := "https://sandbox.safaricom.co.ke/oauth/v1/generate?grant_type=client_credentials"
	authURL := "https://api.safaricom.co.ke/oauth/v1/generate?grant_type=client_credentials"

	auth := base64.StdEncoding.EncodeToString([]byte(creds.ConsumerKey + ":" + creds.ConsumerSecret))

	req, err := http.NewRequest("GET", authURL, nil)
	if err != nil {
		log.Printf("Error creating access token request: %v", err)
		return "", err
	}

	req.Header.Add("Authorization", "Basic "+auth)
	req.Header.Add("Cache-Control", "no-cache")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error making access token request: %v", err)
		return "", err
	}
	defer resp.Body.Close()

	// Read the raw response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading access token response: %v", err)
		return "", err
	}

	log.Printf("Access token response: %s", string(body))

	// Parse the response

	var result struct {
		AccessToken  string `json:"access_token"`
		ErrorCode    string `json:"errorCode"`
		ErrorMessage string `json:"errorMessage"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("Error decoding access token response: %v", err)
		return "", err
	}

	if result.AccessToken == "" {
		log.Printf("Error: %s - %s", result.ErrorCode, result.ErrorMessage)
		return "", fmt.Errorf("%s: %s", result.ErrorCode, result.ErrorMessage)
	}

	return result.AccessToken, nil
}

// generatePassword creates the MPesa API password
func generatePassword(shortCode, passKey, timestamp string) string {
	str := shortCode + passKey + timestamp
	return base64.StdEncoding.EncodeToString([]byte(str))
}

// InitiateSTKPush handles the STK push request
func InitiateSTKPush(organizationID int64, req STKPushRequest) (*STKPushResponse, error) {
	
	log.Printf("Starting STK Push for organization %d with request: %+v", organizationID, req)
	logger := log.New(&lumberjack.Logger{
		Filename:   "/var/www/STOCK/logs/callbacks.log",
		MaxSize:    100, 
		MaxBackups: 3,
		MaxAge:     28,   
		Compress:   true,
	}, "", log.LstdFlags|log.Lshortfile)
	
	logger = log.New(os.Stdout, "[Initiate stkpush] ", log.LstdFlags|log.Lshortfile)


	creds, err := loadMPesaCredentials(organizationID)
	if err != nil {
		log.Printf("Error loading MPesa credentials: %v", err)
		return nil, fmt.Errorf("failed to load MPesa credentials: %v", err)
	}

	// Debug: Log credentials (excluding sensitive data)
	logger.Printf("Using credentials: BusinessShortCode=%s, Environment=%s, CallbackURL=%s",
		creds.BusinessShortCode, creds.Environment, creds.CallbackURL)

	//Generate transaction ID if not provided
	if req.TransactionID == "" {
		req.TransactionID = generateTransactionID()
		logger.Printf("Generated transaction ID: %s", req.TransactionID)
	}

	token, err := getAccessToken(creds)
	
	if err != nil {
		log.Printf("Failed to get access token: %v", err)
		return nil, fmt.Errorf("failed to get access token: %v", err)
	}

	if token != "" {
		logger.Printf("Successfully obtained access token: %s", maskString(token))
		
	} else {
		log.Printf("Warning: Received empty access token")
		return nil, fmt.Errorf("received empty access token")
	}

	// Prepare STK Push request
	timestamp := time.Now().Format("20060102150405")
	password := generatePassword(fmt.Sprintf("%d", creds.BusinessShortCode), creds.PassKey, timestamp)

	// stkURL := "https://sandbox.safaricom.co.ke/mpesa/stkpush/v1/processrequest"
	stkURL := "https://api.safaricom.co.ke/mpesa/stkpush/v1/processrequest"

	log.Printf("Using STK Push URL: %s", stkURL)
	// Format phone number
	//phoneNumber := formatPhoneNumber(req.PhoneNumber)
	phoneNumber := formatPhoneNumber(strconv.FormatInt(req.PhoneNumber, 10))
	logger.Printf("Formatted phone number: %s", phoneNumber)

	requestBody := map[string]interface{}{
		"BusinessShortCode": creds.BusinessShortCode,
		"Password":          password,
		"Timestamp":         timestamp,
		"TransactionType":   "CustomerPayBillOnline",
		"Amount":            req.Amount,
		"PartyA":            phoneNumber,
		"PartyB":            creds.BusinessShortCode,
		"PhoneNumber":       phoneNumber,
		"CallBackURL":       creds.CallbackURL,
		"AccountReference":  "PATNER COMMUNICATION",
		"TransactionDesc":   fmt.Sprintf("Payment %s", req.TransactionID),
	}
	logger.Printf("request body: %v", requestBody)

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("Error marshalling request body: %v", err)
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	// Log request body (excluding sensitive data)
	logBody := make(map[string]interface{})
	json.Unmarshal(jsonBody, &logBody)
	logBody["Password"] = "REDACTED"
	logBodyBytes, _ := json.Marshal(logBody)
	logger.Printf("STK Push request body: %s", string(logBodyBytes))

	// Make HTTP request
	httpReq, err := http.NewRequest("POST", stkURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Printf("Error creating HTTP request: %v", err)
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+token)
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		log.Printf("Error making HTTP request: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Read the raw response
	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return nil, err
	}

	logger.Printf("Raw response from Safaricom: %s", string(rawBody))
	logger.Printf("Response status code: %d", resp.StatusCode)

	// Try to decode the response
	var stkResponse STKPushResponse
	if err := json.Unmarshal(rawBody, &stkResponse); err != nil {
		log.Printf("Error decoding response: %v", err)
		return nil, err
	}

	// Add the transaction ID to the response
	stkResponse.TransactionID = req.TransactionID

	// Save transaction record
	record := TransactionRecord{
		TransactionID:     req.TransactionID,
		OrganizationID:    organizationID,
		PhoneNumber:       phoneNumber,
		Amount:            req.Amount,
		MerchantRequestID: stkResponse.MerchantRequestID,
		CheckoutRequestID: stkResponse.CheckoutRequestID,
		Status:            "PENDING",
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	if err := db.GetDB().Create(&record).Error; err != nil {
		log.Printf("Error saving transaction record: %v", err)
		return nil, fmt.Errorf("failed to save transaction record: %v", err)
	}

	logger.Printf("Successfully saved transaction record with ID: %d", record.ID)
	return &stkResponse, nil
	
}

func HandleMpesaCallback(c echo.Context) error {
	log.Printf("callback triggered")
	logger := log.New(&lumberjack.Logger{
		Filename:   "/var/www/STOCK/logs/callbacks.log",
		MaxSize:    100, 
		MaxBackups: 3,
		MaxAge:     28,   
		Compress:   true,
	}, "", log.LstdFlags|log.Lshortfile)
	
	// Initialize logger with timestamp and caller info
	logger = log.New(os.Stdout, "[MPESA-CALLBACK] ", log.LstdFlags|log.Lshortfile)
	
	// Log initial request details
	logger.Printf("=== START PROCESSING MPESA CALLBACK ===")
	logger.Printf("Request Details - Method: %s, URL: %s, RemoteAddr: %s", 
		c.Request().Method, 
		c.Request().URL, 
		c.Request().RemoteAddr)
	logger.Printf("Request Headers: %+v", c.Request().Header)

	// Read and log request body
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		logger.Printf(" ERROR: Failed to read request body: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to read request body",
		})
	}
	logger.Printf("📥 Raw callback body: %s", string(body))

	// Restore body
	c.Request().Body = io.NopCloser(bytes.NewBuffer(body))

	// Parse callback
	var callbackResp MpesaCallbackResponse
	if err := json.Unmarshal(body, &callbackResp); err != nil {
		logger.Printf("❌ ERROR: Failed to unmarshal callback: %v", err)
		logger.Printf("❌ Failed body content: %s", string(body))
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid callback format",
		})
	}
	logger.Printf("✅ Successfully parsed callback response")

	// Extract and validate callback data
	stkCallback := callbackResp.Body.StkCallback
	logger.Printf("📋 Processing STK callback - MerchantRequestID: %s, CheckoutRequestID: %s", 
		stkCallback.MerchantRequestID, 
		stkCallback.CheckoutRequestID)

	// Check for duplicate callback
	var existingCallback STKPushResponse
	if err := db.GetDB().Where("merchant_request_id = ?", stkCallback.MerchantRequestID).
		First(&existingCallback).Error; err == nil {
		logger.Printf("⚠️ Duplicate callback detected for MerchantRequestID: %s", stkCallback.MerchantRequestID)
		logger.Printf("⚠️ Previous callback processed at: %v", existingCallback.CallbackReceived)
		return c.JSON(http.StatusOK, map[string]string{
			"message": "Callback already processed",
		})
	}

	// Start transaction
	logger.Printf("🔄 Starting database transaction")
	tx := db.GetDB().Begin()
	if tx.Error != nil {
		logger.Printf("❌ ERROR: Failed to begin transaction: %v", tx.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Database error",
		})
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			logger.Printf("❌ PANIC RECOVERED: %v", r)
			logger.Printf("❌ Stack trace: %s", debug.Stack())
		}
	}()

	// Find transaction with lock
	logger.Printf("🔍 Searching for transaction record...")
	var transaction TransactionRecord
	result := tx.Set("gorm:query_option", "FOR UPDATE").
		Where("merchant_request_id = ? OR checkout_request_id = ?",
			stkCallback.MerchantRequestID, stkCallback.CheckoutRequestID).
		First(&transaction)

	if result.Error != nil {
		tx.Rollback()
		logger.Printf("❌ ERROR: Transaction not found - MID: %s, CID: %s, Error: %v",
			stkCallback.MerchantRequestID, stkCallback.CheckoutRequestID, result.Error)
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Transaction not found",
		})
	}
	logger.Printf("✅ Found transaction record: %+v", transaction)

	// Handle failed transaction
	if stkCallback.ResultCode != 0 {
		logger.Printf("⚠️ Transaction failed with ResultCode: %d, ResultDesc: %s",
			stkCallback.ResultCode, stkCallback.ResultDesc)
		
		// Update transaction status
		transaction.Status = "FAILED"
		transaction.UpdatedAt = time.Now()
		logger.Printf("🔄 Updating transaction status to FAILED")

		if err := tx.Save(&transaction).Error; err != nil {
			tx.Rollback()
			logger.Printf("❌ ERROR: Failed to update transaction status: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to update transaction",
			})
		}

		// Update sales status
		logger.Printf("🔄 Updating related sales records...")
		var sales []models.Sale
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			Table("sales_transactions").
			Where("transaction_id = ?", transaction.TransactionID).
			Find(&sales).Error; err != nil {
			tx.Rollback()
			logger.Printf("❌ ERROR: Failed to fetch sales records: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to fetch sales",
			})
		}
		logger.Printf("📊 Found %d sales records to update", len(sales))

		for i, sale := range sales {
			logger.Printf("🔄 Updating sale record %d/%d (ID: %d)", i+1, len(sales), sale.SaleID)
			sale.TransactionStatus = "FAILED"
			if err := tx.Save(&sale).Error; err != nil {
				tx.Rollback()
				logger.Printf("❌ ERROR: Failed to update sale status: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Failed to update sale status",
				})
			}
		}

		// Save callback details
		logger.Printf("💾 Saving failed callback details")
		callback := STKPushResponse{
			MerchantRequestID:   stkCallback.MerchantRequestID,
			CheckoutRequestID:   stkCallback.CheckoutRequestID,
			ResponseCode:        fmt.Sprintf("%d", stkCallback.ResultCode),
			ResponseDescription: stkCallback.ResultDesc,
			CustomerMessage:     "Transaction failed",
			TransactionID:       transaction.TransactionID,
			CallbackReceived:    time.Now(),
		}

		if err := tx.Create(&callback).Error; err != nil {
			tx.Rollback()
			logger.Printf("❌ ERROR: Failed to save callback data: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to save callback data",
			})
		}

		if err := tx.Commit().Error; err != nil {
			logger.Printf("❌ ERROR: Failed to commit transaction: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to commit transaction",
			})
		}

		logger.Printf("✅ Failed transaction processed successfully")
		return c.JSON(http.StatusOK, map[string]string{
			"message": "Transaction failed",
			"reason":  stkCallback.ResultDesc,
		})
	}

	// Process successful transaction
	logger.Printf("🔄 Processing successful transaction")
	mpesaReceipt := ""
	amount := 0.0
	metadata := stkCallback.CallbackMetadata

	logger.Printf("📋 Processing callback metadata items:")
	for _, item := range metadata.Item {
		logger.Printf("  - Item Name: %s, Value: %v", item.Name, item.Value)
		if item.Name == "MpesaReceiptNumber" {
			mpesaReceipt = item.Value.(string)
			logger.Printf("✅ Found MPesa receipt number: %s", mpesaReceipt)
		}
		if item.Name == "Amount" {
			amount = item.Value.(float64)
			logger.Printf("✅ Found transaction amount: %.2f", amount)
		}
	}

	// Validate metadata
	if mpesaReceipt == "" || amount == 0 {
		tx.Rollback()
		logger.Printf("❌ ERROR: Invalid metadata - Receipt: %s, Amount: %.2f", mpesaReceipt, amount)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid callback metadata",
		})
	}

	// Update transaction
	logger.Printf("🔄 Updating transaction record with success details")
	transaction.Status = "COMPLETED"
	transaction.Amount = amount
	transaction.UpdatedAt = time.Now()
	transaction.MpesaReceipt = mpesaReceipt

	if err := tx.Save(&transaction).Error; err != nil {
		tx.Rollback()
		logger.Printf("❌ ERROR: Failed to update transaction: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to update transaction",
		})
	}

	// Update sales and stock
	logger.Printf("🔄 Updating sales and stock records")
	var sales []models.Sale
	if err := tx.Set("gorm:query_option", "FOR UPDATE").
		Table("sales_transactions").
		Where("transaction_id = ?", transaction.TransactionID).
		Find(&sales).Error; err != nil {
		tx.Rollback()
		logger.Printf("❌ ERROR: Failed to fetch sales: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to fetch sales",
		})
	}
	logger.Printf("📊 Found %d sales records to process", len(sales))

	for i, sale := range sales {
		logger.Printf("🔄 Processing sale %d/%d (ID: %d, ProductID: %d)", i+1, len(sales), sale.SaleID, sale.ProductID)
		
		// Update sale status
		sale.TransactionStatus = "COMPLETED"
		if err := tx.Save(&sale).Error; err != nil {
			tx.Rollback()
			logger.Printf("❌ ERROR: Failed to update sale status: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to update sale status",
			})
		}

		// Update stock
		var stock models.Stock
		logger.Printf("🔍 Fetching stock for ProductID: %d", sale.ProductID)
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			First(&stock, "product_id = ?", sale.ProductID).Error; err != nil {
			tx.Rollback()
			logger.Printf("❌ ERROR: Failed to fetch stock: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to fetch stock",
			})
		}

		logger.Printf("📊 Current stock quantity: %d, Reducing by: %d", stock.Quantity, sale.Quantity)
		if stock.Quantity < sale.Quantity {
			tx.Rollback()
			logger.Printf("❌ ERROR: Insufficient stock - Required: %d, Available: %d", sale.Quantity, stock.Quantity)
			return c.JSON(http.StatusConflict, map[string]string{
				"error": fmt.Sprintf("Insufficient stock for product %d", sale.ProductID),
			})
		}

		stock.Quantity -= sale.Quantity
		if err := tx.Save(&stock).Error; err != nil {
			tx.Rollback()
			logger.Printf("❌ ERROR: Failed to update stock: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to update stock",
			})
		}
		logger.Printf("✅ Updated stock quantity to: %d", stock.Quantity)
	}

	// Save success callback
	logger.Printf("💾 Saving successful callback details")
	callback := STKPushResponse{
		MerchantRequestID:   stkCallback.MerchantRequestID,
		CheckoutRequestID:   stkCallback.CheckoutRequestID,
		ResponseCode:        fmt.Sprintf("%d", stkCallback.ResultCode),
		ResponseDescription: stkCallback.ResultDesc,
		CustomerMessage:     fmt.Sprintf("Receipt: %s, Amount: %.2f", mpesaReceipt, amount),
		TransactionID:       transaction.TransactionID,
		CallbackReceived:    time.Now(),
	}

	if err := tx.Create(&callback).Error; err != nil {
		tx.Rollback()
		logger.Printf("❌ ERROR: Failed to save callback data: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to save callback data",
		})
	}

	if err := tx.Commit().Error; err != nil {
		logger.Printf("❌ ERROR: Failed to commit transaction: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to commit transaction",
		})
	}

	logger.Printf("✅ SUCCESS: Callback processed successfully")
	logger.Printf("=== END PROCESSING MPESA CALLBACK ===")
	
	return c.JSON(http.StatusOK, map[string]string{
		"message": "Callback processed successfully",
	})
}
func maskString(s string) string {
	if len(s) <= 4 {
		return "**"
	}
	return s[:4] + "**" + s[len(s)-4:]
}

// Helper function to format phone number
func formatPhoneNumber(phone string) string {
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, "+", "")

	if strings.HasPrefix(phone, "0") {
		phone = "254" + phone[1:]
	}

	// If number doesn't start with 254, add it
	if !strings.HasPrefix(phone, "254") {
		phone = "254" + phone
	}

	return phone
}

func AddMPesaSettings(c echo.Context) error {
	organizationID, ok := c.Get("organizationID").(uint)
	if !ok {
		log.Println("AddMPesaSettings - Failed to get organizationID from context")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	log.Printf("AddMPesaSettings - OrganizationID: %d", organizationID)

	var settings MPesaSettings
	if err := c.Bind(&settings); err != nil {
		log.Printf("AddMPesaSettings - Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request payload"})
	}

	// Set the organization ID from the context
	settings.OrganizationID = int64(organizationID)

	// Save to the database
	if err := db.GetDB().Create(&settings).Error; err != nil {
		log.Printf("AddMPesaSettings - Error saving settings: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to save MPesa settings"})
	}

	log.Printf("AddMPesaSettings - MPesaSettings saved: %+v", settings)
	return c.JSON(http.StatusCreated, settings)
}

// EditMPesaSettings handles the update of existing MPesa settings.
func EditMPesaSettings(c echo.Context) error {
	organizationID, ok := c.Get("organizationID").(uint)
	if !ok {
		log.Println("EditMPesaSettings - Failed to get organizationID from context")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	log.Printf("EditMPesaSettings - OrganizationID: %d", organizationID)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("EditMPesaSettings - Invalid ID: %v", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid ID"})
	}
	log.Printf("EditMPesaSettings - Entry with ID: %d", id)

	var settings MPesaSettings
	if err := c.Bind(&settings); err != nil {
		log.Printf("EditMPesaSettings - Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid request payload"})
	}

	// Check if the settings exist and belong to the organization
	var existingSettings MPesaSettings
	if err := db.GetDB().Where("id = ? AND organization_id = ?", id, organizationID).First(&existingSettings).Error; err != nil {
		log.Printf("EditMPesaSettings - Settings not found: %v", err)
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Settings not found or not authorized"})
	}

	// Update settings
	existingSettings.ConsumerKey = settings.ConsumerKey
	existingSettings.ConsumerSecret = settings.ConsumerSecret
	existingSettings.BusinessShortCode = settings.BusinessShortCode
	existingSettings.PassKey = settings.PassKey
	existingSettings.Password = settings.Password
	existingSettings.CallbackURL = settings.CallbackURL
	existingSettings.Environment = settings.Environment

	if err := db.GetDB().Save(&existingSettings).Error; err != nil {
		log.Printf("EditMPesaSettings - Error updating settings: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to update MPesa settings"})
	}

	log.Printf("EditMPesaSettings - MPesaSettings updated: %+v", existingSettings)
	return c.JSON(http.StatusOK, existingSettings)
}

// GetMPesaSettingsByOrganization retrieves MPesa settings for the authenticated organization.
func GetMPesaSettingsByOrganization(c echo.Context) error {
	organizationID, ok := c.Get("organizationID").(uint)
	if !ok {
		log.Println("GetMPesaSettingsByOrganization - Failed to get organizationID from context")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}
	log.Printf("GetMPesaSettingsByOrganization - OrganizationID: %d", organizationID)

	var settings []MPesaSettings
	if err := db.GetDB().Where("organization_id = ?", organizationID).Find(&settings).Error; err != nil {
		log.Printf("GetMPesaSettingsByOrganization - Error retrieving settings: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to retrieve MPesa settings"})
	}

	if len(settings) == 0 {
		log.Println("GetMPesaSettingsByOrganization - No settings found")
		return c.JSON(http.StatusNotFound, echo.Map{"error": "No settings found for this organization"})
	}

	log.Printf("GetMPesaSettingsByOrganization - Settings found: %+v", settings)
	return c.JSON(http.StatusOK, settings)
}

func GetAllTransactions(c echo.Context) error {
	log.Println("GetAllTransactionsPerOrganization - Entry")

	// Retrieve organizationID from context
	organizationID, ok := c.Get("organizationID").(uint)
	if !ok {
		log.Println("GetAllTransactionsPerOrganization - Failed to get organizationID from context")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	log.Printf("GetAllTransactionsPerOrganization - OrganizationID: %d", organizationID)

	// Fetch all transactions for the given organization
	var transactions []TransactionRecord
	if err := db.GetDB().Where("organization_id = ?", organizationID).Find(&transactions).Error; err != nil {
		log.Printf("GetAllTransactionsPerOrganization - Error fetching transactions: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch transactions"})
	}

	log.Printf("GetAllTransactionsPerOrganization - Retrieved %d transactions", len(transactions))
	log.Println("GetAllTransactionsPerOrganization - Exit")
	return c.JSON(http.StatusOK, transactions)
}


