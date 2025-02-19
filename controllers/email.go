package controllers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"gopkg.in/gomail.v2"
	"strconv"
	"errors"
	//"net/http"
	"net/smtp"
	"os"
	"stock/db"
	"stock/models"
	//"strings"
	"time"

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
	}

	// Debugging logs
	log.Printf("[DEBUG] SMTP Host: %s", emailConfig.SMTPHost)
	log.Printf("[DEBUG] SMTP Port: %s", emailConfig.SMTPPort)
	log.Printf("[DEBUG] SMTP From Email: %s", emailConfig.FromEmail)

	if emailConfig.SMTPHost == "" || emailConfig.SMTPPort == "" || emailConfig.SMTPPassword == "" || emailConfig.FromEmail == "" {
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

func sendActivationEmail(user *models.User, token, originalPassword string) error {
	fmt.Println("[DEBUG] Preparing to send activation email to:", user.Email)

	// Configure auth with the correct host
	auth := smtp.PlainAuth(
		"",
		emailConfig.FromEmail,
		emailConfig.SMTPPassword,
		emailConfig.SMTPHost,
	)

	// activationLink := fmt.Sprintf("http://%s/activate-account?token=%s", os.Getenv("BASE_URL"), token)
	activationLink := fmt.Sprintf("http://%s", os.Getenv("BASE_URL"))


	to := []string{user.Email}
	subject := "Account Activation - " + user.Organization

	// Message content with login credentials
	message := fmt.Sprintf(`From: %s
To: %s
Subject: %s
MIME-version: 1.0
Content-Type: text/html; charset="UTF-8"

<html>
<body>
<p>Hello %s,</p>

<p>You have been added to %s organization. Below are your login credentials:</p>

<p><strong>Email:</strong> %s</p>
<p><strong>Password:</strong> %s</p>

<p>You can log in using the following link:</p>

<p><a href="%s">Click here to log in</a></p>

<p>For security, please change your password after logging in.</p>

<p>Best regards,<br>
Your Admin</p>
</body>
</html>
`, emailConfig.FromEmail, user.Email, subject, user.FullName, user.Organization, user.Email, originalPassword, activationLink)

	// Connect to SMTP server with full address including port
	smtpAddr := fmt.Sprintf("%s:%s", emailConfig.SMTPHost, emailConfig.SMTPPort)
	fmt.Printf("[DEBUG] Connecting to SMTP server at: %s\n", smtpAddr)

	err := smtp.SendMail(
		smtpAddr,
		auth,
		emailConfig.FromEmail,
		to,
		[]byte(message),
	)

	if err != nil {
		fmt.Printf("[ERROR] Failed to send email: %v\n", err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	fmt.Println("[DEBUG] Activation email sent successfully to:", user.Email)
	return nil
}

// Handler for sending activation email
func HandleSendActivationEmail(c echo.Context) error {
	userID := c.Param("user_id")
	fmt.Println("[DEBUG] Handling activation email request for User ID:", userID)

	var user models.User
	if err := db.GetDB().First(&user, userID).Error; err != nil {
		fmt.Println("[ERROR] User not found:", err)
		return c.JSON(404, map[string]string{"error": "User not found"})
	}
	log.Printf("received user_id")

	token, err := generateToken()
	if err != nil {
		return c.JSON(500, map[string]string{"error": "Failed to generate token"})
	}

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

	originalPassword := user.Password

	if err := sendActivationEmail(&user, token, originalPassword); err != nil {
		return c.JSON(500, map[string]string{"error": "Failed to send email"})
	}

	fmt.Println("[DEBUG] Activation email process completed for User ID:", user.ID)
	return c.JSON(200, map[string]string{"message": "Activation email sent successfully"})
}

func sendSignupNotification(username, email, phone, Organization string) error {	
    // Load SMTP configuration from .env
    smtpHost := os.Getenv("SMTP_HOST")
    smtpPort := os.Getenv("SMTP_PORT")
    fromEmail := os.Getenv("FROM_EMAIL")
    smtpPass := os.Getenv("SMTP_PASSWORD")

    // Log the configuration (matching your debug logs)
    log.Printf("[DEBUG] SMTP Host: %s", smtpHost)
    log.Printf("[DEBUG] SMTP Port: %s", smtpPort)
    log.Printf("[DEBUG] From Email: %s", fromEmail)

    // Validate SMTP configuration
    if smtpHost == "" || smtpPort == "" || fromEmail == "" || smtpPass == "" {
        log.Println("[ERROR] SMTP configuration missing in .env file")
        return errors.New("SMTP configuration missing")
    }

    port, err := strconv.Atoi(smtpPort)
    if err != nil {
        log.Printf("[ERROR] Invalid SMTP port: %v", err)
        return err
    }

    // Email details
    to := "support@infinitytechafrica.com"
    subject := "New STOCK Client Registeration."

    body := fmt.Sprintf("New STOCK Client Registered with below details:\n \nUsername %s\nEmail: %s\nPhone: %s\nOrganization: %s", username, email, phone, Organization)

    // Configure email
    m := gomail.NewMessage()
    m.SetHeader("From", fromEmail)
    m.SetHeader("To", to)
    m.SetHeader("Subject", subject)
    m.SetBody("text/plain", body)

    d := gomail.NewDialer(smtpHost, port, fromEmail, smtpPass)

    // Send email
    if err := d.DialAndSend(m); err != nil {
        log.Printf("[ERROR] Failed to send email: %v", err)
        return err
    }

    log.Println("[INFO] Signup notification email sent successfully")
    return nil
}