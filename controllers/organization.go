package controllers

import (
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"stock/db"
	"stock/models"
	"stock/utils"
	"time"
)

func OrganizationAdminLogin(c echo.Context) error {
	var user models.User
	if err := c.Bind(&user); err != nil {
		return err
	}

	var storedUser models.User
	if err := db.GetDB().Where("username = ?", user.Username).First(&storedUser).Error; err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Invalid credentials"})
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(user.Password)); err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"message": "Invalid credentials"})
	}

	token, err := utils.GenerateJWT(storedUser.ID, storedUser.RoleID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]string{"token": token})
}

func OrganizationAdminLogout(c echo.Context) error {
	// Clear token or handle session invalidation if necessary
	return c.JSON(http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

func OrganizationAdminAddUser(c echo.Context) error {
	log.Println("OrganizationAdminAddUser called")

	roleID, ok := c.Get("roleID").(int)
	if !ok {
		log.Println("Failed to get roleID from context")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	userID, ok := c.Get("userID").(int)
	if !ok {
		log.Println("Failed to get userID from context")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	log.Printf("Received RoleID: %d, UserID: %d", roleID, userID)

	if roleID != 6 {
		log.Println("Permission denied: non-organization admin trying to add user")
		return c.JSON(http.StatusForbidden, echo.Map{"error": "Permission denied"})
	}

	var newUser models.User
	if err := c.Bind(&newUser); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	log.Printf("New user data: %+v", newUser)

	if newUser.RoleID != 7 && newUser.RoleID != 8 {
		log.Println("Invalid role ID")
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid role ID"})
	}

	var existingUser models.User
	if err := db.GetDB().Where("username = ?", newUser.Username).First(&existingUser).Error; err == nil {
		log.Println("Username already exists")
		return c.JSON(http.StatusConflict, echo.Map{"error": "Username already exists"})
	}

	hashedPassword, err := utils.HashPassword(newUser.Password)
	if err != nil {
		log.Printf("HashPassword error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not hash password"})
	}
	newUser.Password = hashedPassword
	newUser.CreatedBy = uint(userID) // Convert userID to uint

	log.Printf("Saving new user to database")

	if err := db.GetDB().Create(&newUser).Error; err != nil {
		log.Printf("Create error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	log.Println("User added successfully")
	return c.JSON(http.StatusCreated, echo.Map{"message": "User added successfully", "user": newUser})
}

func OrganizationAdminEditUser(c echo.Context) error {
	log.Println("OrganizationAdminEditUser called")

	roleID, ok := c.Get("roleID").(int)
	if !ok {
		log.Println("Failed to get roleID from context")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	userID, ok := c.Get("userID").(int)
	if !ok {
		log.Println("Failed to get userID from context")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	log.Printf("Received RoleID: %d, UserID: %d", roleID, userID)

	if roleID != 6 {
		log.Println("Permission denied: non-organization admin trying to edit user")
		return c.JSON(http.StatusForbidden, echo.Map{"error": "Permission denied"})
	}

	userIDParam := c.Param("id")
	var user models.User
	if err := c.Bind(&user); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	log.Printf("Updating user data: %+v", user)

	if err := db.GetDB().Model(&models.User{}).Where("id = ? AND organization_id = ?", userIDParam, userID).Updates(user).Error; err != nil {
		log.Printf("Update error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	log.Println("User updated successfully")
	return c.JSON(http.StatusOK, echo.Map{"message": "User updated successfully"})
}

func OrganizationAdminGetUsers(c echo.Context) error {
	log.Println("OrganizationAdminGetUsers called")

	roleID, ok := c.Get("roleID").(int)
	if !ok {
		log.Println("Failed to get roleID from context")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	userID, ok := c.Get("userID").(int)
	if !ok {
		log.Println("Failed to get userID from context")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	log.Printf("Received RoleID: %d, UserID: %d", roleID, userID)

	if roleID != 6 {
		log.Println("Permission denied: non-organization admin trying to view users")
		return c.JSON(http.StatusForbidden, echo.Map{"error": "Permission denied"})
	}

	var users []models.User
	if err := db.GetDB().Where("organization_id = ?", userID).Find(&users).Error; err != nil {
		log.Printf("Find error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	log.Println("Users retrieved successfully")
	return c.JSON(http.StatusOK, users)
}

func OrganizationAdminGetUserByID(c echo.Context) error {
	log.Println("OrganizationAdminGetUserByID called")

	roleID, ok := c.Get("roleID").(int)
	if !ok {
		log.Println("Failed to get roleID from context")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	userID, ok := c.Get("userID").(int)
	if !ok {
		log.Println("Failed to get userID from context")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	log.Printf("Received RoleID: %d, UserID: %d", roleID, userID)

	if roleID != 6 {
		log.Println("Permission denied: non-organization admin trying to get user by ID")
		return c.JSON(http.StatusForbidden, echo.Map{"error": "Permission denied"})
	}

	userIDParam := c.Param("id")
	var user models.User
	if err := db.GetDB().Where("id = ? AND organization_id = ?", userIDParam, userID).First(&user).Error; err != nil {
		log.Printf("Find error: %v", err)
		return c.JSON(http.StatusNotFound, echo.Map{"error": "User not found"})
	}

	log.Println("User retrieved successfully")
	return c.JSON(http.StatusOK, user)
}

func OrganizationAdminSoftDeleteUser(c echo.Context) error {
	log.Println("OrganizationAdminSoftDeleteUser called")

	roleID, ok := c.Get("roleID").(int)
	if !ok {
		log.Println("Failed to get roleID from context")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	userID, ok := c.Get("userID").(int)
	if !ok {
		log.Println("Failed to get userID from context")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	log.Printf("Received RoleID: %d, UserID: %d", roleID, userID)

	if roleID != 6 {
		log.Println("Permission denied: non-organization admin trying to soft-delete user")
		return c.JSON(http.StatusForbidden, echo.Map{"error": "Permission denied"})
	}

	userIDParam := c.Param("id")
	if err := db.GetDB().Model(&models.User{}).Where("id = ? AND organization_id = ?", userIDParam, userID).Update("deleted_at", time.Now()).Error; err != nil {
		log.Printf("Update error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	log.Println("User soft-deleted successfully")
	return c.JSON(http.StatusOK, echo.Map{"message": "User soft-deleted successfully"})
}

func OrganizationAdminActivateDeactivateUser(c echo.Context) error {
	log.Println("OrganizationAdminActivateDeactivateUser called")

	roleID, ok := c.Get("roleID").(int)
	if !ok {
		log.Println("Failed to get roleID from context")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	userID, ok := c.Get("userID").(int)
	if !ok {
		log.Println("Failed to get userID from context")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	log.Printf("Received RoleID: %d, UserID: %d", roleID, userID)

	if roleID != 6 {
		log.Println("Permission denied: non-organization admin trying to activate/deactivate user")
		return c.JSON(http.StatusForbidden, echo.Map{"error": "Permission denied"})
	}

	userIDParam := c.Param("id")
	var user models.User
	if err := db.GetDB().Where("id = ? AND organization_id = ?", userIDParam, userID).First(&user).Error; err != nil {
		log.Printf("Find error: %v", err)
		return c.JSON(http.StatusNotFound, echo.Map{"error": "User not found"})
	}

	user.IsActive = !user.IsActive
	if err := db.GetDB().Save(&user).Error; err != nil {
		log.Printf("Save error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	log.Println("User activation/deactivation updated successfully")
	return c.JSON(http.StatusOK, echo.Map{"message": "User activation/deactivation updated successfully", "user": user})
}
