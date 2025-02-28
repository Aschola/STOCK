package controllers

import (
	"fmt"
	"log"
	"time"
	"stock/models"

	"gorm.io/gorm"
)

// StartDailySalesSummary schedules and runs the daily sales summary task at midnight.
func StartDailySalesSummary(db *gorm.DB) {
	go func() {
		for {
			now := time.Now()
			// Schedule the next execution at midnight before the next day starts
			nextRun := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location()) 

			// Ensure the nextRun is in the future, otherwise adjust to the following day
			if now.After(nextRun) {
				nextRun = nextRun.Add(24 * time.Hour)
			}

			sleepDuration := time.Until(nextRun)
			log.Printf("[INFO] [%s] Next daily sales summary scheduled for: %s\n", now.Format("2006-01-02 15:04:05"), nextRun)

			// Wait until the scheduled time
			time.Sleep(sleepDuration) 

			// Execute the summary process
			if err := SummarizeDailySales(db); err != nil {
				log.Printf("[ERROR] [%s] Failed to summarize daily sales: %v\n", time.Now().Format("2006-01-02 15:04:05"), err)
			} else {
				log.Printf("[INFO] [%s] Daily sales summary completed successfully\n", time.Now().Format("2006-01-02 15:04:05"))
			}
		}
	}()
}

// SummarizeDailySales aggregates and updates daily sales for each organization.
func SummarizeDailySales(db *gorm.DB) error {
	var salesData []struct {
		TotalSellingPrice float64
		OrganizationID    uint
	}

	// Get yesterday's date (we summarize sales for the completed day)
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	log.Printf("[INFO] [%s] Starting sales summary for %s\n", time.Now().Format("2006-01-02 15:04:05"), yesterday)

	// Fetch total selling price per organization for yesterday
	err := db.Model(&models.Sale{}).
		Select("SUM(total_selling_price) as total_selling_price, organizations_id as organization_id").
		Where("DATE(date) = ?", yesterday).
		Group("organizations_id").
		Scan(&salesData).Error

	// Error handling for query execution
	if err != nil {
		log.Printf("[ERROR] [%s] Error retrieving sales data for %s: %v\n", time.Now().Format("2006-01-02 15:04:05"), yesterday, err)
		return fmt.Errorf("error calculating total sales: %w", err)
	}

	// No sales data found, log and exit gracefully
	if len(salesData) == 0 {
		log.Printf("[INFO] [%s] No sales data found for %s. Skipping summary.\n", time.Now().Format("2006-01-02 15:04:05"), yesterday)
		return nil
	}

	// Process sales data for each organization
	for _, sale := range salesData {
		var existingRecord models.TotalSales

		// Check if a record already exists for this organization for yesterday
		result := db.Where("organization_id = ? AND DATE(date) = ?", sale.OrganizationID, yesterday).
			First(&existingRecord)

		if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
			// Log unexpected database errors
			log.Printf("[ERROR] [%s] Database error checking existing record for OrgID %d on %s: %v\n",
				time.Now().Format("2006-01-02 15:04:05"), sale.OrganizationID, yesterday, result.Error)
			continue // Skip this iteration to avoid further failures
		}

		if result.RowsAffected == 0 {
			// Insert new record
			totalSale := models.TotalSales{
				TotalSellingPrice: sale.TotalSellingPrice,
				OrganizationID:    sale.OrganizationID,
				Date:              time.Now(),
			}

			if err := db.Create(&totalSale).Error; err != nil {
				log.Printf("[ERROR] [%s] Failed to insert total sales for OrgID %d on %s: %v\n",
					time.Now().Format("2006-01-02 15:04:05"), sale.OrganizationID, yesterday, err)
			} else {
				log.Printf("[INFO] [%s] Inserted total sales for OrgID %d: Amount %.2f for %s\n",
					time.Now().Format("2006-01-02 15:04:05"), sale.OrganizationID, sale.TotalSellingPrice, yesterday)
			}
		} else {
			// Update existing record
			err := db.Model(&models.TotalSales{}).
				Where("organization_id = ? AND DATE(date) = ?", sale.OrganizationID, yesterday).
				Update("total_selling_price", sale.TotalSellingPrice).Error

			if err != nil {
				log.Printf("[ERROR] [%s] Failed to update total sales for OrgID %d on %s: %v\n",
					time.Now().Format("2006-01-02 15:04:05"), sale.OrganizationID, yesterday, err)
			} else {
				log.Printf("[INFO] [%s] Updated total sales for OrgID %d: Amount %.2f for %s\n",
					time.Now().Format("2006-01-02 15:04:05"), sale.OrganizationID, sale.TotalSellingPrice, yesterday)
			}
		}
	}

	log.Printf("[INFO] [%s] Daily sales summary process completed successfully for %s\n",
		time.Now().Format("2006-01-02 15:04:05"), yesterday)

	return nil
}





