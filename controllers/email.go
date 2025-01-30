package controllers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/smtp"
	"stock/db"
	"stock/models"
	"time"
	"os"

	"github.com/labstack/echo/v4"
)

// Configuration struct
type EmailConfig struct {
    SMTPHost     string
    SMTPPort     string
    SMTPPassword string
    FromEmail    string
}

// Initialize email config
var emailConfig EmailConfig
    // SMTPHost:     "smtp.gmail.com",
    // SMTPPort:     "587",
    // SMTPPassword: "Ngatia@01",
    // FromEmail:    "smsafrica@infinitytechafrica.com",

	func InitEmailConfig() error {
		emailConfig = EmailConfig{
			SMTPHost:     os.Getenv("SMTP_HOST"),
			SMTPPort:     os.Getenv("SMTP_PORT"),
			SMTPPassword: os.Getenv("SMTP_PASSWORD"),
			FromEmail:    os.Getenv("FROM_EMAIL"),
			//BaseURL:      os.Getenv
		}
	
		// Validate required configuration
		if emailConfig.SMTPHost == "" ||
			emailConfig.SMTPPort == "" ||
			emailConfig.SMTPPassword == "" ||
			emailConfig.FromEmail == "" {
			//emailConfig.BaseURL == "" 
			return fmt.Errorf("missing required email configuration. Please check your .env file")
		}
	
		log.Println("[INFO] Email configuration initialized successfully")
		return nil
	}

// Generate random token
func generateToken() (string, error) {
    fmt.Println("[DEBUG] Generating activation token...")
    b := make([]byte, 32)
    _, err := rand.Read(b)
    if err != nil {
        fmt.Println("[ERROR] Failed to generate token:", err)
        return "", err
    }
    token := base64.URLEncoding.EncodeToString(b)
    fmt.Println("[DEBUG] Generated token:", token)
    return token, nil
}

// Send activation email
func sendActivationEmail(user *models.User, token string) error {
    fmt.Println("[DEBUG] Preparing to send activation email to:", user.Email)

    auth := smtp.PlainAuth("", emailConfig.FromEmail, emailConfig.SMTPPassword, emailConfig.SMTPHost)
    
    activationLink := fmt.Sprintf("http://localhost:8080/activate-account?token=%s", token)
    message := fmt.Sprintf(`From: %s
To: %s
Subject: Account Activation - %s Organization

Hello %s,

You have been added to %s organization. Please click the link below to activate your account:

%s

This link will expire in 24 hours.

Best regards,
Your Application Team
`, emailConfig.FromEmail, user.Email, user.Organization, user.FullName, user.Organization, activationLink)

    fmt.Println("[DEBUG] Sending email to:", user.Email)
    err := smtp.SendMail(
        emailConfig.SMTPHost+":"+emailConfig.SMTPPort,
        auth,
        emailConfig.FromEmail,
        []string{user.Email},
        []byte(message),
    )

    if err != nil {
        fmt.Println("[ERROR] Failed to send email:", err)
        return err
    }

    fmt.Println("[DEBUG] Activation email sent successfully to:", user.Email)
    return nil
}

//Handler for sending activation email
func HandleSendActivationEmail(c echo.Context) error {
    userID := c.Param("user_id")
    fmt.Println("[DEBUG] Handling activation email request for User ID:", userID)

    var user models.User
    if err := db.GetDB().First(&user, userID).Error; err != nil {
        fmt.Println("[ERROR] User not found:", err)
        return c.JSON(404, map[string]string{"error": "User not found"})
    }
	log.Printf("received user_id")

    // Generate activation token
    token, err := generateToken()
    if err != nil {
        return c.JSON(500, map[string]string{"error": "Failed to generate token"})
    }

    // Save activation token
    activationToken := models.ActivationToken{
        UserID:    user.ID,
        Token:     token,
        ExpiresAt: time.Now().Add(24 * time.Hour),
        Used:      false,
    }

    fmt.Println("[DEBUG] Saving activation token for User ID:", user.ID)
    if err := db.GetDB().Create(&activationToken).Error; err != nil {
        fmt.Println("[ERROR] Failed to save activation token:", err)
        return c.JSON(500, map[string]string{"error": "Failed to save token"})
    }

    // Send email
    if err := sendActivationEmail(&user, token); err != nil {
        return c.JSON(500, map[string]string{"error": "Failed to send email"})
    }

    fmt.Println("[DEBUG] Activation email process completed for User ID:", user.ID)
    return c.JSON(200, map[string]string{"message": "Activation email sent successfully"})
}

// Handler for activating account
func ActivateAccount(c echo.Context) error {
    token := c.QueryParam("token")
    fmt.Println("[DEBUG] Activating account with token:", token)

    var activationToken models.ActivationToken
    if err := db.GetDB().Where("token = ? AND used = ? AND expires_at > ?", token, false, time.Now()).First(&activationToken).Error; err != nil {
        fmt.Println("[ERROR] Invalid or expired token:", err)
        return c.JSON(400, map[string]string{"error": "Invalid or expired token"})
    }

    // Get user
    var user models.User
    if err := db.GetDB().First(&user, activationToken.UserID).Error; err != nil {
        fmt.Println("[ERROR] User not found for activation:", err)
        return c.JSON(404, map[string]string{"error": "User not found"})
    }

    // Update user status
    fmt.Println("[DEBUG] Activating user account:", user.ID)
    user.IsActive = true
    if err := db.GetDB().Save(&user).Error; err != nil {
        fmt.Println("[ERROR] Failed to activate user:", err)
        return c.JSON(500, map[string]string{"error": "Failed to activate account"})
    }

    // Mark token as used
    fmt.Println("[DEBUG] Marking token as used for User ID:", user.ID)
    activationToken.Used = true
    if err := db.GetDB().Save(&activationToken).Error; err != nil {
        fmt.Println("[ERROR] Failed to update token status:", err)
        return c.JSON(500, map[string]string{"error": "Failed to update token status"})
    }

    fmt.Println("[DEBUG] Account activated successfully for User ID:", user.ID)
    return c.JSON(200, map[string]string{"message": "Account activated successfully"})
}
