// package controllers

// import (
// 	"bytes"
// 	"encoding/base64"
// 	"encoding/json"
// 	"fmt"
// 	"net/http"
// 	"log"
// 	"time"

// 	"github.com/google/uuid"
// 	"github.com/labstack/echo/v4"
// 	"stock/db"
// )

// // MPesaSettings represents organization-specific credentials
// type MPesaSettings struct {
// 	ID                int64  `json:"id" gorm:"primaryKey"`
// 	OrganizationID    int64  `json:"organization_id"`
// 	ConsumerKey       string `json:"consumer_key"`
// 	ConsumerSecret    string `json:"consumer_secret"`
// 	BusinessShortCode string `json:"business_short_code"`
// 	PassKey           string `json:"pass_key"`
// 	CallbackURL       string `json:"callback_url"`
// 	Environment       string `json:"environment"` // sandbox or production
// }

// func (MPesaSettings) TableName() string {
// 	return "mpesasettings" // Explicitly set the table name to "mpesasettings"
// }

// // STKPushRequest represents the incoming request
// type STKPushRequest struct {
// 	PhoneNumber   string  `json:"phone_number"`
// 	Amount        float64 `json:"amount"`
// 	TransactionID string  `json:"transaction_id,omitempty"`
// }

// // STKPushResponse represents Safaricom's response
// type STKPushResponse struct {
// 	MerchantRequestID   string `json:"MerchantRequestID"`
// 	CheckoutRequestID   string `json:"CheckoutRequestID"`
// 	ResponseCode        string `json:"ResponseCode"`
// 	ResponseDescription string `json:"ResponseDescription"`
// 	CustomerMessage     string `json:"CustomerMessage"`
// 	TransactionID       string `json:"TransactionID"`
// }

// // TransactionRecord represents the database model for MPesa transactions
// type TransactionRecord struct {
// 	ID                int64     `json:"id" gorm:"primaryKey"`
// 	TransactionID     string    `json:"transaction_id"`
// 	OrganizationID    int64     `json:"organization_id"`
// 	PhoneNumber       string    `json:"phone_number"`
// 	Amount            float64   `json:"amount"`
// 	MerchantRequestID string    `json:"merchant_request_id"`
// 	CheckoutRequestID string    `json:"checkout_request_id"`
// 	Status            string    `json:"status"`
// 	CreatedAt         time.Time `json:"created_at"`
// 	UpdatedAt         time.Time `json:"updated_at"`
// }

// func (TransactionRecord) TableName() string {
// 	return "mpesa_transactions" // Explicitly set the table name to "mpesa_transactions"
// }

// // generateTransactionID creates a unique transaction ID
// func generateTransactionID() string {
// 	return fmt.Sprintf("MPE-%s-%d", uuid.New().String()[:8], time.Now().Unix())
// }

// // HandleSTKPush processes the STK push request
// func HandleSTKPush(c echo.Context) error {
// 	// Parse request body
// 	var req STKPushRequest
// 	if err := c.Bind(&req); err != nil {
// 		log.Println("Error binding request:", err)
// 		return c.JSON(http.StatusBadRequest, map[string]string{
// 			"error": "Invalid request format",
// 		})
// 	}

// 	// Retrieve organization ID from context
// 	organizationID, ok := c.Get("organizationID").(uint)
// 	if !ok {
// 		log.Println("Failed to get organizationID from context")
// 		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
// 	}

// 	// Validate request
// 	if req.PhoneNumber == "" {
// 		return c.JSON(http.StatusBadRequest, map[string]string{
// 			"error": "Phone number is required",
// 		})
// 	}
// 	if req.Amount <= 0 {
// 		return c.JSON(http.StatusBadRequest, map[string]string{
// 			"error": "Amount must be greater than 0",
// 		})
// 	}

// 	// Process STK Push
// 	response, err := InitiateSTKPush(int64(organizationID), req)
// 	if err != nil {
// 		log.Println("Error initiating STK push:", err)
// 		return c.JSON(http.StatusInternalServerError, map[string]string{
// 			"error": err.Error(),
// 		})
// 	}

// 	return c.JSON(http.StatusOK, response)
// }

// // InitiateSTKPush handles the STK push request
// func InitiateSTKPush(organizationID int64, req STKPushRequest) (*STKPushResponse, error) {
// 	// Load MPesa credentials for the given organization
// 	creds, err := loadMPesaCredentials(organizationID)
// 	if err != nil {
// 		log.Println("Error loading MPesa credentials:", err)
// 		return nil, fmt.Errorf("failed to load MPesa credentials: %v", err)
// 	}

// 	// Generate transaction ID if not provided
// 	if req.TransactionID == "" {
// 		req.TransactionID = generateTransactionID()
// 	}

// 	token, err := getAccessToken(creds)
// 	if err != nil {
// 		log.Println("Error getting access token:", err)
// 		return nil, fmt.Errorf("failed to get access token: %v", err)
// 	}

// 	// Prepare STK Push request
// 	timestamp := time.Now().Format("20060102150405")
// 	password := generatePassword(creds.BusinessShortCode, creds.PassKey, timestamp)

// 	stkURL := "https://sandbox.safaricom.co.ke/mpesa/stkpush/v1/processrequest"
// 	if creds.Environment == "production" {
// 		stkURL = "https://api.safaricom.co.ke/mpesa/stkpush/v1/processrequest"
// 	}

// 	requestBody := map[string]interface{}{
// 		"BusinessShortCode": creds.BusinessShortCode,
// 		"Password":          password,
// 		"Timestamp":         timestamp,
// 		"TransactionType":   "CustomerPayBillOnline",
// 		"Amount":            req.Amount,
// 		"PartyA":            req.PhoneNumber,
// 		"PartyB":            creds.BusinessShortCode,
// 		"PhoneNumber":       req.PhoneNumber,
// 		"CallBackURL":       creds.CallbackURL,
// 		"AccountReference":  req.TransactionID,
// 		"TransactionDesc":   fmt.Sprintf("Payment %s", req.TransactionID),
// 	}

