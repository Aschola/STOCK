package controllers

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"

	"stock/db"
	"stock/models"
	"stock/utils"
	"stock/validators"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func SuperAdminSignup(c echo.Context) error {
	var input models.User
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

	log.Printf("Received JSON: %+v", input)
	log.Printf("Received RoleID: %d", input.RoleID)

	if input.RoleID != 1 {
		log.Println("Unauthorized signup attempt detected")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Only super admin can sign up"})
	}

	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		log.Printf("HashPassword error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not hash password"})
	}
	input.Password = hashedPassword

	if err := db.GetDB().Create(&input).Error; err != nil {
		log.Printf("Create error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	token, err := utils.GenerateJWT(input.ID, input.RoleID)
	if err != nil {
		log.Printf("GenerateJWT error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not generate token"})
	}

	log.Println("Super admin signed up successfully")
	return c.JSON(http.StatusOK, echo.Map{"token": token})
}

func AddAdmin(c echo.Context) error {
	log.Println("AddAdmin called")

	// Retrieve roleID and userID from context set by middleware
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

	// Check if the roleID is 1 (Super Admin)
	if roleID != 1 {
		log.Println("Permission denied: non-super admin trying to add admin")
		return c.JSON(http.StatusForbidden, echo.Map{"error": "Permission denied"})
	}

	var newAdmin models.User
	if err := c.Bind(&newAdmin); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	log.Printf("New admin data: %+v", newAdmin)

	loginInput := validators.LoginInput{
		Username: newAdmin.Email,
		Password: newAdmin.Password,
	}
	if err := validators.ValidateLoginInput(loginInput); err != nil {
		log.Printf("Validation error: %v", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	log.Printf("New admin data: %+v", newAdmin)

	newAdmin.RoleID = 2 // Set roleID for new admin

	hashedPassword, err := utils.HashPassword(newAdmin.Password)
	if err != nil {
		log.Printf("HashPassword error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not hash password"})
	}
	newAdmin.Password = hashedPassword

	log.Printf("Saving new admin to database")

	if err := db.GetDB().Create(&newAdmin).Error; err != nil {
		log.Printf("Create error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	log.Println("Admin added successfully")
	return c.JSON(http.StatusOK, echo.Map{"message": "Admin added successfully", "admin": newAdmin})
}

func SuperAdminAddOrganization(c echo.Context) error {
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
	if roleID != 1 {
		log.Println("Permission denied: non-super admin trying to add organization")
		return c.JSON(http.StatusForbidden, echo.Map{"error": "Permission denied"})
	}

	var newOrganization models.Organization
	if err := c.Bind(&newOrganization); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	newOrganization.RoleID = 5

	if err := db.GetDB().Create(&newOrganization).Error; err != nil {
		log.Printf("Create error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	log.Println("Organization added successfully")
	return c.JSON(http.StatusOK, echo.Map{"message": "Organization added successfully", "organization": newOrganization})
}

func SuperAdminAddOrganizationAdmin(c echo.Context) error {
	// Retrieve userID and roleID from context
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
	if roleID != models.SuperAdminRoleID {
		log.Println("Permission denied: non-super admin trying to add organization admin")
		return c.JSON(http.StatusForbidden, echo.Map{"error": "Permission denied"})
	}

	orgIDStr := c.QueryParam("organizationID")
	log.Printf("Received organization ID from query: %s", orgIDStr)
	if orgIDStr == "" {
		log.Println("Organization ID is missing from the request")
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Organization ID is required"})
	}

	orgID, err := strconv.ParseUint(orgIDStr, 10, 64)
	if err != nil {
		log.Printf("Error parsing organization ID: %v", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid organization ID"})
	}

	// Check if the organization exists in the database
	var orgCount int64
	if err := db.GetDB().Model(&models.Organization{}).Where("id = ?", orgID).Count(&orgCount).Error; err != nil {
		log.Printf("Error querying organization: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Error querying organization"})
	}

	if orgCount == 0 {
		log.Println("Organization ID does not exist")
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Organization ID not found"})
	}

	// Bind request body to newUser
	var newUser models.User
	if err := c.Bind(&newUser); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	// Hash the password before saving the user
	hashedPassword, err := utils.HashPassword(newUser.Password)
	if err != nil {
		log.Printf("HashPassword error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not hash password"})
	}
	newUser.Password = hashedPassword

	newUser.RoleID = models.OrganizationAdminRoleID
	newUser.OrganizationID = uint(orgID)
	newUser.CreatedBy = uint(userID) // Set who created this user

	// Create new user in the database
	if err := db.GetDB().Create(&newUser).Error; err != nil {
		log.Printf("Create error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	log.Println("Organization admin added successfully")
	return c.JSON(http.StatusOK, echo.Map{"message": "Organization admin added successfully", "user": newUser})
}

func SoftDeleteOrganization(c echo.Context) error {
	id := c.Param("id")
	log.Printf("SoftDeleteOrganization called with ID: %s", id)

	// Retrieve roleID from the context
	roleID, ok := c.Get("roleID").(int)
	if !ok || roleID != 1 {
		log.Println("Failed to get roleID from context or insufficient permissions")
		return c.JSON(http.StatusForbidden, echo.Map{"error": "Only admins can delete organizations"})
	}

	var organization models.Organization

	// Find the organization by ID
	if err := db.GetDB().First(&organization, id).Error; err != nil {
		log.Printf("Error finding organization: %v", err)
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Organization not found"})
	}

	// Check if the organization is eligible for soft-deletion
	if organization.RoleID != 5 {
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

func ReactivateOrganization(c echo.Context) error {
    id := c.Param("id")
    log.Printf("ActivateOrganization called with ID: %s", id)

    // Retrieve roleID from the context (ensure only Super Admin can activate)
    roleID, ok := c.Get("roleID").(int)
    if !ok || roleID != 1 {
        log.Println("Failed to get roleID from context or insufficient permissions")
        return c.JSON(http.StatusForbidden, echo.Map{"error": "Only admins can activate organizations"})
    }

    var organization models.Organization

    // Find the organization by ID
    if err := db.GetDB().First(&organization, id).Error; err != nil {
        log.Printf("Error finding organization: %v", err)
        return c.JSON(http.StatusNotFound, echo.Map{"error": "Organization not found"})
    }

    // Activate the organization by setting it as active
    organization.IsActive = true
    if err := db.GetDB().Save(&organization).Error; err != nil {
        log.Printf("Error saving organization: %v", err)
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
    }

    log.Println("Organization activated successfully")
    return c.JSON(http.StatusOK, echo.Map{"message": "Organization activated successfully"})
}

func SuperAdminLogin(c echo.Context) error {
	var input struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.Bind(&input); err != nil {
		log.Printf("Bind error: %v", err)
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

	token, err := utils.GenerateJWT(user.ID, user.RoleID)
	if err != nil {
		log.Printf("GenerateJWT error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not generate token"})
	}

	log.Println("Super admin logged in successfully")
	return c.JSON(http.StatusOK, echo.Map{"token": token})
}

func SuperAdminLogout(c echo.Context) error {
	log.Println("Super admin logged out successfully")
	return c.JSON(http.StatusOK, echo.Map{"message": "Successfully logged out"})
}

func AdminLogin(c echo.Context) error {
	log.Println("AdminLogin - Entry")

	var input struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.Bind(&input); err != nil {
		log.Printf("AdminLogin - Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	log.Printf("AdminLogin - Received input: %+v", input)

	var user models.User
	if err := db.GetDB().Where("email = ?", input.Email).First(&user).Error; err != nil {
		log.Printf("AdminLogin - Where error: %v", err)
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid email or password"})
	}

	if err := utils.CheckPasswordHash(input.Password, user.Password); err != nil {
		log.Printf("AdminLogin - CheckPasswordHash error: %v", err)
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid email or password"})
	}

	token, err := utils.GenerateJWT(user.ID, user.RoleID)
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

func AdminAddUser(c echo.Context) error {
	log.Println("AdminAddUser - Entry")

	// Retrieve userID and roleID from context set by middleware
	userID, ok := c.Get("userID").(int)
	if !ok {
		log.Println("AdminAddUser - Unauthorized: userID not found in context")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	roleID, ok := c.Get("roleID").(int)
	if !ok {
		log.Println("AdminAddUser - Unauthorized: roleID not found in context")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	log.Printf("AdminAddUser - Received RoleID: %d, UserID: %d", roleID, userID)

	// Check if the roleID is 2 (Admin)
	if roleID != 2 {
		log.Println("AdminAddUser - Permission denied: non-admin trying to add user")
		return c.JSON(http.StatusForbidden, echo.Map{"error": "Permission denied"})
	}

	var input models.User
	if err := c.Bind(&input); err != nil {
		log.Printf("AdminAddUser - Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	log.Printf("AdminAddUser - New user data: %+v", input)

	// Validate roleID for new user
	if input.RoleID != 3 && input.RoleID != 4 && input.RoleID != 2 {
		log.Println("AdminAddUser - Invalid role ID provided")
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid role ID. Allowed roles: 3 (shopkeeper), 4 (auditor), 2 (admin)"})
	}

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

	// Retrieve roleID from context set by middleware
	roleID, ok := c.Get("roleID").(int)
	if !ok {
		log.Println("AdminViewAllUsers - Failed to get roleID from context")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	log.Printf("AdminViewAllUsers - Received RoleID: %d", roleID)

	// Check if the roleID is 2 (Admin)
	if roleID != 2 {
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
	log.Printf("SoftDeleteUser - Entry with ID: %s", id)

	var user models.User

	if user.RoleID == 1 { // Superadmin
		log.Println("SoftDeleteUser - Only superadmins can delete superadmin users")
		return c.JSON(http.StatusForbidden, "Only superadmins can delete superadmin users")
	}

	if err := db.GetDB().First(&user, id).Error; err != nil {
		log.Printf("SoftDeleteUser - First error: %v", err)
		return c.JSON(http.StatusNotFound, echo.Map{"error": "User not found"})
	}

	user.DeletedAt = gorm.DeletedAt{
		Time:  time.Now(),
		Valid: true,
	}

	if err := db.GetDB().Save(&user).Error; err != nil {
		log.Printf("SoftDeleteUser - Save error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	log.Println("SoftDeleteUser - User soft deleted successfully")
	log.Println("SoftDeleteUser - Exit")
	return c.JSON(http.StatusOK, echo.Map{"message": "User soft deleted successfully"})
}

func ActivateUser(c echo.Context) error {
	userID := c.Param("id")
	var user models.User

	if err := db.GetDB().First(&user, userID).Error; err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"message": "User not found"})
	}

	user.IsActive = true
	if err := db.GetDB().Save(&user).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Error saving user"})
	}

	return c.JSON(http.StatusOK, user)
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

func GetOrganizationByID(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	log.Printf("GetOrganizationByID - Entry with ID: %d", id)

	var org models.Organization
	if err := db.GetDB().First(&org, id).Error; err != nil {
		log.Printf("GetOrganizationByID - First error: %v", err)
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Organization not found"})
	}

	log.Printf("GetOrganizationByID - Organization found: %+v", org)
	log.Println("GetOrganizationByID - Exit")
	return c.JSON(http.StatusOK, org)
}

func GetAllOrganizations(c echo.Context) error {
	log.Println("GetAllOrganizations - Entry")

	var orgs []models.Organization
	if err := db.GetDB().Find(&orgs).Error; err != nil {
		log.Printf("GetAllOrganizations - Find error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Error retrieving organizations"})
	}

	log.Printf("GetAllOrganizations - Organizations found: %+v", orgs)
	log.Println("GetAllOrganizations - Exit")
	return c.JSON(http.StatusOK, orgs)
}

// ActivateOrganization activates an organization
func ActivateOrganization(c echo.Context) error {
	orgID := c.Param("id")
	log.Printf("Activating organization with ID: %s", orgID) // Log the incoming request

	var org models.Organization

	// Try to find the organization
	if err := db.GetDB().First(&org, orgID).Error; err != nil {
		log.Printf("Organization not found with ID: %s. Error: %v", orgID, err)
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Organization not found"})
	}

	// Log current status before activation
	log.Printf("Current status of organization ID %s: IsActive=%v", orgID, org.IsActive)

	org.IsActive = true
	// Try to save the updated organization
	if err := db.GetDB().Save(&org).Error; err != nil {
		log.Printf("Error saving organization ID %s. Error: %v", orgID, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Error saving organization"})
	}

	// Log success and updated status
	log.Printf("Successfully activated organization ID %s. Updated status: IsActive=%v", orgID, org.IsActive)
	return c.JSON(http.StatusOK, org)
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

func GetActiveOrganizations(c echo.Context) error {
	var orgs []models.Organization
	if err := db.GetDB().Where("is_active = ?", true).Find(&orgs).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Error retrieving organizations"})
	}
	return c.JSON(http.StatusOK, orgs)
}

func GetInactiveOrganizations(c echo.Context) error {
	var orgs []models.Organization
	if err := db.GetDB().Where("is_active = ?", false).Find(&orgs).Error; err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Error retrieving organizations"})
	}
	return c.JSON(http.StatusOK, orgs)
}

func AdminLogout(c echo.Context) error {
	log.Println("AdminLogout - Entry")
	log.Println("AdminLogout - Admin logged out successfully")
	log.Println("AdminLogout - Exit")
	return c.JSON(http.StatusOK, echo.Map{"message": "Successfully logged out"})
}

func Login(c echo.Context) error {
	log.Println("Login - Entry")

	var loginData struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.Bind(&loginData); err != nil {
		log.Printf("Login - Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	log.Printf("Login - Received data: %+v", loginData)

	var user models.User
	if err := db.GetDB().Where("email = ?", loginData.Email).First(&user).Error; err != nil {
		log.Printf("Login - Where error: %v", err)
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid email or password"})
	}

	if err := utils.CheckPasswordHash(loginData.Password, user.Password); err != nil {
		log.Printf("Login - CheckPasswordHash error: %v", err)
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid email or password"})
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": user.ID,
		"roleID": user.RoleID,
		"exp":    time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, err := token.SignedString(utils.JwtSecret)
	if err != nil {
		log.Printf("Login - SignedString error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not generate token"})
	}

	log.Println("Login - User logged in successfully")
	log.Println("Login - Exit")
	return c.JSON(http.StatusOK, echo.Map{"token": tokenString})
}

func Logout(c echo.Context) error {
	log.Println("Logout - Entry")
	log.Println("Logout - User logged out successfully")
	log.Println("Logout - Exit")
	return c.JSON(http.StatusOK, echo.Map{"message": "Successfully logged out"})
}

func AuditorLogin(c echo.Context) error {
	log.Println("AuditorLogin - Entry")

	var loginData struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.Bind(&loginData); err != nil {
		log.Printf("AuditorLogin - Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	log.Printf("AuditorLogin - Received data: %+v", loginData)

	var user models.User
	if err := db.GetDB().Where("email = ?", loginData.Email).First(&user).Error; err != nil {
		log.Printf("AuditorLogin - Where error: %v", err)
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid email or password"})
	}

	if err := utils.CheckPasswordHash(loginData.Password, user.Password); err != nil {
		log.Printf("AuditorLogin - CheckPasswordHash error: %v", err)
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid email or password"})
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID": user.ID,
		"roleID": user.RoleID,
		"exp":    time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, err := token.SignedString(utils.JwtSecret)
	if err != nil {
		log.Printf("AuditorLogin - SignedString error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not generate token"})
	}

	log.Println("AuditorLogin - Auditor logged in successfully")
	log.Println("AuditorLogin - Exit")
	return c.JSON(http.StatusOK, echo.Map{"token": tokenString})
}

func AuditorLogout(c echo.Context) error {
	log.Println("AuditorLogout - Entry")
	log.Println("AuditorLogout - Auditor logged out successfully")
	log.Println("AuditorLogout - Exit")
	return c.JSON(http.StatusOK, echo.Map{"message": "Successfully logged out"})
}
