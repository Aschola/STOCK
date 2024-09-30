package controllers

import (
	"log"
	"net/http"
	models "stock/models" // Adjust the import path as necessary

	"github.com/labstack/echo/v4"
)

// GetSalesByUserID fetches sales based on the provided user ID.
func GetSalesByUserID(c echo.Context) error {
	log.Println("Received request to fetch sales by user ID")

	// Extract user ID from query parameters
	userID := c.QueryParam("user_id")
	if userID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "User ID is required")
	}

	// Initialize database connection
	db := getDB()
	if db == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Query sales by user ID
	var sales []models.Sale
	if err := db.Where("user_id = ?", userID).Find(&sales).Error; err != nil {
		log.Printf("Error querying sales by user ID: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	}

	// Log the number of sales fetched
	log.Printf("Fetched %d sales for user ID: %s", len(sales), userID)

	// Return the fetched sales as JSON
	return c.JSON(http.StatusOK, sales)
}
