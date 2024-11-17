package controllers

import (
	"log"
	"net/http"
	"stock/db"
	"stock/models"
	"stock/utils"
	//"time"
	"strings"
	"strconv"
	"gorm.io/gorm"

	"github.com/labstack/echo/v4"
	//"golang.org/x/crypto/bcrypt"
	"stock/validators"

)
// func OrganizationAdminLogin(c echo.Context) error {
// 	var input struct {
// 		Username string `json:"username" binding:"required"`
// 		Password string `json:"password" binding:"required"`
// 	}

// 	if err := c.Bind(&input); err != nil {
// 		log.Printf("Bind error: %v", err)
// 		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
// 	}

// 	loginInput := validators.LoginInput{
// 		Username: input.Username,
// 		Password: input.Password,
// 	}
// 	if err := validators.ValidateLoginInput(loginInput); err != nil {
//         log.Printf("Validation error: %v", err)
//         return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
//     }

// 	var user models.User
// 	if err := db.GetDB().Where("username = ?", input.Username).First(&user).Error; err != nil {
// 		log.Printf("Where error: %v", err)
// 		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid username or password"})
// 	}

// 	if err := utils.CheckPasswordHash(input.Password, user.Password); err != nil {
// 		log.Printf("CheckPasswordHash error: %v", err)
// 		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid username or password"})
// 	}

// 	token, err := utils.GenerateJWT(user.ID, user.RoleName) 
// 	if err != nil {
// 		log.Printf("GenerateJWT error: %v", err)
// 		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not generate token"})
// 	}

// 	log.Println("logged in successfully")
// 	return c.JSON(http.StatusOK, echo.Map{"token": token})
// }

// func OrganizationAdminLogout(c echo.Context) error {
// 	// Clear token or handle session invalidation if necessary
// 	return c.JSON(http.StatusOK, map[string]string{"message": "Logged out successfully"})
// }

func AdminSignup(c echo.Context) error {
	var input models.User
	if err := c.Bind(&input); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	SignupInput := validators.SignupInput{
		Username: input.Username,
		Password: input.Password,
	}
	if err := validators.ValidateLoginInput(validators.LoginInput(SignupInput)); err != nil {
		log.Printf("Validation error: %v", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	// Automatically set role to "Admin"

	log.Printf("Received JSON: %+v", input)
	var user models.User
	if err := db.GetDB().Where("username = ?", input.Username).First(&user).Error; err == nil {
		return c.JSON(http.StatusConflict, echo.Map{"error": "Username already exists"})
	}

	if err := db.GetDB().Where("email = ?", input.Email).First(&user).Error; err == nil {
		return c.JSON(http.StatusConflict, echo.Map{"error": "Email already exists"})
	}

	// Hash the password
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		log.Printf("HashPassword error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not hash password"})
	}
	input.Password = hashedPassword

	// Create the organization and associate the admin
	organization := models.Organization{
		Name:  input.Organization,
		Email: input.Email,
	}
	if err := db.GetDB().Create(&organization).Error; err != nil {
		log.Printf("Organization creation error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not create organization"})
	}

	// Set organization ID for the user
	input.OrganizationID = organization.ID
	input.RoleName = "Admin"


	// Create the user with the assigned organization ID and admin role
	if err := db.GetDB().Create(&input).Error; err != nil {
		log.Printf("User creation error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(input.ID, input.RoleName)
	if err != nil {
		log.Printf("GenerateJWT error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not generate token"})
	}

	log.Println("Admin signed up successfully")
	return c.JSON(http.StatusOK, echo.Map{"token": token})
}

func AdminLogout(c echo.Context) error {
	log.Println("AdminLogout - Entry")
	log.Println("AdminLogout - Admin logged out successfully")
	log.Println("AdminLogout - Exit")
	return c.JSON(http.StatusOK, echo.Map{"message": "Successfully logged out"})
}

