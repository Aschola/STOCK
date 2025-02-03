package controllers

import (
	"log"
	"net/http"
	"stock/utils" 
	"stock/db"    
	"github.com/labstack/echo/v4"
	//"gorm.io/gorm"
)

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=6"`
}

// ChangePassword allows users to change their password
func ChangePassword(c echo.Context) error {
	// Retrieve userID from the token context
	userID, ok := c.Get("userID").(uint)
	if !ok {
		log.Println("[ERROR] User ID missing in token context")
		return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized access")
	}
	log.Printf("[INFO] Change password request received for user ID: %d", userID)

	// Get the DB instance using your custom GetDB function
	db := db.GetDB()

	// Parse the request body
	var req ChangePasswordRequest
	if err := c.Bind(&req); err != nil {
		log.Printf("[ERROR] Failed to parse request body: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request format")
	}

	// Fetch the user record from the database
	var user struct {
		ID       uint
		Password string
	}
	if err := db.Table("users").Where("id = ?", userID).First(&user).Error; err != nil {
		log.Printf("[ERROR] User not found in database for user ID: %d, error: %v", userID, err)
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}
	log.Printf("[INFO] Retrieved user data for user ID: %d", userID)

	// Verify old password using `utils.CheckPasswordHash`
	if err := utils.CheckPasswordHash(req.OldPassword, user.Password); err != nil {
		log.Printf("[ERROR] Incorrect old password for user ID: %d, error: %v", userID, err)
		return echo.NewHTTPError(http.StatusUnauthorized, "Incorrect old password")
	}
	log.Printf("[INFO] Old password verified for user ID: %d", userID)

	// Prevent using the same password
	if req.OldPassword == req.NewPassword {
		log.Printf("[WARN] New password cannot be the same as the old password for user ID: %d", userID)
		return echo.NewHTTPError(http.StatusBadRequest, "New password cannot be the same as old password")
	}

	// Hash the new password using `utils.HashPassword`
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		log.Printf("[ERROR] Password hashing failed for user ID: %d, error: %v", userID, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not hash password")
	}
	log.Printf("[INFO] New password hashed successfully for user ID: %d", userID)

	// Update the password in the database
	if err := db.Table("users").Where("id = ?", userID).Update("password", hashedPassword).Error; err != nil {
		log.Printf("[ERROR] Failed to update password for user ID: %d, error: %v", userID, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not update password")
	}
	log.Printf("[INFO] Password updated successfully for user ID: %d", userID)

	return c.JSON(http.StatusOK, echo.Map{"message": "Password changed successfully"})
}