// package controllers

// import (
// 	"fmt"
// 	"log"
// 	"time"
// 	"stock/models"

// 	"gorm.io/gorm"
// )

// // StartDailySalesSummary runs a background job at 09:03 AM every day
// func StartDailySalesSummary(db *gorm.DB) {
// 	go func() {
// 		for {
// 			now := time.Now()
// 			nextRun := time.Date(now.Year(), now.Month(), now.Day(), 23, 58, 0, 0, now.Location()) // 09:03 AM
// 			//nextRun := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location()) // 00:00 AM


// 			// If past 09:03 AM, schedule for the next day
// 			if now.After(nextRun) {
// 				nextRun = nextRun.Add(24 * time.Hour)
// 			}

// 			sleepDuration := time.Until(nextRun)
// 			log.Println("[INFO] Next daily sales summary scheduled for:", nextRun)

// 			time.Sleep(sleepDuration) // Wait until 09:03 AM
// 			if err := SummarizeDailySales(db); err != nil {
// 				log.Println("[ERROR] Failed to summarize daily sales:", err)
// 			} else {
// 				log.Println("[INFO] Daily sales summary completed successfully at", time.Now())
// 			}
// 		}
// 	}()
// }


// func SummarizeDailySales(db *gorm.DB) error {
// 	var salesData []struct {
// 		TotalSellingPrice float64
// 		OrganizationID    uint
// 	}

// 	// Get today's date in "YYYY-MM-DD" format
// 	today := time.Now().Format("2006-01-02")

// 	// Fetch total selling price per organization for today
// 	err := db.Model(&models.Sale{}).
// 		Select("SUM(total_selling_price) as total_selling_price, organizations_id as organization_id").
// 		Where("DATE(date) = ?", today).
// 		Group("organizations_id").
// 		Scan(&salesData).Error

// 	if err != nil {
// 		return fmt.Errorf("error calculating total sales: %w", err)
// 	}

// 	// Insert or update total_sales per organization
// 	for _, sale := range salesData {
// 		var existingRecord models.TotalSales

// 		// Check if a record exists for today
// 		result := db.Where("organization_id = ? AND DATE(date) = ?", sale.OrganizationID, today).
// 			First(&existingRecord)

// 		if result.RowsAffected == 0 {
// 			// Insert new record
// 			totalSale := models.TotalSales{
// 				TotalSellingPrice: sale.TotalSellingPrice,
// 				OrganizationID:    sale.OrganizationID,
// 				Date:              time.Now(),
// 			}

// 			if err := db.Create(&totalSale).Error; err != nil {
// 				log.Println("[ERROR] Failed to insert total sales for OrgID:", sale.OrganizationID, err)
// 			} else {
// 				log.Println("[INFO] Total sales inserted for OrgID:", sale.OrganizationID, "Amount:", sale.TotalSellingPrice)
// 			}
// 		} else {
// 			// ✅ FIX: Ensure WHERE conditions in the update query
// 			err := db.Model(&models.TotalSales{}).
// 				Where("organization_id = ? AND DATE(date) = ?", sale.OrganizationID, today).
// 				Update("total_selling_price", sale.TotalSellingPrice).Error

// 			if err != nil {
// 				log.Println("[ERROR] Failed to update total sales for OrgID:", sale.OrganizationID, err)
// 			} else {
// 				log.Println("[INFO] Total sales updated for OrgID:", sale.OrganizationID, "Amount:", sale.TotalSellingPrice)
// 			}
// 		}
// 	}

// 	return nil
// } 