func AdminAddUser(c echo.Context) error {
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

// func OrganizationAdminGetUserByID(c echo.Context) error {
// 	log.Println("OrganizationAdminGetUserByID called")

// 	roleID, ok := c.Get("roleID").(int)
// 	if !ok {
// 		log.Println("Failed to get roleID from context")
// 		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
// 	}

// 	userID, ok := c.Get("userID").(int)
// 	if !ok {
// 		log.Println("Failed to get userID from context")
// 		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
// 	}

// 	log.Printf("Received RoleID: %d, UserID: %d", roleID, userID)

// 	if roleID != 6 {
// 		log.Println("Permission denied: non-organization admin trying to get user by ID")
// 		return c.JSON(http.StatusForbidden, echo.Map{"error": "Permission denied"})
// 	}

// 	userIDParam := c.Param("id")
// 	var user models.User
// 	if err := db.GetDB().Where("id = ? AND organization_id = ?", userIDParam, userID).First(&user).Error; err != nil {
// 		log.Printf("Find error: %v", err)
// 		return c.JSON(http.StatusNotFound, echo.Map{"error": "User not found"})
// 	}

// 	log.Println("User retrieved successfully")
// 	return c.JSON(http.StatusOK, user)
// }

func OrganizationAdminSoftDeleteUser(c echo.Context) error {
	id := c.Param("id")
	log.Printf("SoftDeleteOrganization called with ID: %s", id)

	roleName, ok := c.Get("roleName").(string)
	if !ok || roleName != "superadmin" {
		log.Println("Failed to get roleName from context or insufficient permissions")
		return c.JSON(http.StatusForbidden, echo.Map{"error": "Only superadmins can delete organizations"})
	}

	var organization models.Organization

	if err := db.GetDB().First(&organization, id).Error; err != nil {
		log.Printf("Error finding organization: %v", err)
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Organization not found"})
	}

	if roleName != "Organization" {
		log.Println("Unauthorized: Only organizations can be deleted")
		return c.JSON(http.StatusForbidden, echo.Map{"error": "Unauthorized: Only organizations can be deleted"})
	}

	// Soft-delete the organization by setting it as inactive
	organization.IsActive = false
	if err := db.GetDB().Save(&organization).Error; err != nil {
		log.Printf("Error saving organization: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	// Deactivate all users associated with the organization
	if err := db.GetDB().Model(&models.User{}).Where("organization_id = ?", id).Updates(map[string]interface{}{"is_active": false}).Error; err != nil {
		log.Printf("Error deactivating users: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Error deactivating users"})
	}

	log.Println("Organization and associated users soft deleted successfully")
	return c.JSON(http.StatusOK, echo.Map{"message": "Organization and associated users soft deleted successfully"})
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


func AdminLogin(c echo.Context) error {
	log.Println("AdminLogin - Entry")

	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.Bind(&input); err != nil {
		log.Printf("AdminLogin - Bind error: %v", err)
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

	log.Printf("AdminLogin - Received input: %+v", input)

	var user models.User
	if err := db.GetDB().Where("username = ?", input.Username).First(&user).Error; err != nil {
		log.Printf("AdminLogin - Where error: %v", err)
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid username or password"})
	}

	if err := utils.CheckPasswordHash(input.Password, user.Password); err != nil {
		log.Printf("AdminLogin - CheckPasswordHash error: %v", err)
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid email or password"})
	}

	token, err := utils.GenerateJWT(user.ID, user.RoleName)
	if err != nil {
		log.Printf("AdminLogin - GenerateJWT error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not generate token"})
	}

	log.Println("AdminLogin - Admin logged in successfully")
	log.Println("AdminLogin - Exit")
	return c.JSON(http.StatusOK, echo.Map{"token": token})
}

func GetUserByID(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	log.Printf("GetUserByID - Entry with ID: %d", id)

	var user models.User
	if err := db.GetDB().First(&user, id).Error; err != nil {
		log.Printf("GetUserByID - First error: %v", err)
		return c.JSON(http.StatusNotFound, echo.Map{"error": "User not found"})
	}

	log.Printf("GetUserByID - User found: %+v", user)
	log.Println("GetUserByID - Exit")
	return c.JSON(http.StatusOK, user)
}

func EditUser(c echo.Context) error {
	id := c.Param("id")
	log.Printf("EditUser - Entry with ID: %s", id)

	var user models.User
	if err := db.GetDB().First(&user, id).Error; err != nil {
		log.Printf("EditUser - First error: %v", err)
		return c.JSON(http.StatusNotFound, echo.Map{"error": "User not found"})
	}

	log.Printf("EditUser - Current user details: %+v", user)

	if err := c.Bind(&user); err != nil {
		log.Printf("EditUser - Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	if err := db.GetDB().Save(&user).Error; err != nil {
		log.Printf("EditUser - Save error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	log.Println("EditUser - User updated successfully")
	log.Println("EditUser - Exit")
	return c.JSON(http.StatusOK, user)
}

func AdminViewAllUsers(c echo.Context) error {
	log.Println("AdminViewAllUsers - Entry")

	// Retrieve roleName from context set by middleware
	roleName, ok := c.Get("roleName").(string)
	if !ok {
		log.Println("AdminViewAllUsers - Failed to get roleName from context")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	log.Printf("AdminViewAllUsers - Received RoleName: %s", roleName)

	if roleName != "Admin" {
		log.Println("AdminViewAllUsers - Permission denied: non-admin trying to view users")
		return c.JSON(http.StatusForbidden, echo.Map{"error": "Permission denied"})
	}

	var users []models.User
	if err := db.GetDB().Find(&users).Error; err != nil {
		log.Printf("AdminViewAllUsers - Find error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not retrieve users"})
	}

	log.Printf("AdminViewAllUsers - Retrieved users: %+v", users)
	log.Println("AdminViewAllUsers - Exit")
	return c.JSON(http.StatusOK, users)
}

func SoftDeleteUser(c echo.Context) error {
	id := c.Param("id")
	log.Printf("SoftDeleteUser called with ID: %s", id)

	roleName, ok := c.Get("roleName").(string)
	if !ok || (roleName != "Admin" && roleName != "OrganizationAdmin") {
		log.Println("Failed to get roleName from context or insufficient permissions")
		return c.JSON(http.StatusForbidden, echo.Map{"error": "Only admins or organization admins can delete users"})
	}

	var user models.User
	if err := db.GetDB().First(&user, id).Error; err != nil {
		log.Printf("Error finding user: %v", err)
		return c.JSON(http.StatusNotFound, echo.Map{"error": "User not found"})
	}
	user.IsActive = false
	if !strings.HasPrefix(user.Username, "deleted_") { 
		user.Username = "deleted_" + user.Username
	}


	user.IsActive = false
	if err := db.GetDB().Save(&user).Error; err != nil {
		log.Printf("Error saving user: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	log.Println("User soft deleted (deactivated) successfully")
	return c.JSON(http.StatusOK, echo.Map{"message": "User soft deleted successfully", "user": user})
}

func ReactivateUser(c echo.Context) error {
	id := strings.TrimSpace(c.Param("id"))
	log.Printf("ReactivateUser called with ID: '%s'", id)

	roleName, ok := c.Get("roleName").(string)
	if !ok || (roleName != "Admin" && roleName != "OrganizationAdmin") {
		log.Println("Failed to get roleName from context or insufficient permissions")
		return c.JSON(http.StatusForbidden, echo.Map{"error": "Only admins or organization admins can reactivate users"})
	}

	var user models.User

	if err := db.GetDB().Unscoped().Where("id = ?", id).First(&user).Error; err != nil {
		log.Printf("Error finding user: %v", err)
		return c.JSON(http.StatusNotFound, echo.Map{"error": "User not found"})
	}

	if user.IsActive {
		log.Println("User is already active")
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "User is already active"})
	}

	if user.DeletedAt.Valid {
		user.DeletedAt = gorm.DeletedAt{}
	}

	if !user.IsActive {
		user.IsActive = true
	}

	if err := db.GetDB().Save(&user).Error; err != nil {
		log.Printf("Error saving user: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	log.Println("User reactivated successfully")
	return c.JSON(http.StatusOK, echo.Map{"message": "User reactivated successfully", "user": user})
}

func DeactivateUser(c echo.Context) error {
	userID := c.Param("id")
	var user models.User

	if err := db.GetDB().First(&user, userID).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "User not found"})
	}

	user.IsActive = false
	if err := db.GetDB().Save(&user).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Error saving user"})
	}

	return c.JSON(http.StatusOK, user)
}

func GetActiveUsers(c echo.Context) error {
	var users []models.User
	if err := db.GetDB().Where("is_active = ?", true).Find(&users).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Error retrieving users"})
	}
	return c.JSON(http.StatusOK, users)
}

func GetInactiveUsers(c echo.Context) error {
	var users []models.User
	if err := db.GetDB().Where("is_active = ?", false).Find(&users).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Error retrieving users"})
	}
	return c.JSON(http.StatusOK, users)
}

