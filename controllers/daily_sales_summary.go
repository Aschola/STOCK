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
	// Log when the function is first called
	log.Printf("[SCHEDULER_INIT] [%s] Starting daily sales summary scheduler\n", time.Now().Format("2006-01-02 15:04:05"))

	go func() {
		// Log when the goroutine starts
		log.Printf("[GOROUTINE_START] [%s] Daily sales summary goroutine has started successfully\n", time.Now().Format("2006-01-02 15:04:05"))

		for {
			// Use current time for scheduling calculations
			now := time.Now()
			log.Printf("[SCHEDULE_CALC] [%s] Calculating next midnight run time\n", now.Format("2006-01-02 15:04:05"))

			// Schedule the next execution at midnight
			nextRun := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
			
			sleepDuration := time.Until(nextRun)
			log.Printf("[SCHEDULE_INFO] [%s] Next daily sales summary scheduled for: %s (sleeping for %s)\n", 
				now.Format("2006-01-02 15:04:05"), 
				nextRun.Format("2006-01-02 15:04:05"),
				sleepDuration.String())

			// Log every hour until midnight to confirm the goroutine is still alive
			remainingHours := int(sleepDuration.Hours()) + 1
			for i := 0; i < remainingHours; i++ {
				// Calculate how long to sleep for this interval
				// Either sleep for 1 hour, or the remaining time if less than 1 hour
				var intervalSleep time.Duration
				if sleepDuration > time.Hour {
					intervalSleep = time.Hour
				} else {
					intervalSleep = sleepDuration
				}
				
				// Sleep for the calculated interval
				time.Sleep(intervalSleep)
				
				// Update remaining sleep duration
				sleepDuration = time.Until(nextRun)
				
				// Log that we're still waiting
				if sleepDuration > 0 {
					log.Printf("[WAITING] [%s] Still waiting for scheduled time. %s remaining until execution at %s\n", 
						time.Now().Format("2006-01-02 15:04:05"), 
						sleepDuration.String(),
						nextRun.Format("2006-01-02 15:04:05"))
				}
			}

			// After the sleep loop, we should be very close to midnight
			remainingSeconds := time.Until(nextRun)
			if remainingSeconds > 0 {
				log.Printf("[FINAL_WAIT] [%s] Final wait of %s before execution\n", 
					time.Now().Format("2006-01-02 15:04:05"), 
					remainingSeconds.String())
				time.Sleep(remainingSeconds)
			}

			// Log that we're about to execute
			execTime := time.Now()
			log.Printf("[EXECUTING] [%s] Time to execute daily sales summary\n", execTime.Format("2006-01-02 15:04:05"))

			// Calculate the date to summarize (previous day)
			summaryDate := execTime.AddDate(0, 0, -1)
			log.Printf("[DATE_INFO] [%s] Will summarize sales for date: %s\n", 
				execTime.Format("2006-01-02 15:04:05"),
				summaryDate.Format("2006-01-02"))

			// Execute the summary process for the day that just ended
			if err := SummarizeDailySales(db, summaryDate); err != nil {
				log.Printf("[ERROR] [%s] Failed to summarize daily sales: %v\n", 
					time.Now().Format("2006-01-02 15:04:05"), err)
			} else {
				log.Printf("[SUCCESS] [%s] Daily sales summary completed successfully\n", 
					time.Now().Format("2006-01-02 15:04:05"))
			}
			
			// Log that we're starting the next cycle
			log.Printf("[CYCLE_COMPLETE] [%s] Daily summary cycle complete, starting next cycle\n", 
				time.Now().Format("2006-01-02 15:04:05"))
		}
	}()

	// Log that the scheduler has been initiated
	log.Printf("[SCHEDULER_READY] [%s] Daily sales summary scheduler initiated successfully\n", 
		time.Now().Format("2006-01-02 15:04:05"))
}

