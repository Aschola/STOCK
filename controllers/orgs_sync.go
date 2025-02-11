package controllers

import (
	"log"
	"stock/models"

	"gorm.io/gorm"
)

// CheckAndInsertMissingOrganizations checks if any organization_id is missing in company_settings
// and inserts records for the missing organization_ids into the company_settings table.
func CheckAndInsertMissingOrganizations(db *gorm.DB) error {
	// Defer the recovery function to catch any panics during the execution of the function.
	defer func() {
		// If a panic occurs, this block will execute and log the panic message.
		if r := recover(); r != nil {
			log.Printf("Recovered from panic: %v", r)
		}
	}()

	// Step 1: Retrieve all organizations from the 'organizations' table.
	var organizations []models.Organization
	// Query the database to fetch all organization records. If there's an error, log it and return the error.
	if err := db.Find(&organizations).Error; err != nil {
		log.Printf("Error retrieving organizations: %v", err)
		return err // Return the error to halt the execution if organizations are not fetched properly.
	}

	// Step 2: Retrieve all company settings from the 'company_settings' table.
	var companySettings []models.CompanySetting
	// Query the database to fetch all company setting records. If there's an error, log it and return the error.
	if err := db.Find(&companySettings).Error; err != nil {
		log.Printf("Error retrieving company settings: %v", err)
		return err // Return the error to halt the execution if company settings are not fetched properly.
	}

	// Step 3: Create a map to store the existing organization IDs that are already present in company_settings.
	existingOrganizationIDs := make(map[uint]bool)
	// Loop through the existing company settings and store each organization_id in the map.
	for _, cs := range companySettings {
		// The key is the organization ID, and the value is a boolean indicating its presence.
		existingOrganizationIDs[cs.OrganizationID] = true
	}

	// Step 4: Iterate over all organizations and check for missing organization_ids in company_settings.
	missingOrgFound := false // Flag to track if any missing organizations were found and inserted.

	for _, org := range organizations {
		// If the organization_id already exists in company_settings, skip it and move to the next organization.
		if existingOrganizationIDs[org.ID] {
			continue // Skip this organization and move to the next one.
		}

		// Step 5: If the organization_id is missing from company_settings, insert a new record.
		newCompanySetting := models.CompanySetting{
			Name:           org.Name, // Use the organization's name as the company setting name.
			Address:        "",       // Leave the address empty initially (can be updated later).
			Telephone:      "",       // Leave the telephone number empty initially (can be updated later).
			OrganizationID: org.ID,   // Use the organization's ID as the foreign key in company_settings.
		}

		// Attempt to create a new record in the company_settings table.
		if err := db.Create(&newCompanySetting).Error; err != nil {
			// If there's an error while inserting the new company setting, log it.
			log.Printf("Error inserting missing organization: %v", err)
		} else {
			// If the insertion is successful, log the inserted organization ID.
			log.Printf("Inserted missing organization_id %d into company_settings", org.ID)
			// Set the flag to true to indicate that a missing organization was found and inserted.
			missingOrgFound = true
		}
	}

	// Step 6: After checking all organizations, log if no missing organizations were found.
	if !missingOrgFound {
		// This message will be logged if no organization was missing in the company_settings.
		log.Println("No missing organizations found in company_settings.")
	}

	// Return nil to indicate the operation was successful (even if no organizations were inserted).
	return nil
}