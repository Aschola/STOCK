package controllers

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"time"
	"fmt"
	"stock/db"
	"github.com/google/uuid"
	"encoding/base64"

	"github.com/labstack/echo/v4"
)
type MPesaSettingss struct {
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
func (MPesaSettingss) TableName() string {
	return "mpesasettings"
}

type STKPushRequests struct {
	BusinessShortCode string `json:"BusinessShortCode"`
	Password         string `json:"Password"`
	Timestamp        string `json:"Timestamp"`
	TransactionType  string `json:"TransactionType"`
	Amount           int `json:"Amount"`
	PartyA           string `json:"PartyA"`
	PartyB           string `json:"PartyB"`
	PhoneNumber      int `json:"PhoneNumber"`
	CallBackURL      string `json:"CallBackURL"`
	AccountReference string `json:"AccountReference"`
	TransactionDesc  string `json:"TransactionDesc"`
}

// type STKPushResponses struct {
// 	MerchantRequestID   string `json:"MerchantRequestID"`
// 	CheckoutRequestID   string `json:"CheckoutRequestID"`
// 	ResponseCode        string `json:"ResponseCode"`
// 	ResponseDescription string `json:"ResponseDescription"`
// 	CustomerMessage     string `json:"CustomerMessage"`
// }
type STKPushResponses struct {
	MerchantRequestID   string `json:"MerchantRequestID"`
	CheckoutRequestID   string `json:"CheckoutRequestID"`
	//ResponseCode        string `json:"ResponseCode"`
	ResponseDescription string `json:"ResponseDescription"`
	CustomerMessage     string `json:"CustomerMessage"`
	TransactionID       string `json:"TransactionID"`
	ResultCode         int    `json:"ResultCode"`
	CallbackReceived    time.Time `json:"callback_received"`
	
}
func (STKPushResponses) TableName() string {
	return "mpesa_callbacks"
}


type MpesaCallbackResponses struct {
	Body struct {
		StkCallback struct {
			MerchantRequestID   string `json:"MerchantRequestID"`
			CheckoutRequestID   string `json:"CheckoutRequestID"`
			ResultCode          int    `json:"ResultCode"`
			ResultDesc          string `json:"ResultDesc"`
			CallbackMetadata    *struct {
				Item []struct {
					Name  string      `json:"Name"`
					Value interface{} `json:"Value"`
				} `json:"Item"`
			} `json:"CallbackMetadata,omitempty"`
		} `json:"stkCallback"`
	} `json:"Body"`
}
func generateTransactionIDs() string {
	return fmt.Sprintf("MPE-%s-%d", uuid.New().String()[:8], time.Now().Unix())
}

// HandleSTKPush processes the STK push request
func HandleSTKPushs(c echo.Context) error {
	// Parse request body
	var req STKPushRequests
	if err := c.Bind(&req); err != nil {
		log.Printf("Error binding request: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}

	log.Printf("Received STK Push request: %+v", req)

	// organizationID, ok := c.Get("organizationID").(uint)
	// if !ok {
	// 	log.Printf("Failed to get organizationID from context")
	// 	return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	// }
	organizationID, ok := c.Get("organizationID").(uint)
if !ok {
    log.Printf("Failed to get organizationID from context")
    return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
}
log.Printf("Processing request for organization ID: %d", organizationID)

	log.Printf("Processing request for organization ID: %d", organizationID)

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
	response, err := InitiateSTKPushs(int64(organizationID), req)
	if err != nil {
		log.Printf("Error initiating STK push: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response)
}

// loadMPesaCredentials fetches MPesa credentials for the given organization
func loadMPesaCredentialss(organizationID int64) (MPesaSettingss, error) {
	var creds MPesaSettingss
	
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
func getAccessTokens(creds MPesaSettingss) (string, error) {
	//authURL := "https://sandbox.safaricom.co.ke/oauth/v1/generate?grant_type=client_credentials"
	authURL := "https:api.safaricom.co.ke/mpesa/stkpush/v1/processrequest"


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
func generatePasswords(shortCode, passKey, timestamp string) string {
	str := shortCode + passKey + timestamp
	return base64.StdEncoding.EncodeToString([]byte(str))
}
func InitiateSTKPushs(organizationID int64, req STKPushRequests) (*STKPushResponses, error) {
	// Load MPesa credentials for the organization
	creds, err := loadMPesaCredentialss(organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to load credentials: %v", err)
	}

	// Generate access token
	accessToken, err := getAccessTokens(creds)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %v", err)
	}

	// Prepare STK push request payload
	timestamp := time.Now().Format("20060102150405")
	password := generatePasswords(
		strconv.Itoa(int(creds.BusinessShortCode)), 
		creds.PassKey, 
		timestamp,
	)

	payload := map[string]interface{}{
		"BusinessShortCode": creds.BusinessShortCode,
		"Password":          password,
		"Timestamp":         timestamp,
		"TransactionType":   "CustomerPayBillOnline",
		"Amount":            req.Amount,
		"PartyA":            req.PhoneNumber,
		"PartyB":            creds.BusinessShortCode,
		"PhoneNumber":       req.PhoneNumber,
		"CallBackURL":       creds.CallbackURL,
		"AccountReference":  req.AccountReference,
		"TransactionDesc":   req.TransactionDesc,
	}

	// Convert payload to JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %v", err)
	}

	// Create STK push request
	// stkPushURL := "https://sandbox.safaricom.co.ke/mpesa/stkpush/v1/processrequest"
	stkPushURL := "https:api.safaricom.co.ke/mpesa/stkpush/v1/processrequest"

	httpReq, err := http.NewRequest("POST", stkPushURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+accessToken)

	// Send request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// Parse response
	var stkResponse STKPushResponses
	if err := json.Unmarshal(body, &stkResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	go func(checkoutRequestID string) {
		log.Printf("callback processed")
		// Prepare context for HandleMpesaCallbacks
		e := echo.New()
		req, _ := http.NewRequest(http.MethodPost, "/mpesa/call-backs", 
			bytes.NewBuffer([]byte{}))
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Call HandleMpesaCallbacks with the mock context and callback
		HandleMpesaCallback(c)
	}(stkResponse.CheckoutRequestID)

	return &stkResponse, nil
}
func HandleMpesaCallbackss(c echo.Context) error {
	startTime := time.Now()
	log.Printf("[MPESA CALLBACK] Processing started at %s", startTime.Format(time.RFC3339))

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		log.Printf("[MPESA CALLBACK] ERROR - Failed to read request body: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to read request body",
		})
	}

	c.Request().Body = io.NopCloser(bytes.NewBuffer(body))

	var callbackResp MpesaCallbackResponses
	if err := json.Unmarshal(body, &callbackResp); err != nil {
		log.Printf("[MPESA CALLBACK] ERROR - Failed to unmarshal callback response: %v", err)
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid callback format",
		})
	}

	stkCallback := callbackResp.Body.StkCallback
	
	log.Printf("[MPESA CALLBACK] Received - MerchantRequestID: %s, ResultCode: %d, ResultDesc: %s", 
		stkCallback.MerchantRequestID, stkCallback.ResultCode, stkCallback.ResultDesc)

	// Transaction unsuccessful cases
	if stkCallback.ResultCode != 0 {
		log.Printf("[MPESA CALLBACK] Transaction failed. ResultCode: %d, Description: %s", 
			stkCallback.ResultCode, stkCallback.ResultDesc)

		return c.JSON(http.StatusOK, map[string]interface{}{
			"Body": map[string]interface{}{
				"stkCallback": map[string]interface{}{
					"MerchantRequestID": stkCallback.MerchantRequestID,
					"CheckoutRequestID": stkCallback.CheckoutRequestID,
					"ResultCode":        stkCallback.ResultCode,
					"ResultDesc":        stkCallback.ResultDesc,
				},
			},
		})
	}

	// Check if metadata exists
	if len(stkCallback.CallbackMetadata.Item) == 0 {
		log.Printf("[MPESA CALLBACK] Successful transaction, but no metadata")
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Missing transaction metadata",
		})
	}

	var (
		amount         float64
		mpesaReceipt   string
		transactionDate string
		phoneNumber    string
	)

	for _, item := range stkCallback.CallbackMetadata.Item {
		switch item.Name {
		case "Amount":
			amount = getFloatValue(item.Value)
		case "MpesaReceiptNumber":
			mpesaReceipt = getStringValue(item.Value)
		case "TransactionDate":
			transactionDate = getStringValue(item.Value)
		case "PhoneNumber":
			phoneNumber = getStringValue(item.Value)
		}
	}

	log.Printf("[MPESA CALLBACK] Successful transaction. Amount: %.2f, Receipt: %s", amount, mpesaReceipt)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"Body": map[string]interface{}{
			"stkCallback": map[string]interface{}{
				"MerchantRequestID": stkCallback.MerchantRequestID,
				"CheckoutRequestID": stkCallback.CheckoutRequestID,
				"ResultCode":        stkCallback.ResultCode,
				"ResultDesc":        stkCallback.ResultDesc,
				"CallbackMetadata": map[string]interface{}{
					"Item": []map[string]interface{}{
						{"Name": "Amount", "Value": amount},
						{"Name": "MpesaReceiptNumber", "Value": mpesaReceipt},
						{"Name": "TransactionDate", "Value": transactionDate},
						{"Name": "PhoneNumber", "Value": phoneNumber},
					},
				},
			},
		},
	})
}


// Helper functions remain the same as in previous implementation
func getFloatValue(val interface{}) float64 {
	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	default:
		log.Printf("[MPESA CALLBACK] Unexpected float value type: %T", val)
		return 0
	}
}

func getStringValue(val interface{}) string {
	switch v := val.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		log.Printf("[MPESA CALLBACK] Unexpected string value type: %T", val)
		return ""
	}
}