// SummarizeDailySales aggregates and updates daily sales for each organization.
func SummarizeDailySales(db *gorm.DB, summaryDate time.Time) error {
	startTime := time.Now()
	summaryDateStr := summaryDate.Format("2006-01-02")
	
	log.Printf("[SUMMARY_STARTED] [%s] Starting sales summary for date: %s\n", 
		startTime.Format("2006-01-02 15:04:05"), summaryDateStr)

	var salesData []struct {
		TotalSellingPrice float64
		OrganizationID    uint
	}

	// Log that we're about to query the database
	log.Printf("[DB_QUERY] [%s] Querying database for sales on date: %s\n", 
		time.Now().Format("2006-01-02 15:04:05"), summaryDateStr)

	// Fetch total selling price per organization for the specified date
	err := db.Model(&models.Sale{}).
		Select("SUM(total_selling_price) as total_selling_price, organizations_id as organization_id").
		Where("DATE(date) = ?", summaryDateStr).
		Group("organizations_id").
		Scan(&salesData).Error

	// Error handling for query execution
	if err != nil {
		log.Printf("[DB_ERROR] [%s] Error retrieving sales data for %s: %v\n", 
			time.Now().Format("2006-01-02 15:04:05"), summaryDateStr, err)
		return fmt.Errorf("error calculating total sales: %w", err)
	}

	// Log the result of the query
	log.Printf("[DB_RESULT] [%s] Found %d organization sales records for date: %s\n", 
		time.Now().Format("2006-01-02 15:04:05"), len(salesData), summaryDateStr)

	// No sales data found, log and exit gracefully
	if len(salesData) == 0 {
		log.Printf("[NO_DATA] [%s] No sales data found for %s. Skipping summary.\n", 
			time.Now().Format("2006-01-02 15:04:05"), summaryDateStr)
		return nil
	}

	// Process sales data for each organization
	for i, sale := range salesData {
		log.Printf("[PROCESSING] [%s] Processing organization %d of %d (OrgID: %d)\n", 
			time.Now().Format("2006-01-02 15:04:05"), i+1, len(salesData), sale.OrganizationID)
		
		var existingRecord models.TotalSales

		// Check if a record already exists for this organization for this date
		log.Printf("[DB_CHECK] [%s] Checking for existing record for OrgID %d on %s\n", 
			time.Now().Format("2006-01-02 15:04:05"), sale.OrganizationID, summaryDateStr)
			
		result := db.Where("organization_id = ? AND DATE(date) = ?", sale.OrganizationID, summaryDateStr).
			First(&existingRecord)

		if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
			// Log unexpected database errors
			log.Printf("[DB_ERROR] [%s] Database error checking existing record for OrgID %d on %s: %v\n",
				time.Now().Format("2006-01-02 15:04:05"), sale.OrganizationID, summaryDateStr, result.Error)
			continue // Skip this iteration to avoid further failures
		}

		if result.RowsAffected == 0 {
			// Insert new record - use the summary date for the record
			log.Printf("[INSERT_PREP] [%s] Preparing to insert new record for OrgID %d with amount %.2f for date %s\n",
				time.Now().Format("2006-01-02 15:04:05"), sale.OrganizationID, sale.TotalSellingPrice, summaryDateStr)
				
			totalSale := models.TotalSales{
				TotalSellingPrice: sale.TotalSellingPrice,
				OrganizationID:    sale.OrganizationID,
				Date:              summaryDate, // This is the date passed in (previous day)
			}

			if err := db.Create(&totalSale).Error; err != nil {
				log.Printf("[INSERT_ERROR] [%s] Failed to insert total sales for OrgID %d on %s: %v\n",
					time.Now().Format("2006-01-02 15:04:05"), sale.OrganizationID, summaryDateStr, err)
			} else {
				log.Printf("[INSERT_SUCCESS] [%s] Inserted total sales for OrgID %d: Amount %.2f for %s\n",
					time.Now().Format("2006-01-02 15:04:05"), sale.OrganizationID, sale.TotalSellingPrice, summaryDateStr)
			}
		} else {
			// Update existing record
			log.Printf("[UPDATE_PREP] [%s] Preparing to update existing record for OrgID %d with new amount %.2f for date %s\n",
				time.Now().Format("2006-01-02 15:04:05"), sale.OrganizationID, sale.TotalSellingPrice, summaryDateStr)
				
			err := db.Model(&models.TotalSales{}).
				Where("organization_id = ? AND DATE(date) = ?", sale.OrganizationID, summaryDateStr).
				Update("total_selling_price", sale.TotalSellingPrice).Error

			if err != nil {
				log.Printf("[UPDATE_ERROR] [%s] Failed to update total sales for OrgID %d on %s: %v\n",
					time.Now().Format("2006-01-02 15:04:05"), sale.OrganizationID, summaryDateStr, err)
			} else {
				log.Printf("[UPDATE_SUCCESS] [%s] Updated total sales for OrgID %d: Amount %.2f for %s\n",
					time.Now().Format("2006-01-02 15:04:05"), sale.OrganizationID, sale.TotalSellingPrice, summaryDateStr)
			}
		}
	}

	endTime := time.Now()
	duration := endTime.Sub(startTime)
	
	log.Printf("[SUMMARY_COMPLETE] [%s] Daily sales summary process completed successfully for %s (took %s)\n",
		endTime.Format("2006-01-02 15:04:05"), summaryDateStr, duration.String())

	return nil
}