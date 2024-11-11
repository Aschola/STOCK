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
    users := models.Users{}
    
    if err := db.GetDB().Where("email = ?", email).First(&users).Error; err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"message": "User not found"})
    }

    resetToken := utils.GenerateResetToken()

    users.ResetToken = resetToken
    users.ResetTokenExpiry = time.Now().Add(1 * time.Hour)
    if err := db.GetDB().Save(&users).Error; err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to save reset token"})
    }

    // Send password reset email
    resetLink := "http://185.192.96.72:6502/reset-password?token=" + resetToken
    utils.SendPasswordResetEmail(users.Email, resetLink)

    return c.JSON(http.StatusOK, map[string]string{"message": "Password reset link has been sent to your email"})
}

func ResetPassword(c echo.Context) error {
    token := c.QueryParam("token")
    newPassword := c.FormValue("new_password")

    // Verify the token
    users := models.Users{}
    if err := db.GetDB().Where("reset_token = ? AND reset_token_expiry > ?", token, time.Now()).First(&users).Error; err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"message": "Invalid or expired token"})
    }

    // Update the password
    hashedPassword, err := utils.HashPassword(newPassword)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Failed to hash password"})
    }

    users.Password = hashedPassword
    users.ResetToken = "" 
    users.ResetTokenExpiry = time.Time{}
    db.GetDB().Save(&users)

    return c.JSON(http.StatusOK, map[string]string{"message": "Password successfully reset"})
}
