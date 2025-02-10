package controllers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"stock/db"
	"stock/models"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

// Request structs for JSON binding
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

type ResetPasswordRequest struct {
	NewPassword string `json:"new_password"`
}

// EmailConfig struct for email settings
type EmailConfigs struct {
	SMTPHost     string
	SMTPPort     string
	SMTPPassword string
	FromEmail    string
}

var emailConfigs EmailConfig

func InitEmailConfigs() error {
	log.Println("[INFO] Initializing email configuration...")
	
	emailConfig = EmailConfig{
		SMTPHost:     os.Getenv("SMTP_HOST"),
		SMTPPort:     os.Getenv("SMTP_PORT"),
		SMTPPassword: os.Getenv("SMTP_PASSWORD"),
		FromEmail:    os.Getenv("FROM_EMAIL"),
	}

	// Validate configuration
	if emailConfig.SMTPHost == "" || emailConfig.SMTPPort == "" || 
	   emailConfig.SMTPPassword == "" || emailConfig.FromEmail == "" {
		log.Println("[ERROR] Missing required email configuration")
		return fmt.Errorf("missing required email configuration. Please check your .env file")
	}

	log.Printf("[DEBUG] SMTP Configuration loaded:")
	log.Printf("[DEBUG] - Host: %s", emailConfig.SMTPHost)
	log.Printf("[DEBUG] - Port: %s", emailConfig.SMTPPort)
	log.Printf("[DEBUG] - From Email: %s", emailConfig.FromEmail)
	log.Println("[INFO] Email configuration initialized successfully")
	
	return nil
}

func generateResetToken() (string, error) {
	log.Println("[DEBUG] Generating reset token...")
	
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		log.Printf("[ERROR] Failed to generate token: %v", err)
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	
	token := base64.URLEncoding.EncodeToString(b)
	log.Printf("[DEBUG] Token generated successfully: %s", token[:10]) 
	return token, nil
}

func sendResetPasswordEmail(user *models.User, token string) error {
	if user.Email == "" {
		log.Println("[ERROR] Cannot send email: email address is empty")
		return fmt.Errorf("invalid email address: email cannot be empty")
	}

	log.Printf("[INFO] Preparing to send password reset email to: %s", user.Email)

	auth := smtp.PlainAuth(
		"",
		emailConfig.FromEmail,
		emailConfig.SMTPPassword,
		emailConfig.SMTPHost,
	)

	resetLink := fmt.Sprintf("http://%s/forgot-password?token=%s", os.Getenv("BASE_URL"), token)
	log.Printf("[DEBUG] Generated reset link: %s", resetLink)

	var message strings.Builder
	
	headers := map[string]string{
		"From":         emailConfig.FromEmail,
		"To":          user.Email,
		"Subject":     "Password Reset Request",
		"MIME-Version": "1.0",
		"Content-Type": "text/html; charset=\"UTF-8\"",
	}

	for key, value := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}
	
	message.WriteString("\r\n")

	emailBody := fmt.Sprintf(`
<html>
<body>
    <p>Hello %s,</p>

    <p>We received a request to reset your password. Please click the link below to set a new password:</p>

    <p><a href="%s">Reset Your Password</a></p>

    <p>This link will expire in 24 hours. If you didn't request this reset, please ignore this email.</p>

    <p>Best regards,<br>
    Your Application Team</p>
</body>
</html>`, user.FullName, resetLink)

	message.WriteString(emailBody)

	smtpAddr := fmt.Sprintf("%s:%s", emailConfig.SMTPHost, emailConfig.SMTPPort)
	log.Printf("[DEBUG] Connecting to SMTP server at: %s", smtpAddr)

	err := smtp.SendMail(
		smtpAddr,
		auth,
		emailConfig.FromEmail,
		[]string{user.Email},
		[]byte(message.String()),
	)

	if err != nil {
		log.Printf("[ERROR] Failed to send email: %v", err)
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("[INFO] Password reset email sent successfully to: %s", user.Email)
	return nil
}

