package controllers

import (
    "net/http"
    "time"
    "stock/db"      
    "stock/models"  
    "stock/utils"
    "github.com/labstack/echo/v4"
)


func ForgotPassword(c echo.Context) error {
    email := c.FormValue("email")
    user := models.User{}
    
    // Check if user exists with this email
    if err := db.GetDB().Where("email = ?", email).First(&user).Error; err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"message": "User not found"})
    }

    // Generate a reset token
    resetToken := utils.GenerateResetToken()

    // Save token with an expiration time (e.g., 1 hour)
    user.ResetToken = resetToken
    user.ResetTokenExpiry = time.Now().Add(1 * time.Hour)
    if err := db.GetDB().Save(&user).Error; err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to save reset token"})
    }

    // Send password reset email
    resetLink := "http://185.192.96.72:6502/reset-password?token=" + resetToken
    utils.SendPasswordResetEmail(user.Email, resetLink)

    return c.JSON(http.StatusOK, map[string]string{"message": "Password reset link has been sent to your email"})
}

func ResetPassword(c echo.Context) error {
    token := c.QueryParam("token")
    newPassword := c.FormValue("new_password")

    // Verify the token
    user := models.User{}
    if err := db.GetDB().Where("reset_token = ? AND reset_token_expiry > ?", token, time.Now()).First(&user).Error; err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid or expired token"})
    }

    // Update the password
    hashedPassword, err := utils.HashPassword(newPassword)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to hash password"})
    }

    user.Password = hashedPassword
    user.ResetToken = "" // Clear the token after use
    user.ResetTokenExpiry = time.Time{}
    db.GetDB().Save(&user)

    return c.JSON(http.StatusOK, map[string]string{"message": "Password successfully reset"})
}