// 	jsonBody, err := json.Marshal(requestBody)
// 	if err != nil {
// 		log.Println("Error marshalling request body:", err)
// 		return nil, fmt.Errorf("failed to marshal request body: %v", err)
// 	}

// 	// Make HTTP request
// 	httpReq, err := http.NewRequest("POST", stkURL, bytes.NewBuffer(jsonBody))
// 	if err != nil {
// 		log.Println("Error creating HTTP request:", err)
// 		return nil, err
// 	}

// 	httpReq.Header.Set("Authorization", "Bearer "+token)
// 	httpReq.Header.Set("Content-Type", "application/json")

// 	client := &http.Client{}
// 	resp, err := client.Do(httpReq)
// 	if err != nil {
// 		log.Println("Error making HTTP request:", err)
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	var stkResponse STKPushResponse
// 	if err := json.NewDecoder(resp.Body).Decode(&stkResponse); err != nil {
// 		log.Println("Error decoding response:", err)
// 		return nil, err
// 	}

// 	// Add the transaction ID to the response
// 	stkResponse.TransactionID = req.TransactionID

// 	// Save transaction record to database (using mpesa_transactions table)
// 	record := TransactionRecord{
// 		TransactionID:     req.TransactionID,
// 		OrganizationID:    organizationID,
// 		PhoneNumber:       req.PhoneNumber,
// 		Amount:            req.Amount,
// 		MerchantRequestID: stkResponse.MerchantRequestID,
// 		CheckoutRequestID: stkResponse.CheckoutRequestID,
// 		Status:            "PENDING",
// 		CreatedAt:         time.Now(),
// 		UpdatedAt:         time.Now(),
// 	}

// 	if err := db.GetDB().Create(&record).Error; err != nil {
// 		log.Println("Error saving transaction record:", err)
// 		return nil, fmt.Errorf("failed to save transaction record: %v", err)
// 	}

// 	return &stkResponse, nil
// }

// // loadMPesaCredentials fetches MPesa credentials for the given organization
// func loadMPesaCredentials(organizationID int64) (MPesaSettings, error) {
// 	var creds MPesaSettings
// 	err := db.GetDB().Where("organization_id = ?", organizationID).First(&creds).Error
// 	if err != nil {
// 		log.Println("Error fetching MPesa credentials from DB:", err)
// 		return creds, fmt.Errorf("could not find credentials for organization ID %d: %v", organizationID, err)
// 	}
// 	return creds, nil
// }

// // getAccessToken generates an OAuth access token
// func getAccessToken(creds MPesaSettings) (string, error) {
// 	authURL := "https://sandbox.safaricom.co.ke/oauth/v1/generate"
// 	if creds.Environment == "production" {
// 		authURL = "https://api.safaricom.co.ke/oauth/v1/generate"
// 	}

// 	auth := base64.StdEncoding.EncodeToString([]byte(creds.ConsumerKey + ":" + creds.ConsumerSecret))

// 	req, err := http.NewRequest("GET", authURL, nil)
// 	if err != nil {
// 		log.Println("Error creating access token request:", err)
// 		return "", err
// 	}

// 	req.Header.Add("Authorization", "Basic "+auth)
// 	req.Header.Add("Cache-Control", "no-cache")

// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		log.Println("Error making access token request:", err)
// 		return "", err
// 	}
// 	defer resp.Body.Close()

// 	var result struct {
// 		AccessToken string `json:"access_token"`
// 	}

// 	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
// 		log.Println("Error decoding access token response:", err)
// 		return "", err
// 	}

// 	return result.AccessToken, nil
// }

// // generatePassword creates the MPesa API password
// func generatePassword(shortCode, passKey, timestamp string) string {
// 	str := shortCode + passKey + timestamp
// 	return base64.StdEncoding.EncodeToString([]byte(str))
// }
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

// STKPushResponse represents Safaricom's response
type STKPushResponse struct {
	MerchantRequestID   string `json:"MerchantRequestID"`
	CheckoutRequestID   string `json:"CheckoutRequestID"`
	ResponseCode        string `json:"ResponseCode"`
	ResponseDescription string `json:"ResponseDescription"`
	CustomerMessage     string `json:"CustomerMessage"`
	TransactionID       string `json:"TransactionID"`
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
}

func (TransactionRecord) TableName() string {
	return "mpesa_transactions"
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

	// Load MPesa credentials for the given organization
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

	// Get access token
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

// Helper function to mask sensitive strings
func maskString(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return s[:4] + "****" + s[len(s)-4:]
}

// Helper function to format phone number
func formatPhoneNumber(phone string) string {
	// Remove any spaces or special characters
	phone = strings.ReplaceAll(phone, " ", "")
	phone = strings.ReplaceAll(phone, "-", "")
	phone = strings.ReplaceAll(phone, "+", "")

	// If number starts with 0, replace with 254
	if strings.HasPrefix(phone, "0") {
		phone = "254" + phone[1:]
	}

	// If number doesn't start with 254, add it
	if !strings.HasPrefix(phone, "254") {
		phone = "254" + phone
	}

	return phone
}

// HandleMpesaCallback processes M-Pesa callback
func HandleMpesaCallback(c echo.Context) error {
	var callbackData map[string]interface{}
	if err := c.Bind(&callbackData); err != nil {
		log.Printf("Error binding callback data: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid callback data",
		})
	}

	log.Printf("Received M-Pesa callback: %+v", callbackData)

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Callback processed successfully",
	})
}