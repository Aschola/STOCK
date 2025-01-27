package controllers

import (
	"log"
	"net/http"
	"stock/models"

	"github.com/labstack/echo/v4"
)

// GetCompanySettings fetches all company settings for a specific organization
func GetCompanySettings(c echo.Context) error {
	// Retrieve organizationID from context using the existing utility function
	organizationID, err := getOrganizationID(c)
	if err != nil {
		return err
	}

	// Get the database instance
	db := getDB() // Assuming this is a function to retrieve your DB instance
	if db == nil {
		log.Println("Error: Failed to get database instance")
		return errorResponse(c, http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Initialize a slice to store the company settings
	var companySettings []models.CompanySetting

	// Query the database to fetch all company settings for the specific organization
	if err := db.Where("organization_id = ?", organizationID).Find(&companySettings).Error; err != nil {
		log.Printf("Error fetching company settings: %v", err)
		return errorResponse(c, http.StatusInternalServerError, "Failed to fetch company settings")
	}

	// Check if no settings are found
	if len(companySettings) == 0 {
		log.Printf("No company settings found for organization ID %d", organizationID)
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "No company settings found for this organization",
		})
	}

	// Return the company settings as JSON
	return c.JSON(http.StatusOK, companySettings)
}

// UpdateCompanySettings updates the company settings for a specific organization
func UpdateCompanySettings(c echo.Context) error {
	// Retrieve organizationID from context using the existing utility function
	organizationID, err := getOrganizationID(c)
	if err != nil {
		return err
	}

	// Bind the request body to a CompanySetting struct to get the updated details
	var updatedSettings models.CompanySetting
	if err := c.Bind(&updatedSettings); err != nil {
		log.Printf("Error binding request data: %v", err)
		return errorResponse(c, http.StatusBadRequest, "Invalid input data")
	}

	// Get the database instance
	db := getDB() // Assuming this is a function to retrieve your DB instance
	if db == nil {
		log.Println("Error: Failed to get database instance")
		return errorResponse(c, http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Check if the company setting for the organization exists
	var existingSetting models.CompanySetting
	if err := db.Where("organization_id = ?", organizationID).First(&existingSetting).Error; err != nil {
		log.Printf("Error finding company settings: %v", err)
		return errorResponse(c, http.StatusNotFound, "Company settings not found for this organization")
	}

	// Update the company settings (only the ones that are provided in updatedSettings)
	if err := db.Model(&existingSetting).Where("organization_id = ?", organizationID).Updates(updatedSettings).Error; err != nil {
		log.Printf("Error updating company settings: %v", err)
		return errorResponse(c, http.StatusInternalServerError, "Failed to update company settings")
	}

	// Return the updated company settings
	return c.JSON(http.StatusOK, updatedSettings)
}
