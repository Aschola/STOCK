package controllers

import (
	"log"
	"net/http"
	"stock/db"
	"stock/models"
	"stock/utils"
	"time"

	"github.com/labstack/echo/v4"
	//"golang.org/x/crypto/bcrypt"
	"stock/validators"

)
func OrganizationAdminLogin(c echo.Context) error {
	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.Bind(&input); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	loginInput := validators.LoginInput{
		Username: input.Username,
		Password: input.Password,
	}
	if err := validators.ValidateLoginInput(loginInput); err != nil {
        log.Printf("Validation error: %v", err)
        return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
    }

	var user models.User
	if err := db.GetDB().Where("username = ?", input.Username).First(&user).Error; err != nil {
		log.Printf("Where error: %v", err)
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid username or password"})
	}

	if err := utils.CheckPasswordHash(input.Password, user.Password); err != nil {
		log.Printf("CheckPasswordHash error: %v", err)
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid username or password"})
	}

	token, err := utils.GenerateJWT(user.ID, user.RoleName) 
	if err != nil {
		log.Printf("GenerateJWT error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not generate token"})
	}

	log.Println("logged in successfully")
	return c.JSON(http.StatusOK, echo.Map{"token": token})
}

func OrganizationAdminLogout(c echo.Context) error {
	// Clear token or handle session invalidation if necessary
	return c.JSON(http.StatusOK, map[string]string{"message": "Logged out successfully"})
}

func OrganizationLogout(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"message": "logged out successfully"})
}

func OrganizationLogin(c echo.Context) error {
	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.Bind(&input); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	loginInput := validators.LoginInput{
		Username: input.Username,
		Password: input.Password,
	}
	if err := validators.ValidateLoginInput(loginInput); err != nil {
        log.Printf("Validation error: %v", err)
        return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
    }

	var user models.User
	if err := db.GetDB().Where("username = ?", input.Username).First(&user).Error; err != nil {
		log.Printf("Where error: %v", err)
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid username or password"})
	}

	if err := utils.CheckPasswordHash(input.Password, user.Password); err != nil {
		log.Printf("CheckPasswordHash error: %v", err)
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid username or password"})
	}

	token, err := utils.GenerateJWT(user.ID, user.RoleName) 
	if err != nil {
		log.Printf("GenerateJWT error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not generate token"})
	}

	log.Println("logged in successfully")
	return c.JSON(http.StatusOK, echo.Map{"token": token})
}
func OrganizationAdminAddUser(c echo.Context) error {
	log.Println("AdminAddUser - Entry")

	userID, ok := c.Get("userID").(int)
	if !ok {
		log.Println("AdminAddUser - Unauthorized: userID not found in context")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	roleName, ok := c.Get("roleName").(string) 
	if !ok {
		log.Println("AdminAddUser - Unauthorized: roleName not found in context")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	log.Printf("AdminAddUser - Received RoleName: %s, UserID: %d", roleName, userID)

	// Check if the roleName is "admin"
	if roleName != "admin" {
		log.Println("AdminAddUser - Permission denied: non-admin trying to add user")
		return c.JSON(http.StatusForbidden, echo.Map{"error": "Permission denied"})
	}

	var input models.User
	if err := c.Bind(&input); err != nil {
		log.Printf("AdminAddUser - Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	log.Printf("AdminAddUser - New user data: %+v", input)

	// Validate roleName for new user
	if input.RoleName != "shopkeeper" && input.RoleName != "auditor" && input.RoleName != "admin" {
		log.Println("AdminAddUser - Invalid role name provided")
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid role name. Allowed roles: shopkeeper, auditor, admin"})
	}
	input.CreatedBy = uint(userID)

	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		log.Printf("AdminAddUser - HashPassword error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not hash password"})
	}
	input.Password = hashedPassword

	if err := db.GetDB().Create(&input).Error; err != nil {
		log.Printf("AdminAddUser - Create error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	log.Println("AdminAddUser - User created successfully")
	log.Println("AdminAddUser - Exit")
	return c.JSON(http.StatusOK, echo.Map{"message": "User created successfully"})
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

// DeactivateOrganization deactivates an organization
func OrgAdminDeactivateOrganization(c echo.Context) error {
	userID, ok := c.Get("userID").(int)
	if !ok {
		log.Println("Failed to get userID from context")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	roleID, ok := c.Get("roleID").(int)
	if !ok {
		log.Println("Failed to get roleID from context")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	// Log the received userID and roleID
	log.Printf("Received UserID: %d, RoleID: %d", userID, roleID)

	// Check if the roleID is for a Super Admin (roleID = 1)
	if roleID != 6 {
		log.Println("Permission denied: non-super admin trying to deactivate organization")
		return c.JSON(http.StatusForbidden, echo.Map{"error": "Permission denied"})
	}

	orgID := c.QueryParam("organization_id")
	if orgID == "" {
		log.Println("Organization ID is required")
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Organization ID is required"})
	}

	var organization models.Organization
	if err := db.GetDB().First(&organization, orgID).Error; err != nil {
		log.Printf("Organization not found: %v", err)
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Organization not found"})
	}

	organization.IsActive = false
	if err := db.GetDB().Save(&organization).Error; err != nil {
		log.Printf("Error deactivating organization: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	log.Println("Organization deactivated successfully")
	return c.JSON(http.StatusOK, echo.Map{"message": "Organization deactivated successfully", "organization": organization})
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
