package controllers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	models "stock/models"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

const (
	authURL     = "https://auth.smsafrica.tech/auth/api-key"
	smsURL      = "https://sms-service.smsafrica.tech/message/send/transactional"
	callbackURL = "https://callback.io/123/dlr"
)

type TokenResponse struct {
	Token string `json:"token"`
}

type SmsRequest struct {
	Message     string `json:"message"`
	Msisdn      string `json:"msisdn"`
	SenderID    string `json:"sender_id"`
	CallbackURL string `json:"callback_url"`
}

// GetApiToken fetches an API token for authentication
func GetApiToken() (string, error) {
	username := "254700000000"                                                     // Replace with your username
	password := "e05b0e0d42c608dd08151cfc325da68f1eadd7bf60e457a043bc2e1de39635e2" // Replace with your password
	base64Credentials := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))

	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequest("GET", authURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Basic "+base64Credentials)

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusAccepted {
		var response TokenResponse
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read response body: %v", err)
		}
		if err := json.Unmarshal(body, &response); err != nil {
			return "", err
		}
		return response.Token, nil
	} else {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read response body: %v", err)
		}
		return "", fmt.Errorf("failed to get token. Status: %d, Response: %s", resp.StatusCode, body)
	}
}

// SendSms sends an SMS message
func SendSms(apiToken, message, phoneNumber string) error {
	smsRequest := SmsRequest{
		Message:     message,
		Msisdn:      phoneNumber,
		SenderID:    "SMSAFRICA",
		CallbackURL: callbackURL,
	}

	postData, err := json.Marshal(smsRequest)
	if err != nil {
		log.Printf("Error marshalling JSON: %v", err)
		return err
	}

	client := &http.Client{Timeout: 60 * time.Second}
	req, err := http.NewRequest("POST", smsURL, bytes.NewBuffer(postData))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", apiToken)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error sending SMS: %v", err)
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return err
	}

	log.Printf("Sending SMS to %s: %s", phoneNumber, message)
	log.Printf("HTTP Status Code: %d\nResponse: %s", resp.StatusCode, body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Error sending SMS. Status: %d, Response: %s", resp.StatusCode, body)
	}

	return nil
}

// SendSmsHandler handles SMS sending requests
func SendSmsHandler(c echo.Context) error {
	apiToken, err := GetApiToken()
	if err != nil {
		log.Printf("Failed to retrieve API token: %v", err)
		return c.String(http.StatusInternalServerError, "Failed to retrieve API token")
	}

	var smsRequest SmsRequest
	if err := c.Bind(&smsRequest); err != nil {
		log.Printf("Invalid JSON input: %v", err)
		return c.String(http.StatusBadRequest, "Invalid JSON input")
	}

	if err := SendSms(apiToken, smsRequest.Message, smsRequest.Msisdn); err != nil {
		log.Printf("Failed to send SMS: %v", err)
		return c.String(http.StatusInternalServerError, "Failed to send SMS")
	}

	return c.String(http.StatusOK, "SMS sent successfully")
}

// CheckStockLevelsAndLog checks stock levels and sends SMS for low stock alerts
func CheckStockLevelsAndLog(db *gorm.DB) {
	// Perform a join between 'products' and 'stock' tables to find low-stock items for any organization
	var stocks []models.Stock
	err := db.Table("stock").
		Joins("JOIN products ON products.product_id = stock.product_id").
		Where("stock.quantity <= products.reorder_level"). // Check if stock quantity is below reorder level
		Preload("Product").                                // Preload the related Product struct
		Find(&stocks).Error                                // Fetch all stocks that are below reorder level

	if err != nil {
		log.Printf("Error checking stock levels: %v", err)
		return
	}

	// If there are low-stock items, process and send notifications
	if len(stocks) > 0 {
		// Group stocks by organization to send specific notifications to each organization's users
		organizationStocks := make(map[uint][]models.Stock) // Map to group stocks by organization_id

		for _, stock := range stocks {
			organizationStocks[stock.OrganizationID] = append(organizationStocks[stock.OrganizationID], stock)
		}

		// Loop through each organization and send notifications
		for orgID, stocksForOrg := range organizationStocks {
			// Compose message for the organization
			message := fmt.Sprintf("Low stock alert for organization %d:\n", orgID)
			for _, stock := range stocksForOrg {
				message += fmt.Sprintf("Product '%s' (ID: %d) is below reorder level. Current stock: %d, Reorder level: %d.\n",
					stock.Product.ProductName, stock.Product.ProductID, stock.Quantity, stock.Product.ReorderLevel)
			}

			// Retrieve API token before sending SMS
			apiToken, err := GetApiToken()
			if err != nil {
				log.Printf("Error retrieving API token: %v", err)
				return
			}

			// Fetch all phone numbers from the users table for the current organization
			var users []models.User
			err = db.Model(&models.User{}).Where("is_active = 1").Where("organization_id = ?", orgID).Find(&users).Error
			if err != nil {
				log.Printf("Error fetching phone numbers from users table: %v", err)
				return
			}

			// Loop through the users and send SMS to each active user's phone number
			for _, user := range users {
				// Format the phone number with a "+" sign and country code (assuming '254' is the country code)
				phoneNumber := fmt.Sprintf("+%d", user.Phonenumber) // Convert the int64 phone number to a string
				if err := SendSms(apiToken, message, phoneNumber); err != nil {
					log.Printf("Error sending SMS to %s: %v", phoneNumber, err)
				} else {
					log.Printf("SMS sent to %s successfully", phoneNumber)
				}
			}
		}
	} else {
		log.Println("No products are below the reorder level for any organization.")
	}
}

// StartReorderLevelNotification starts a ticker to check stock levels every 30 seconds
func StartReorderLevelNotification(db *gorm.DB) {
	// Create a ticker that runs every 30 seconds
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Continuously check stock levels every 30 seconds
	for {
		select {
		case <-ticker.C:
			// Log the stock levels and check if any product is below the reorder level
			CheckStockLevelsAndLog(db) // Only pass db, no organizationID
		}
	}
}
