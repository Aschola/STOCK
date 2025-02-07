package controllers

import (
	"log"
	"stock/models"
	"time"

	"gorm.io/gorm"
)

// CheckAndInsertMissingOrganizations checks every 30 seconds if any organization_id is missing in company_settings and inserts missing records
func CheckAndInsertMissingOrganizations(db *gorm.DB) {

	// Defer the recovery function to catch panics
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic: %v", r)
		}
	}()

	ticker := time.NewTicker(30 * time.Second) // Set the ticker to 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Step 1: Get all organization_ids from the 'organizations' table with retry logic
			var organizations []models.Organization
			const maxRetries = 3
			var err error

			for i := 0; i < maxRetries; i++ {
				if err = db.Find(&organizations).Error; err != nil {
					log.Printf("Error retrieving organizations: %v", err)
					time.Sleep(time.Second * time.Duration(i)) // Exponential backoff
					continue
				}
				break
			}

			if err != nil {
				// If after maxRetries, we still have an error, log and continue to the next iteration
				log.Printf("Failed to retrieve organizations after %d retries: %v", maxRetries, err)
				continue
			}

			// Step 2: Get all organization_ids from the 'company_settings' table
			var companySettings []models.CompanySetting
			if err := db.Find(&companySettings).Error; err != nil {
				log.Printf("Error retrieving company settings: %v", err)
				continue
			}

			// Step 3: Create a map of existing organization_ids in company_settings
			existingOrganizationIDs := make(map[uint]bool)
			for _, cs := range companySettings {
				existingOrganizationIDs[cs.OrganizationID] = true
			}

			// Step 4: Check for missing organization_ids and insert them into company_settings
			missingOrgFound := false // Flag to track if any missing organizations were found

			for _, org := range organizations {
				// Skip if the organization_id already exists in company_settings
				if existingOrganizationIDs[org.ID] {
					continue
				}

				// Step 5: Insert the missing organization_id into company_settings with empty address and telephone
				newCompanySetting := models.CompanySetting{
					Name:           org.Name,
					Address:        "", // Leave address empty
					Telephone:      "", // Leave telephone empty
					OrganizationID: org.ID,
				}

				if err := db.Create(&newCompanySetting).Error; err != nil {
					log.Printf("Error inserting missing organization: %v", err)
				} else {
					log.Printf("Inserted missing organization_id %d into company_settings", org.ID)
					missingOrgFound = true
				}
			}

			// Step 6: Log a message if no missing organizations were found
			if !missingOrgFound {
				log.Println("No missing organizations found in company_settings.")
			}
		}
	}
}