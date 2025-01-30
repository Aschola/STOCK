package controllers
import (
	"github.com/labstack/echo/v4"
      "log"
	  "net/http"
	  "stock/models"
	  "stock/db"
	  "golang.org/x/crypto/bcrypt"
//"github.com/golang-jwt/jwt"

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

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		log.Println("ChangePassword - Error: Old password is incorrect")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Old password is incorrect"})
	}

	// Hash the new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("ChangePassword - Error hashing password: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not update password"})
	}

	// Update the user's password
	if err := db.GetDB().Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		log.Printf("ChangePassword - Error updating password: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not update password"})
	}

	log.Println("ChangePassword - Password updated successfully")
	return c.JSON(http.StatusOK, echo.Map{"message": "Password changed successfully"})
}
