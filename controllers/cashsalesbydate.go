package controllers

import (
	"log"
	"net/http"
	models "stock/models" // Adjust the import path as necessary
	"time"

	"github.com/labstack/echo/v4"
)

// DateRequest defines the structure for the JSON input
type DateRequest struct {
	Date string `json:"date"`
}

// GetSalesByDate fetches sales based on the provided date.
func GetSalesByDate(c echo.Context) error {
	log.Println("Received request to fetch sales by date")

	// Create an instance of DateRequest to hold the parsed JSON
	var request DateRequest

	// Bind the incoming JSON to the struct
	if err := c.Bind(&request); err != nil {
		log.Println("Failed to parse JSON:", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON")
	}

	// Validate the date
	if request.Date == "" {
		log.Println("Date is required")
		return echo.NewHTTPError(http.StatusBadRequest, "Date is required")
	}

	// Parse the date string
	date, err := time.Parse("2006-01-02", request.Date) // Format: YYYY-MM-DD
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD")
	}

	// Initialize database connection
	db := getDB()
	if db == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Query sales by date (only matching the date part) from the sales_by_cash table
	var sales []models.SalebyCash // Use your SalebyCash struct
	if err := db.Where("DATE(date) = ?", date.Format("2006-01-02")).Find(&sales).Error; err != nil {
		log.Printf("Error querying sales by date: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	}

	// Log the number of sales fetched
	log.Printf("Fetched %d sales for date: %s", len(sales), date.Format("2006-01-02"))

	// Return the fetched sales as JSON
	return c.JSON(http.StatusOK, sales)
}
