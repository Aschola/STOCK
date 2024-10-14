package controllers

import (
	"fmt"
	"log"
	"net/smtp"
	"strconv"
	"time"

	"stock/models"

	"github.com/labstack/echo/v4"
)

// Set your parameters here
const (
	smsSenderID = "SMSAFRICA"
	smtpServer  = "smtp.gmail.com"         // Gmail's SMTP server
	smtpPort    = "587"                    // SMTP port for TLS
	emailUser   = "pascalongeri@gmail.com" // Your email
	emailPass   = "@Eldoret2"              // Your email password (consider using an app password for security)
)

// Utility function for error responses
func respondWithError(c echo.Context, statusCode int, message string) error {
	log.Println(message)
	return echo.NewHTTPError(statusCode, message)
}

// GetProductsfornotification fetches all products and checks for reorder levels
func GetProductsfornotification() {
	log.Println("Starting product notification check...")
	ticker := time.NewTicker(1 * time.Minute) // Set the ticker to 1 minute
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("Checking products for reorder levels...")
			db := getDB()
			if db == nil {
				log.Println("Failed to connect to the database")
				continue // Skip this cycle if the database connection fails
			}

			var products []models.Product
			if err := db.Table("active_products").Find(&products).Error; err != nil {
				log.Printf("Error fetching products from the database: %v", err)
				continue // Skip this cycle if there's an error fetching products
			}

			log.Printf("Successfully fetched %d products", len(products))

			// Fetch all active users from users
			var users []models.User
			if err := db.Table("users").Where("is_active = ?", 1).Find(&users).Error; err != nil {
				log.Printf("Error fetching users from the database: %v", err)
				continue
			}

			// Check for products needing reordering
			notificationSent := false
			for _, product := range products {
				if product.Quantity <= product.ReorderLevel {
					sendNotification(product, users) // Pass users to sendNotification
					notificationSent = true
				}
			}

			if !notificationSent {
				log.Println("No products are below or at their reorder level.")
			}
		}
	}
}

// SendNotification logs a notification for a product reaching its reorder level and sends SMS and Email
func sendNotification(product models.Product, users []models.User) {
	message := fmt.Sprintf("Alert: Product '%s' (ID: %d) is below reorder level. Current Quantity: %d", product.ProductName, product.ProductID, product.Quantity)
	log.Printf("Notification: %s\n", message)

	// Send SMS notifications to all users
	apiToken, err := GetApiToken()
	if err != nil {
		log.Printf("Failed to retrieve API token for SMS: %v", err)
		return
	}

	for _, user := range users {
		Phonenumber := user.Phonenumber
		phoneNumberStr := strconv.FormatInt(Phonenumber, 10) // Convert int64 to string
		SendSms(apiToken, message, phoneNumberStr)           // Send SMS to each user's phone number

		// Send Email notification
		if err := sendEmailNotification(user.Email, message); err != nil {
			log.Printf("Failed to send email to %s: %v", user.Email, err)
		}
	}
}

// sendEmailNotification sends an email to the specified recipient with the given message
func sendEmailNotification(to string, message string) error {
	auth := smtp.PlainAuth("", emailUser, emailPass, smtpServer)

	msg := []byte("To: " + to + "\r\n" +
		"Subject: Product Reorder Notification\r\n" +
		"\r\n" +
		message + "\r\n")

	err := smtp.SendMail(smtpServer+":"+smtpPort, auth, emailUser, []string{to}, msg)
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	log.Printf("Email sent to %s", to)
	return nil
}
