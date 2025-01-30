package controllers

import (
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "net/http"
    "os"
    "time"
    
    "github.com/labstack/echo/v4"
    "gopkg.in/gomail.v2"
    "gorm.io/gorm"
)

type AuthController struct {
    DB *gorm.DB
}

type User struct {
    gorm.Model
    Email                string
    Password             string
    ResetPasswordToken   string
    ResetPasswordExpires time.Time
}

type ForgotPasswordRequest struct {
    Email string `json:"email" validate:"required,email"`
}

type ResetPasswordRequest struct {
    Token    string `json:"token" validate:"required"`
    Password string `json:"password" validate:"required,min=8"`
}

func NewAuthController(db *gorm.DB) *AuthController {
    return &AuthController{DB: db}
}

func generateResetToken() (string, error) {
    bytes := make([]byte, 32)
    if _, err := rand.Read(bytes); err != nil {
        return "", err
    }
    return hex.EncodeToString(bytes), nil
}

func sendResetEmail(email, token string) error {
    m := gomail.NewMessage()
    
    // Get email configuration from environment variables
    senderEmail := os.Getenv("SMTP_FROM_EMAIL")
    smtpHost := os.Getenv("SMTP_HOST")
    smtpPort := os.Getenv("SMTP_PORT")
    smtpUser := os.Getenv("SMTP_USER")
    smtpPass := os.Getenv("SMTP_PASSWORD")
    appDomain := os.Getenv("APP_DOMAIN")
    
    m.SetHeader("From", senderEmail)
    m.SetHeader("To", email)
    m.SetHeader("Subject", "Password Reset Request")
    
    resetLink := fmt.Sprintf("%s/reset-password?token=%s", appDomain, token)
    body := fmt.Sprintf("Click the following link to reset your password: %s\nThis link will expire in 1 hour.", resetLink)
    
    m.SetBody("text/plain", body)
    
    // Convert SMTP port to integer
    var port int
    fmt.Sscanf(smtpPort, "%d", &port)
    
    d := gomail.NewDialer(smtpHost, port, smtpUser, smtpPass)
    
    return d.DialAndSend(m)
}

func (ac *AuthController) HandleForgotPassword(c echo.Context) error {
    var req ForgotPasswordRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
    }
    
    // Find user by email
    var user User
    result := ac.DB.Where("email = ?", req.Email).First(&user)
    if result.Error != nil {
        return c.JSON(http.StatusNotFound, map[string]string{"error": "User not found"})
    }
    
    // Generate reset token
    token, err := generateResetToken()
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate reset token"})
    }
    
    // Update user with reset token and expiration
    user.ResetPasswordToken = token
    user.ResetPasswordExpires = time.Now().Add(1 * time.Hour)
    if err := ac.DB.Save(&user).Error; err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to save reset token"})
    }
    
    // Send reset email
    if err := sendResetEmail(user.Email, token); err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to send reset email"})
    }
    
    return c.JSON(http.StatusOK, map[string]string{"message": "Password reset email sent"})
}

func (ac *AuthController) HandleResetPassword(c echo.Context) error {
    var req ResetPasswordRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
    }
    
    // Find user by reset token
    var user User
    result := ac.DB.Where("reset_password_token = ?", req.Token).First(&user)
    if result.Error != nil {
        return c.JSON(http.StatusNotFound, map[string]string{"error": "Invalid reset token"})
    }
    
    // Check if token is expired
    if time.Now().After(user.ResetPasswordExpires) {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Reset token has expired"})
    }
    
    // Update password (in practice, hash the password before saving)
    user.Password = req.Password
    user.ResetPasswordToken = ""
    user.ResetPasswordExpires = time.Time{}
    
    if err := ac.DB.Save(&user).Error; err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to update password"})
    }
    
    return c.JSON(http.StatusOK, map[string]string{"message": "Password successfully reset"})
}
