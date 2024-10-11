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

	// Create an anonymous struct to hold the request body
	var request struct {
		UserID string `json:"user_id"`
	}

	// Bind the incoming JSON to the struct
	if err := c.Bind(&request); err != nil {
		log.Println("Failed to parse JSON:", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON")
	}

	// Validate the user ID
	if request.UserID == "" {
		log.Println("User ID is required")
		return echo.NewHTTPError(http.StatusBadRequest, "User ID is required")
	}
	log.Printf("User ID received: %s", request.UserID)

	// Initialize database connection
	db := getDB()
	if db == nil {
		log.Println("Failed to connect to the database")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Query sales by user ID
	var sales []models.SalebyCash
	if err := db.Where("user_id = ?", request.UserID).Find(&sales).Error; err != nil {
		log.Printf("Error querying sales by user ID: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	}

	// Log the number of sales fetched
	log.Printf("Fetched %d sales for user ID: %s", len(sales), request.UserID)

	// Return the fetched sales as JSON
	return c.JSON(http.StatusOK, sales)
}