func ForgotPassword(c echo.Context) error {
	log.Println("[INFO] Processing forgot password request...")

	// Bind JSON request
	req := new(ForgotPasswordRequest)
	if err := c.Bind(req); err != nil {
		log.Printf("[ERROR] Failed to bind request: %v", err)
		return c.JSON(400, map[string]string{"error": "Invalid request format"})
	}

	email := strings.TrimSpace(req.Email)
	if email == "" {
		log.Println("[ERROR] Email is empty in request")
		return c.JSON(400, map[string]string{"error": "Email is required"})
	}

	log.Printf("[DEBUG] Processing forgot password request for email: %s", email)

	var user models.User
	if err := db.GetDB().Where("email = ?", email).First(&user).Error; err != nil {
		log.Printf("[DEBUG] User not found for email: %s", email)
		// Don't reveal if email exists for security
		return c.JSON(200, map[string]string{"message": "If your email is registered, you will receive reset instructions"})
	}

	token, err := generateResetToken()
	if err != nil {
		log.Printf("[ERROR] Token generation failed: %v", err)
		return c.JSON(500, map[string]string{"error": "Internal server error"})
	}

	resetToken := models.ResetToken{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		Used:      false,
	}

	log.Printf("[DEBUG] Saving reset token for user ID: %d", user.ID)
	if err := db.GetDB().Create(&resetToken).Error; err != nil {
		log.Printf("[ERROR] Failed to save reset token: %v", err)
		return c.JSON(500, map[string]string{"error": "Internal server error"})
	}

	if err := sendResetPasswordEmail(&user, token); err != nil {
		log.Printf("[ERROR] Failed to send reset email: %v", err)
		return c.JSON(500, map[string]string{"error": "Failed to send reset email"})
	}
	log.Printf("[INFO] Redirecting user to reset password page for email: %s", email)
	// return c.Redirect(302, "/reset-password")

	return c.JSON(200, map[string]string{"message": "link sent to your email"})
}

func ResetPassword(c echo.Context) error {
	log.Println("[INFO] Processing password reset request...")
	
	token := strings.TrimSpace(c.QueryParam("token"))
	if token == "" {
		log.Println("[ERROR] Reset token is missing")
		return c.JSON(400, map[string]string{"error": "Reset token is required"})
	}

	// Bind JSON request
	req := new(ResetPasswordRequest)
	if err := c.Bind(req); err != nil {
		log.Printf("[ERROR] Failed to bind request: %v", err)
		return c.JSON(400, map[string]string{"error": "Invalid request format"})
	}

	newPassword := strings.TrimSpace(req.NewPassword)
	if len(newPassword) < 8 {
		log.Println("[ERROR] Password is too short")
		return c.JSON(400, map[string]string{"error": "Password must be at least 8 characters long"})
	}

	log.Printf("[DEBUG] Looking up reset token: %s", token[:10]) // Log only first 10 chars for security
	var resetToken models.ResetToken
	if err := db.GetDB().Where("token = ? AND used = ? AND expires_at > ?", 
		token, false, time.Now()).First(&resetToken).Error; err != nil {
		log.Printf("[ERROR] Invalid or expired reset token: %v", err)
		return c.JSON(400, map[string]string{"error": "Invalid or expired reset token"})
	}

	log.Printf("[DEBUG] Found valid reset token for user ID: %d", resetToken.UserID)
	var user models.User
	if err := db.GetDB().First(&user, resetToken.UserID).Error; err != nil {
		log.Printf("[ERROR] Failed to find user: %v", err)
		return c.JSON(500, map[string]string{"error": "Internal server error"})
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("[ERROR] Failed to hash password: %v", err)
		return c.JSON(500, map[string]string{"error": "Failed to process new password"})
	}

	log.Printf("[DEBUG] Updating password for user ID: %d", user.ID)
	user.Password = string(hashedPassword)
	if err := db.GetDB().Save(&user).Error; err != nil {
		log.Printf("[ERROR] Failed to save new password: %v", err)
		return c.JSON(500, map[string]string{"error": "Failed to update password"})
	}

	log.Printf("[DEBUG] Marking reset token as used for user ID: %d", user.ID)
	resetToken.Used = true
	if err := db.GetDB().Save(&resetToken).Error; err != nil {
		log.Printf("[ERROR] Failed to mark token as used: %v", err)
	}

	log.Printf("[INFO] Password reset completed successfully for user ID: %d", user.ID)
	return c.JSON(200, map[string]string{"message": "Password reset successfully"})
}