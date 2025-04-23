package controllers

import (
	"fmt"
	"log"
	"time"
	"stock/models"

	"gorm.io/gorm"
)

// StartDailySalesSummary schedules and runs the daily sales summary task at midnight.
// func StartDailySalesSummary(db *gorm.DB) {
// 	go func() {
// 		for {

// 			 now := time.Now().AddDate(0, 0, -1) 
// 			// Schedule the next execution at midnight before the next day starts
// 			nextRun := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location()) 

// 			// Ensure the nextRun is in the future, otherwise adjust to the following day
// 			if now.After(nextRun) {
// 				nextRun = nextRun.Add(24 * time.Hour)
// 			}

// 			sleepDuration := time.Until(nextRun)
// 			log.Printf("[INFO] [%s] Next daily sales summary scheduled for: %s\n", now.Format("2006-01-02 15:04:05"), nextRun)

// 			// Wait until the scheduled time
// 			time.Sleep(sleepDuration) 

// 			// Execute the summary process
// 			if err := SummarizeDailySales(db); err != nil {
// 				log.Printf("[ERROR] [%s] Failed to summarize daily sales: %v\n", time.Now().Format("2006-01-02 15:04:05"), err)
// 			} else {
// 				log.Printf("[INFO] [%s] Daily sales summary completed successfully\n", time.Now().Format("2006-01-02 15:04:05"))
// 			}
// 		}
// 	}()
// }

// // SummarizeDailySales aggregates and updates daily sales for each organization.
// func SummarizeDailySales(db *gorm.DB) error {
// 	var salesData []struct {
// 		TotalSellingPrice float64
// 		OrganizationID    uint
// 	}

// 	// Get yesterday's date (we summarize sales for the completed day)
// 	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
// 	log.Printf("[INFO] [%s] Starting sales summary for %s\n", time.Now().Format("2006-01-02 15:04:05"), yesterday)

// 	// Fetch total selling price per organization for yesterday
// 	err := db.Model(&models.Sale{}).
// 		Select("SUM(total_selling_price) as total_selling_price, organizations_id as organization_id").
// 		Where("DATE(date) = ?", yesterday).
// 		Group("organizations_id").
// 		Scan(&salesData).Error

// 	// Error handling for query execution
// 	if err != nil {
// 		log.Printf("[ERROR] [%s] Error retrieving sales data for %s: %v\n", time.Now().Format("2006-01-02 15:04:05"), yesterday, err)
// 		return fmt.Errorf("error calculating total sales: %w", err)
// 	}

// 	// No sales data found, log and exit gracefully
// 	if len(salesData) == 0 {
// 		log.Printf("[INFO] [%s] No sales data found for %s. Skipping summary.\n", time.Now().Format("2006-01-02 15:04:05"), yesterday)
// 		return nil
// 	}

// 	// Process sales data for each organization
// 	for _, sale := range salesData {
// 		var existingRecord models.TotalSales

// 		// Check if a record already exists for this organization for yesterday
// 		result := db.Where("organization_id = ? AND DATE(date) = ?", sale.OrganizationID, yesterday).
// 			First(&existingRecord)

// 		if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
// 			// Log unexpected database errors
// 			log.Printf("[ERROR] [%s] Database error checking existing record for OrgID %d on %s: %v\n",
// 				time.Now().Format("2006-01-02 15:04:05"), sale.OrganizationID, yesterday, result.Error)
// 			continue // Skip this iteration to avoid further failures
// 		}

// 		if result.RowsAffected == 0 {
// 			// Insert new record
// 			totalSale := models.TotalSales{
// 				TotalSellingPrice: sale.TotalSellingPrice,
// 				OrganizationID:    sale.OrganizationID,
// 				Date:              time.Now(),
// 			}

// 			if err := db.Create(&totalSale).Error; err != nil {
// 				log.Printf("[ERROR] [%s] Failed to insert total sales for OrgID %d on %s: %v\n",
// 					time.Now().Format("2006-01-02 15:04:05"), sale.OrganizationID, yesterday, err)
// 			} else {
// 				log.Printf("[INFO] [%s] Inserted total sales for OrgID %d: Amount %.2f for %s\n",
// 					time.Now().Format("2006-01-02 15:04:05"), sale.OrganizationID, sale.TotalSellingPrice, yesterday)
// 			}
// 		} else {
// 			// Update existing record
// 			err := db.Model(&models.TotalSales{}).
// 				Where("organization_id = ? AND DATE(date) = ?", sale.OrganizationID, yesterday).
// 				Update("total_selling_price", sale.TotalSellingPrice).Error

// 			if err != nil {
// 				log.Printf("[ERROR] [%s] Failed to update total sales for OrgID %d on %s: %v\n",
// 					time.Now().Format("2006-01-02 15:04:05"), sale.OrganizationID, yesterday, err)
// 			} else {
// 				log.Printf("[INFO] [%s] Updated total sales for OrgID %d: Amount %.2f for %s\n",
// 					time.Now().Format("2006-01-02 15:04:05"), sale.OrganizationID, sale.TotalSellingPrice, yesterday)
// 			}
// 		}
// 	}

// 	log.Printf("[INFO] [%s] Daily sales summary process completed successfully for %s\n",
// 		time.Now().Format("2006-01-02 15:04:05"), yesterday)

// 	return nil
// }


