package controllers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
	"strconv"


	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"stock/db"
)

// MPesaSettings represents organization-specific credentials
type MPesaSettings struct {
	ID                int64  `json:"id" gorm:"primaryKey"`
	OrganizationID    int64  `json:"organization_id"`
	ConsumerKey       string `json:"consumer_key"`
	ConsumerSecret    string `json:"consumer_secret"`
	BusinessShortCode int64 `json:"business_short_code"`
	PassKey           string `json:"pass_key"`
	Password          string `json:"password"`
	CallbackURL       string `json:"callback_url"`
	Environment       string `json:"environment"` 
}

func (MPesaSettings) TableName() string {
	return "mpesasettings"
}

// STKPushRequest represents the incoming request
type STKPushRequest struct {
	PhoneNumber   string  `json:"phone_number"`
	Amount        float64 `json:"amount"`
	TransactionID string  `json:"transaction_id,omitempty"`
}

type STKPushResponse struct {
	MerchantRequestID   string `json:"MerchantRequestID"`
	CheckoutRequestID   string `json:"CheckoutRequestID"`
	ResponseCode        string `json:"ResponseCode"`
	ResponseDescription string `json:"ResponseDescription"`
	CustomerMessage     string `json:"CustomerMessage"`
	TransactionID       string `json:"TransactionID"`
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
	MpesaReceipt       string    `json:"mpesa_receipt"` 
}

func (TransactionRecord) TableName() string {
	return "mpesa_transactions"
}
type MpesaCallbackResponse struct {
    Body struct {
        StkCallback struct {
            MerchantRequestID  string `json:"MerchantRequestID"`
            CheckoutRequestID  string `json:"CheckoutRequestID"`
            ResultCode         int    `json:"ResultCode"`
            ResultDesc         string `json:"ResultDesc"`
            CallbackMetadata   struct {
                Item []struct {
                    Name  string      `json:"Name"`
                    Value interface{} `json:"Value"`
                } `json:"Item"`
            } `json:"CallbackMetadata"`
        } `json:"stkCallback"`
    } `json:"Body"`
}

// generateTransactionID creates a unique transaction ID
func generateTransactionID() string {
	return fmt.Sprintf("MPE-%s-%d", uuid.New().String()[:8], time.Now().Unix())
}

