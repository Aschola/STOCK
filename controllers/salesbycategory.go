package controllers

import (
	"log"
	"net/http"
	models "stock/models"

	"github.com/labstack/echo/v4"
)

// GetSalesByCategory fetches sales based on the provided category name.
func GetSalesByCategory(c echo.Context) error {
	log.Println("Received request to fetch sales by category")

	// Extract category name from query parameters
	categoryName := c.QueryParam("category_name")
	if categoryName == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Category name is required")
	}

	// Initialize database connection
	db := getDB()
	if db == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Query sales by category name
	var sales []models.Sale
	if err := db.Where("category_name = ?", categoryName).Find(&sales).Error; err != nil {
		log.Printf("Error querying sales by category: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	}

	// Log the number of sales fetched
	log.Printf("Fetched %d sales for category: %s", len(sales), categoryName)

	// Return the fetched sales as JSON
	return c.JSON(http.StatusOK, sales)
}