func StartDailySalesSummary(db *gorm.DB) {
	log.Printf("[SCHEDULER_INIT] [%s] Starting daily sales summary scheduler\n", time.Now().Format("2006-01-02 15:04:05"))

	go func() {
		log.Printf("[GOROUTINE_START] [%s] Daily sales summary goroutine has started successfully\n", time.Now().Format("2006-01-02 15:04:05"))

		for {
			now := time.Now()
			nextRun := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
			sleepDuration := time.Until(nextRun)

			log.Printf("[SCHEDULE_INFO] [%s] Next daily sales summary scheduled for: %s (sleeping for %s)\n",
				now.Format("2006-01-02 15:04:05"), nextRun.Format("2006-01-02 15:04:05"), sleepDuration)

			time.Sleep(sleepDuration)

			execTime := time.Now()
			summaryDate := execTime.AddDate(0, 0, -1)

			log.Printf("[EXECUTING] [%s] Running sales summary for: %s\n",
				execTime.Format("2006-01-02 15:04:05"), summaryDate.Format("2006-01-02"))

			if err := SummarizeDailySales(db, summaryDate); err != nil {
				log.Printf("[ERROR] [%s] Summary failed: %v\n", execTime.Format("2006-01-02 15:04:05"), err)
			} else {
				log.Printf("[SUCCESS] [%s] Summary completed for: %s\n",
					execTime.Format("2006-01-02 15:04:05"), summaryDate.Format("2006-01-02"))
			}
		}
	}()

	log.Printf("[SCHEDULER_READY] [%s] Scheduler is now running\n", time.Now().Format("2006-01-02 15:04:05"))
}

func SummarizeDailySales(db *gorm.DB, summaryDate time.Time) error {
	startTime := time.Now()
	summaryDateStr := summaryDate.Format("2006-01-02")
	log.Printf("[SUMMARY_STARTED] [%s] Starting summary for: %s\n", startTime.Format("2006-01-02 15:04:05"), summaryDateStr)

	var salesData []struct {
		TotalSellingPrice float64
		OrganizationID    uint
	}

	startOfDay := summaryDate
	endOfDay := summaryDate.AddDate(0, 0, 1)

	err := db.Model(&models.Sale{}).
		Select("SUM(total_selling_price) as total_selling_price, organizations_id as organization_id").
		Where("date >= ? AND date < ?", startOfDay, endOfDay).
		Group("organizations_id").
		Scan(&salesData).Error

	if err != nil {
		log.Printf("[DB_ERROR] [%s] Failed to fetch sales: %v\n", time.Now().Format("2006-01-02 15:04:05"), err)
		return fmt.Errorf("fetch error: %w", err)
	}

	if len(salesData) == 0 {
		log.Printf("[NO_DATA] [%s] No sales found for: %s\n", time.Now().Format("2006-01-02 15:04:05"), summaryDateStr)
		return nil
	}

	for i, sale := range salesData {
		log.Printf("[PROCESSING] [%s] Org %d/%d - ID: %d\n",
			time.Now().Format("2006-01-02 15:04:05"), i+1, len(salesData), sale.OrganizationID)

		var existing models.TotalSales

		result := db.Where("organization_id = ? AND date >= ? AND date < ?", sale.OrganizationID, startOfDay, endOfDay).First(&existing)

		if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
			log.Printf("[DB_CHECK_ERR] [%s] Org %d: %v\n", time.Now().Format("2006-01-02 15:04:05"), sale.OrganizationID, result.Error)
			continue
		}

		if result.RowsAffected == 0 {
			log.Printf("[INSERT] [%s] New record for Org %d: %.2f\n",
				time.Now().Format("2006-01-02 15:04:05"), sale.OrganizationID, sale.TotalSellingPrice)

			newTotal := models.TotalSales{
				TotalSellingPrice: sale.TotalSellingPrice,
				OrganizationID:    sale.OrganizationID,
				Date:              summaryDate,
			}

			if err := db.Create(&newTotal).Error; err != nil {
				log.Printf("[INSERT_ERR] [%s] Org %d: %v\n", time.Now().Format("2006-01-02 15:04:05"), sale.OrganizationID, err)
			} else {
				log.Printf("[INSERT_OK] [%s] Org %d inserted successfully\n", time.Now().Format("2006-01-02 15:04:05"), sale.OrganizationID)
			}
		} else {
			log.Printf("[UPDATE] [%s] Updating Org %d: %.2f\n",
				time.Now().Format("2006-01-02 15:04:05"), sale.OrganizationID, sale.TotalSellingPrice)

			err := db.Model(&models.TotalSales{}).
				Where("organization_id = ? AND date >= ? AND date < ?", sale.OrganizationID, startOfDay, endOfDay).
				Update("total_selling_price", sale.TotalSellingPrice).Error

			if err != nil {
				log.Printf("[UPDATE_ERR] [%s] Org %d: %v\n", time.Now().Format("2006-01-02 15:04:05"), sale.OrganizationID, err)
			} else {
				log.Printf("[UPDATE_OK] [%s] Org %d updated successfully\n", time.Now().Format("2006-01-02 15:04:05"), sale.OrganizationID)
			}
		}
	}

	duration := time.Since(startTime)
	log.Printf("[SUMMARY_DONE] [%s] Summary for %s finished in %s\n",
		time.Now().Format("2006-01-02 15:04:05"), summaryDateStr, duration)

	return nil
}
