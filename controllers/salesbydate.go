package controllers

import (
	"log"
	"net/http"
	models "stock/models" // Adjust the import path as necessary
	"time"

	"github.com/labstack/echo/v4"
)

// GetSalesByDate fetches sales based on the provided date.
func GetSalesByDate(c echo.Context) error {
	log.Println("Received request to fetch sales by date")

	// Extract date from query parameters
	dateStr := c.QueryParam("date")
	if dateStr == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Date is required")
	}

	// Parse the date string
	date, err := time.Parse("2006-01-02", dateStr) // Format: YYYY-MM-DD
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD")
	}

	// Initialize database connection
	db := getDB()
	if db == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Query sales by date (only matching the date part)
	var sales []models.Sale
	if err := db.Where("DATE(date) = ?", date.Format("2006-01-02")).Find(&sales).Error; err != nil {
		log.Printf("Error querying sales by date: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	}

	// Log the number of sales fetched
	log.Printf("Fetched %d sales for date: %s", len(sales), date.Format("2006-01-02"))

	// Return the fetched sales as JSON
	return c.JSON(http.StatusOK, sales)
}