// HandleSTKPush processes the STK push request
func HandleSTKPush(c echo.Context) error {
	// Parse request body
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

	log.Printf("Processing request for organization ID: %d", organizationID)

	// Validate request
	if req.PhoneNumber == "" {
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

// loadMPesaCredentials fetches MPesa credentials for the given organization
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
	if creds.CallbackURL == "" {
		return creds, fmt.Errorf("callback URL is missing")
	}

	return creds, nil
}

// getAccessToken generates an OAuth access token
func getAccessToken(creds MPesaSettings) (string, error) {
	authURL := "https://sandbox.safaricom.co.ke/oauth/v1/generate?grant_type=client_credentials"

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
		AccessToken string `json:"access_token"`
		ErrorCode   string `json:"errorCode"`
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

	creds, err := loadMPesaCredentials(organizationID)
	if err != nil {
		log.Printf("Error loading MPesa credentials: %v", err)
		return nil, fmt.Errorf("failed to load MPesa credentials: %v", err)
	}

	// Debug: Log credentials (excluding sensitive data)
	log.Printf("Using credentials: BusinessShortCode=%s, Environment=%s, CallbackURL=%s",
		creds.BusinessShortCode, creds.Environment, creds.CallbackURL)

	// Generate transaction ID if not provided
	if req.TransactionID == "" {
		req.TransactionID = generateTransactionID()
		log.Printf("Generated transaction ID: %s", req.TransactionID)
	}

	token, err := getAccessToken(creds)
	if err != nil {
		log.Printf("Failed to get access token: %v", err)
		return nil, fmt.Errorf("failed to get access token: %v", err)
	}

	if token != "" {
		log.Printf("Successfully obtained access token: %s", maskString(token))
	} else {
		log.Printf("Warning: Received empty access token")
		return nil, fmt.Errorf("received empty access token")
	}

	// Prepare STK Push request
	timestamp := time.Now().Format("20060102150405")
	password := generatePassword(fmt.Sprintf("%d", creds.BusinessShortCode), creds.PassKey, timestamp)

	stkURL := "https://sandbox.safaricom.co.ke/mpesa/stkpush/v1/processrequest"
	log.Printf("Using STK Push URL: %s", stkURL)

	// Format phone number
	phoneNumber := formatPhoneNumber(req.PhoneNumber)
	log.Printf("Formatted phone number: %s", phoneNumber)

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
		"AccountReference":  req.TransactionID,
		"TransactionDesc":   fmt.Sprintf("Payment %s", req.TransactionID),
	}
	log.Printf("request body: %v", requestBody)

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
	log.Printf("STK Push request body: %s", string(logBodyBytes))

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

	log.Printf("Raw response from Safaricom: %s", string(rawBody))
	log.Printf("Response status code: %d", resp.StatusCode)

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

	log.Printf("Successfully saved transaction record with ID: %d", record.ID)
	return &stkResponse, nil
	
}
func HandleMpesaCallback(c echo.Context) error {
    log.Printf("Request Method: %s, URL: %s", c.Request().Method, c.Request().URL)
    log.Printf("Headers: %+v", c.Request().Header)

    // Read the request body
    body, err := io.ReadAll(c.Request().Body)
    if err != nil {
        log.Printf("Failed to read request body: %v", err)
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "Failed to read request body",
        })
    }
    log.Printf("Raw callback body: %s", string(body))

    // Restore the request body for further processing
    c.Request().Body = io.NopCloser(bytes.NewBuffer(body))

    // Parse the callback response
    var callbackResp MpesaCallbackResponse
    if err := json.Unmarshal(body, &callbackResp); err != nil {
        log.Printf("Failed to unmarshal callback response: %v", err)
        log.Printf("Failed body content: %s", string(body))
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "Invalid callback format",
        })
    }
    log.Printf("Callback response parsed: %+v", callbackResp)

    // Extract metadata from the callback
    stkCallback := callbackResp.Body.StkCallback
    log.Printf("Processing STK callback: %+v", stkCallback)

    // Find the corresponding transaction
    var transaction TransactionRecord
    result := db.GetDB().Where("merchant_request_id = ? OR checkout_request_id = ?",
        stkCallback.MerchantRequestID, stkCallback.CheckoutRequestID).First(&transaction)

    if result.Error != nil {
        log.Printf("Transaction not found - MerchantRequestID: %s, CheckoutRequestID: %s, Error: %v",
            stkCallback.MerchantRequestID, stkCallback.CheckoutRequestID, result.Error)
        return c.JSON(http.StatusNotFound, map[string]string{
            "error": "Transaction not found",
        })
    }
    log.Printf("Found transaction: %+v", transaction)

    // Check if the transaction failed
    if stkCallback.ResultCode != 0 {
        log.Printf("Transaction failed with ResultCode: %d, ResultDesc: %s", 
            stkCallback.ResultCode, stkCallback.ResultDesc)

        // Update the transaction as FAILED
        transaction.Status = "FAILED"
        transaction.UpdatedAt = time.Now()

        if err := db.GetDB().Save(&transaction).Error; err != nil {
            log.Printf("Failed to update failed transaction: %v", err)
            return c.JSON(http.StatusInternalServerError, map[string]string{
                "error": "Failed to update transaction",
            })
        }
        log.Printf("Transaction status updated to FAILED: %+v", transaction)

        // Save the callback details
        callback := STKPushResponse{
            MerchantRequestID:   stkCallback.MerchantRequestID,
            CheckoutRequestID:   stkCallback.CheckoutRequestID,
            ResponseCode:        fmt.Sprintf("%d", stkCallback.ResultCode),
            ResponseDescription: stkCallback.ResultDesc,
            CustomerMessage:     "Transaction failed",
        }

        if err := db.GetDB().Create(&callback).Error; err != nil {
            log.Printf("Failed to save callback data: %v", err)
            return c.JSON(http.StatusInternalServerError, map[string]string{
                "error": "Failed to save callback data",
            })
        }

        log.Println("Failed transaction recorded successfully")
        return c.JSON(http.StatusOK, map[string]string{
            "message": "Transaction failed",
            "reason":  stkCallback.ResultDesc,
        })
    }

    // If transaction was successful, process normally
    mpesaReceipt := ""
    amount := 0.0
    log.Println("Processing successful transaction metadata")
    for _, item := range stkCallback.CallbackMetadata.Item {
        if item.Name == "MpesaReceiptNumber" {
            mpesaReceipt = item.Value.(string)
            log.Printf("Mpesa receipt number found: %s", mpesaReceipt)
        }
        if item.Name == "Amount" {
            amount = item.Value.(float64)
            log.Printf("Transaction amount found: %.2f", amount)
        }
    }

    // Update the transaction as COMPLETED
    transaction.Status = "COMPLETED"
    transaction.Amount = amount
    transaction.UpdatedAt = time.Now()
    transaction.MpesaReceipt = mpesaReceipt

    if err := db.GetDB().Save(&transaction).Error; err != nil {
        log.Printf("Failed to update completed transaction: %v", err)
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "Failed to update transaction",
        })
    }
    log.Printf("Transaction updated successfully as COMPLETED: %+v", transaction)

    // Save the callback details
    callback := STKPushResponse{
        MerchantRequestID:   stkCallback.MerchantRequestID,
        CheckoutRequestID:   stkCallback.CheckoutRequestID,
        ResponseCode:        fmt.Sprintf("%d", stkCallback.ResultCode),
        ResponseDescription: stkCallback.ResultDesc,
        CustomerMessage:     fmt.Sprintf("Receipt: %s, Amount: %.2f", mpesaReceipt, amount),
    }

    if err := db.GetDB().Create(&callback).Error; err != nil {
        log.Printf("Failed to save success callback data: %v", err)
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "Failed to save callback data",
        })
    }
    log.Printf("Success callback data saved: %+v", callback)

    return c.JSON(http.StatusOK, map[string]string{
        "message": "Callback processed successfully",
    })
}
func maskString(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return s[:4] + "****" + s[len(s)-4:]
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

// func GetCompletedTransactions(c echo.Context) error {
// 	log.Println("GetCompletedTransactions - Entry")

// 	// Retrieve organizationID from context
// 	organizationID, ok := c.Get("organizationID").(uint)
// 	if !ok {
// 		log.Println("GetCompletedTransactions - Failed to get organizationID from context")
// 		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
// 	}

// 	log.Printf("GetCompletedTransactions - OrganizationID: %d", organizationID)

// 	// Fetch completed transactions for the given organization
// 	var transactions []TransactionRecord
// 	if err := db.GetDB().
// 		Where("organization_id = ? AND status = ?", organizationID, "COMPLETED").
// 		Find(&transactions).Error; err != nil {
// 		log.Printf("GetCompletedTransactions - Error fetching transactions: %v", err)
// 		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to fetch transactions"})
// 	}

// 	log.Printf("GetCompletedTransactions - Retrieved %d completed transactions", len(transactions))
// 	log.Println("GetCompletedTransactions - Exit")
// 	return c.JSON(http.StatusOK, transactions)
// }

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
