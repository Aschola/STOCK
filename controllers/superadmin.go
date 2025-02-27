package controllers

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"

	"stock/db"
	"stock/models"
	"stock/utils"
	"stock/validators"

	"github.com/labstack/echo/v4"
)

func SuperAdminLogin(c echo.Context) error {
	var input struct {
		Email string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.Bind(&input); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	// loginInput := validators.LoginInput{
	// 	//Username: input.Username,
	// 	Password: input.Password,
	// }
	// if err := validators.ValidateLoginInput(loginInput); err != nil {
	// 	log.Printf("Validation error: %v", err)
	// 	return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	// }

	var user models.User
	if err := db.GetDB().Where("email = ?", input.Email).First(&user).Error; err != nil {
		log.Printf("Where error: %v", err)
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid email or password"})
	}
	if !user.IsActive {
		log.Printf("Login attempt by inactive user: %s", input.Email)
		return c.JSON(http.StatusForbidden, echo.Map{"error": "Account is inactive. Please contact administrator"})
	}

	if err := utils.CheckPasswordHash(input.Password, user.Password); err != nil {
		log.Printf("CheckPasswordHash error: %v", err)
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid email or password"})
	}
	token, err := utils.GenerateJWT(user.ID, user.RoleName, user.OrganizationID)
if err != nil {
	log.Printf("GenerateJWT error: %v", err)
	return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not generate token"})
}

return c.JSON(http.StatusOK, echo.Map{
	"token":    token,
	"user_id":  user.ID,
	"username": user.Username,
	"role":     user.RoleName,
})


	// token, err := utils.GenerateJWT(user.ID, user.RoleName, user.OrganizationID)
	// if err != nil {
	// 	log.Printf("GenerateJWT error: %v", err)
	// 	return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not generate token"})
	// }


	// log.Println("Super admin logged in successfully")
	// return c.JSON(http.StatusOK, echo.Map{"token": token})
}
func SuperAdminLogout(c echo.Context) error {
	log.Println("Super admin logged out successfully")
	return c.JSON(http.StatusOK, echo.Map{"message": "Successfully logged out"})
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

func ActivateOrganization(c echo.Context) error {
	orgID := c.Param("id")
	log.Printf("Activating organization with ID: %s", orgID)

	var org models.Organization

	if err := db.GetDB().First(&org, orgID).Error; err != nil {
		log.Printf("Organization not found with ID: %s. Error: %v", orgID, err)
		return c.JSON(http.StatusNotFound, map[string]string{"message": "Organization not found"})
	}

	log.Printf("Current status of organization name %s: IsActive=%v", orgID, org.IsActive)

	org.IsActive = true
	if err := db.GetDB().Save(&org).Error; err != nil {
		log.Printf("Error saving organization name %s. Error: %v", orgID, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"message": "Error saving organization"})
	}

	log.Printf("Successfully activated organization ID %s. Updated status: IsActive=%v", orgID, org.IsActive)
	return c.JSON(http.StatusOK, org)
}
func DeleteOrganization(c echo.Context) error {
	id := c.Param("id")
	log.Printf("DeleteOrganization called with ID: %s", id)

	roleName, ok := c.Get("roleName").(string)
	if !ok || roleName != "Superadmin" {
		log.Println("Failed to get roleName from context or insufficient permissions")
		return c.JSON(http.StatusForbidden, echo.Map{"error": "Only Super Admin can delete organizations"})
	}

	var org models.Organization
	if err := db.GetDB().First(&org, id).Error; err != nil {
		log.Printf("Error finding organization: %v", err)
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Organization not found"})
	}

	if err := db.GetDB().Delete(&org).Error; err != nil {
		log.Printf("Error deleting organization: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	log.Println("Organization deleted successfully")
	return c.JSON(http.StatusOK, echo.Map{"message": "Organization deleted successfully"})
}
func EditOrganization(c echo.Context) error {
	id := c.Param("id")
	log.Printf("EditOrganization - Entry with ID: %s", id)

	var org models.Organization
	if err := db.GetDB().First(&org, id).Error; err != nil {
		log.Printf("EditOrganization - First error: %v", err)
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Organization not found"})
	}

	log.Printf("EditOrganization - Current organization details: %+v", org)

	if err := c.Bind(&org); err != nil {
		log.Printf("EditOrganization - Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	if err := db.GetDB().Save(&org).Error; err != nil {
		log.Printf("EditOrganization - Save error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	log.Println("EditOrganization - Organization updated successfully")
	log.Println("EditOrganization - Exit")
	return c.JSON(http.StatusOK, org)
}

func SoftDeleteOrganization(c echo.Context) error {
	id := c.Param("id")
	log.Printf("SoftDeleteOrganization called with ID: %s", id)

	roleName, ok := c.Get("roleName").(string)
	if !ok || roleName != "Superadmin" {
		log.Println("Failed to get roleName from context or insufficient permissions")
		return c.JSON(http.StatusForbidden, echo.Map{"error": "Only Superadmin can delete organizations"})
	}

	var org models.Organization
	if err := db.GetDB().First(&org, id).Error; err != nil {
		log.Printf("Error finding organization: %v", err)
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Organization not found"})
	}

	org.IsActive = false
	if !strings.HasPrefix(org.Name, "deleted_") {
		org.Name = "deleted_" + org.Name
	}

	if err := db.GetDB().Save(&org).Error; err != nil {
		log.Printf("Error saving organization: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	log.Println("Organization soft deleted (deactivated) successfully")
	return c.JSON(http.StatusOK, echo.Map{"message": "Organization soft deleted successfully", "organization": org})
}

func ReactivateOrganization(c echo.Context) error {
	id := c.Param("id")
	log.Printf("ReactivateOrganization called with ID: %s", id)

	roleName, ok := c.Get("roleName").(string)
	if !ok || roleName != "Admin" {
		log.Println("Failed to get roleName from context or insufficient permissions")
		return c.JSON(http.StatusForbidden, echo.Map{"error": "Only admins can reactivate organizations"})
	}

	var organization models.Organization

	if err := db.GetDB().First(&organization, id).Error; err != nil {
		log.Printf("Error finding organization: %v", err)
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Organization not found"})
	}

	if organization.IsActive {
		log.Println("Organization is already active")
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Organization is already active"})
	}

	organization.IsActive = true
	if err := db.GetDB().Save(&organization).Error; err != nil {
		log.Printf("Error saving organization: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	if err := db.GetDB().Model(&models.User{}).Where("organization_id = ?", id).Updates(map[string]interface{}{"is_active": true}).Error; err != nil {
		log.Printf("Error reactivating users: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Error reactivating users"})
	}

	log.Println("Organization and associated users reactivated successfully")
	return c.JSON(http.StatusOK, echo.Map{"message": "Organization and associated users reactivated successfully"})
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

func AddAdmin(c echo.Context) error {
	log.Println("AddAdmin called")

	roleName, ok := c.Get("roleName").(string)
	if !ok {
		log.Println("Failed to get roleName from context")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	userID, ok := c.Get("userID").(uint)
	if !ok {
		log.Println("Failed to get userID from context")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	log.Printf("Received RoleName: %s, UserID: %d", roleName, userID)

	if roleName != "Superadmin" {
		log.Println("Permission denied: only superadmin can add admin")
		return c.JSON(http.StatusForbidden, echo.Map{"error": "Permission denied"})
	}

	var newAdmin models.User
	if err := c.Bind(&newAdmin); err != nil {
		log.Printf("Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	log.Printf("New admin data: %+v", newAdmin)

	//validations
	signupInput := validators.SignupInput{
		Username: newAdmin.Username,
		Password: newAdmin.Password,
	}
	if err := validators.ValidateSignupInput(signupInput); err != nil {
		log.Printf("Validation error: %v", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	newAdmin.CreatedBy = uint(userID)

	newAdmin.RoleName = "Admin"

	log.Printf("Saving new admin to database")

	// Save the new admin in the database
	if err := db.GetDB().Create(&newAdmin).Error; err != nil {
		log.Printf("Create error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	log.Println("Admin added successfully")
	return c.JSON(http.StatusOK, echo.Map{"message": "Admin added successfully", "admin": newAdmin})
}

// func SuperAdminAddOrganization(c echo.Context) error {
// 	userID, ok := c.Get("userID").(uint)
// 	if !ok {
// 		log.Println("Failed to get userID from context")
// 		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
// 	}

// 	roleName, ok := c.Get("roleName").(string)
// 	if !ok {
// 		log.Println("Failed to get roleName from context")
// 		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
// 	}

// 	log.Printf("Received UserID: %d, RoleName: %s", userID, roleName)

// 	if roleName != "Superadmin" {
// 		log.Println("Permission denied: non-super admin trying to add an organization")
// 		return c.JSON(http.StatusForbidden, echo.Map{"error": "Permission denied"})
// 	}

// 	// Bind request body to the organization model
// 	var newOrganization models.Organization
// 	if err := c.Bind(&newOrganization); err != nil {
// 		log.Printf("Bind error: %v", err)
// 		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
// 	}

// 	// Validate the organization details
// 	if newOrganization.Name == "" {
// 		log.Println("Organization name is missing")
// 		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Organization name is required"})
// 	}

// 	//newOrganization.CreatedBy = uint(userID)

// 	if err := db.GetDB().Create(&newOrganization).Error; err != nil {
// 		log.Printf("Create error: %v", err)
// 		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
// 	}

// 	log.Println("Organization added successfully")
// 	return c.JSON(http.StatusOK, echo.Map{"message": "Organization added successfully", "organization": newOrganization})
// }

// func SuperAdminAddOrganizationAdmin(c echo.Context) error {
// 	// Retrieve userID and roleName from context
// 	userID, ok := c.Get("userID").(int)
// 	if !ok {
// 		log.Println("Failed to get userID from context")
// 		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
// 	}

// 	roleName, ok := c.Get("roleName").(string)
// 	if !ok {
// 		log.Println("Failed to get roleName from context")
// 		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
// 	}

// 	log.Printf("Received UserID: %d, RoleName: %s", userID, roleName)

// 	// Check if the roleName is for a Super Admin (roleName = "SuperAdmin")
// 	if roleName != "SuperAdmin" {
// 		log.Println("Permission denied: non-super admin trying to add organization admin")
// 		return c.JSON(http.StatusForbidden, echo.Map{"error": "Permission denied"})
// 	}

// 	orgIDStr := c.QueryParam("organizationID")
// 	log.Printf("Received organization ID from query: %s", orgIDStr)
// 	if orgIDStr == "" {
// 		log.Println("Organization ID is missing from the request")
// 		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Organization ID is required"})
// 	}

// 	orgID, err := strconv.ParseUint(orgIDStr, 10, 64)
// 	if err != nil {
// 		log.Printf("Error parsing organization ID: %v", err)
// 		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid organization ID"})
// 	}

// 	// Check if the organization exists in the database
// 	var orgCount int64
// 	if err := db.GetDB().Model(&models.Organization{}).Where("id = ?", orgID).Count(&orgCount).Error; err != nil {
// 		log.Printf("Error querying organization: %v", err)
// 		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Error querying organization"})
// 	}

// 	if orgCount == 0 {
// 		log.Println("Organization ID does not exist")
// 		return c.JSON(http.StatusNotFound, echo.Map{"error": "Organization ID not found"})
// 	}

// 	var newUser models.User
// 	if err := c.Bind(&newUser); err != nil {
// 		log.Printf("Bind error: %v", err)
// 		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
// 	}

// 	hashedPassword, err := utils.HashPassword(newUser.Password)
// 	if err != nil {
// 		log.Printf("HashPassword error: %v", err)
// 		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not hash password"})
// 	}
// 	newUser.Password = hashedPassword

// 	newUser.RoleName = models.OrganizationAdminRoleName
// 	newUser.OrganizationID = uint(orgID)
// 	newUser.CreatedBy = uint(userID)

// 	// Create new user in the database
// 	if err := db.GetDB().Create(&newUser).Error; err != nil {
// 		log.Printf("Create error: %v", err)
// 		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
// 	}

// 	log.Println("Organization admin added successfully")
// 	return c.JSON(http.StatusOK, echo.Map{"message": "Organization admin added successfully", "user": newUser})
// }


func Login(c echo.Context) error {
	log.Println("Login - Entry")

	var loginData struct {
		Email string `json:"email" binding:"required"`
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
	if !user.IsActive {
		log.Printf("Login attempt by inactive user: %s", user.Email)
		return c.JSON(http.StatusForbidden, echo.Map{"error": "Account is inactive. Please contact administrator"})
	}

	var organization models.Organization
	if err := db.GetDB().Where("id = ?", user.OrganizationID).First(&organization).Error; err != nil {
		log.Printf("Login - Organization lookup error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not verify organization status"})
	}
	if !user.IsActive {
		log.Printf("Login - User is inactive: %v", user.ID)
		return c.JSON(http.StatusForbidden, echo.Map{"error": "user inactive"})
	}

	if !organization.IsActive {
		log.Printf("Login - Organization is inactive: %v", organization.ID)
		return c.JSON(http.StatusForbidden, echo.Map{"error": "Organization is inactive"})
	}

	if err := utils.CheckPasswordHash(loginData.Password, user.Password); err != nil {
		log.Printf("Login - CheckPasswordHash error: %v", err)
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid email or password"})
	}

	token, err := utils.GenerateJWT(user.ID, user.RoleName, user.OrganizationID)
	if err != nil {
		log.Printf("AdminLogin - GenerateJWT error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not generate token"})
	}

	log.Println("AdminLogin - Admin logged in successfully")
	log.Println("AdminLogin - Exit")
	//return c.JSON(http.StatusOK, echo.Map{"token": token})
	return c.JSON(http.StatusOK, echo.Map{
		"user_id": user.ID,
		"organization": user.OrganizationID,
		"token": token,
		"user":user.Username,
		"role_name": user.RoleName,
		"redirectUrl": "/login",
	})
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
		Email string `json:"username" binding:"required"`
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
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "email or password"})
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"userID":   user.ID,
		"roleName": user.RoleName,
		"organization": user.OrganizationID,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
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
