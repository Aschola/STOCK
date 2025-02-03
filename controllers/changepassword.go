package controllers

import (
    "github.com/labstack/echo/v4"
    "log"
    "net/http"
    "stock/models"
    "stock/db"
	"stock/utils"
    "golang.org/x/crypto/bcrypt"
)



func ChangePassword(c echo.Context) error {
    log.Println("ChangePassword - Entry")

    // Retrieve user ID from the context
    userID, ok := c.Get("userID").(uint)
    if !ok {
        log.Println("ChangePassword - Unauthorized: userID not found in context")
        return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
    }

    // Parse the request body
    var req models.ChangePasswordRequest
    if err := c.Bind(&req); err != nil {
        log.Printf("ChangePassword - Bind error: %v", err)
        return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid input"})
    }

    // Fetch the user from the database
    var user models.User
    if err := db.GetDB().Where("id = ?", userID).First(&user).Error; err != nil {
        log.Printf("ChangePassword - User not found: %v", err)
        return c.JSON(http.StatusNotFound, echo.Map{"error": "User not found"})
    }
    // Store the old password hash for logging
    oldPasswordHash := user.Password

    log.Printf("Stored Hashed Password: %s", user.Password)
log.Printf("Entered Old Password: %s", req.OldPassword)

if err := utils.CheckPasswordHash(req.OldPassword, user.Password); err != nil {
    log.Println("ChangePassword - Error: Old password is incorrect")
    return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Old password is incorrect"})
}

// Hash new password
hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
if err != nil {
    log.Printf("ChangePassword - Error hashing password: %v", err)
    return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not update password"})
}

// Save the new password
user.Password = string(hashedPassword)
if err := db.GetDB().Save(&user).Error; err != nil {
    log.Printf("ChangePassword - Error updating password: %v", err)
    return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not update password"})
}

// Verify the update
var updatedUser models.User
if err := db.GetDB().Where("id = ?", userID).First(&updatedUser).Error; err != nil {
    log.Printf("ChangePassword - Error fetching updated user: %v", err)
    return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not verify password update"})
}

log.Printf("Updated Hashed Password: %s", updatedUser.Password)

if oldPasswordHash == updatedUser.Password {
    log.Println("ChangePassword - Error: Password not actually updated")
    return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Password update failed"})
}

log.Println("ChangePassword - Password updated successfully")
return c.JSON(http.StatusOK, echo.Map{"message": "Password changed successfully"})
}