package controllers

import (
	"log"
	"net/http"
	"stock/db"
	"stock/models"

	"github.com/labstack/echo/v4"
)

// GetCompanySettings fetches all company settings for a specific organization
func GetCompanySettings(c echo.Context) error {
	// Retrieve the organizationID from the context.
	// It is assumed that the getOrganizationID function extracts this value from the request context.
	organizationID, err := getOrganizationID(c)
	if err != nil {
		// If the organization ID could not be retrieved, return the error.
		return err
	}

	// Get the database instance by calling db.GetDB().
	// This function retrieves the DB connection, which is required for querying.
	db := db.GetDB() // Assuming this is a function to retrieve your DB instance
	if db == nil {
		// If the database instance is not available, log the error and return an internal server error.
		log.Println("Error: Failed to get database instance")
		return errorResponse(c, http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Ensure that missing organizations are inserted into company_settings table before querying.
	// The CheckAndInsertMissingOrganizations function is executed synchronously to guarantee missing records are added.
	err = CheckAndInsertMissingOrganizations(db)
	if err != nil {
		// If there is an error while inserting missing organizations, log it and return an internal server error.
		log.Printf("Error in CheckAndInsertMissingOrganizations: %v", err)
		return errorResponse(c, http.StatusInternalServerError, "Error inserting missing organizations")
	}

	// Initialize a slice to store the company settings that will be fetched from the database.
	var companySettings []models.CompanySetting

	// Query the database to fetch all company settings for the specific organization using the organizationID filter.
	// The Where clause ensures that only company settings related to the given organization are fetched.
	if err := db.Where("organization_id = ?", organizationID).Find(&companySettings).Error; err != nil {
		// If there is an error fetching the company settings, log it and return an internal server error.
		log.Printf("Error fetching company settings: %v", err)
		return errorResponse(c, http.StatusInternalServerError, "Failed to fetch company settings")
	}

	// Check if no company settings were found for the provided organizationID.
	if len(companySettings) == 0 {
		// If no settings are found, log the message and return a "not found" response.
		log.Printf("No company settings found for organization ID %d", organizationID)
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "No company settings found for this organization",
		})
	}

	// Here you can add additional processing logic (such as modifying fields or adding computed values).
	// In this case, KraPin is just being reassigned without any change, but this could be extended if necessary.
	for i := range companySettings {
		companySettings[i].KraPin = companySettings[i].KraPin // This line seems redundant but is kept for consistency.
	}

	// Return the company settings as a JSON response with a status of OK (200).
	// The response is automatically wrapped as an array of company settings.
	return c.JSON(http.StatusOK, companySettings) // Already an array, no further modifications needed
}

// UpdateCompanySettings updates the company settings for a specific organization
func UpdateCompanySettings(c echo.Context) error {
	// Retrieve the organizationID from the context.
	// This is fetched using a utility function that gets the organization ID associated with the request.
	organizationID, err := getOrganizationID(c)
	if err != nil {
		// If there is an error retrieving the organization ID, return the error.
		return err
	}

	// Bind the incoming request data to a CompanySetting struct.
	// This will extract the fields from the incoming JSON request body and assign them to the struct.
	var updatedSettings models.CompanySetting
	if err := c.Bind(&updatedSettings); err != nil {
		// If there is an error binding the incoming data, log the error and return a bad request response.
		log.Printf("Error binding request data: %v", err)
		return errorResponse(c, http.StatusBadRequest, "Invalid input data")
	}

	// Get the database instance to interact with the database.
	// The getDB() function is used to retrieve the database connection.
	db := getDB() // Assuming this is a function to retrieve your DB instance
	if db == nil {
		// If the database instance is not available, log the error and return an internal server error.
		log.Println("Error: Failed to get database instance")
		return errorResponse(c, http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Check if the company settings already exist for the given organization ID.
	// The Where clause is used to search for an existing company setting that matches the organizationID.
	var existingSetting models.CompanySetting
	if err := db.Where("organization_id = ?", organizationID).First(&existingSetting).Error; err != nil {
		// If no existing settings are found, log the error and return a "not found" response.
		log.Printf("Error finding company settings: %v", err)
		return errorResponse(c, http.StatusNotFound, "Company settings not found for this organization")
	}

	// If the company settings exist, update them with the new values provided in the updatedSettings struct.
	// The Model function is used to specify the record to be updated, and Updates is called to apply the changes.
	if err := db.Model(&existingSetting).Where("organization_id = ?", organizationID).Updates(updatedSettings).Error; err != nil {
		// If there is an error while updating the company settings, log it and return an internal server error.
		log.Printf("Error updating company settings: %v", err)
		return errorResponse(c, http.StatusInternalServerError, "Failed to update company settings")
	}

	// Optional step: Add KraPin to the updated company settings (this line is redundant as KraPin is already part of the struct).
	// This is just for consistency in case additional logic is needed.
	updatedSettings.KraPin = updatedSettings.KraPin

	// Return the updated company settings as a JSON response.
	// The updatedSettings is wrapped in an array to keep the response format consistent with the GetCompanySettings function.
	return c.JSON(http.StatusOK, []models.CompanySetting{updatedSettings}) // Wrapping in an array
}