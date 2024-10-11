package controllers

import (
	"log"
	"net/http"
	models "stock/models"

	"github.com/labstack/echo/v4"
)

// SalesRequest defines the structure for the JSON input
type SalesRequestt struct {
	CategoryName string `json:"category_name"`
}

// GetSalesBySTKPUSH fetches sales based on the provided category name from the sales_by_STKPUSH table.
func GetSalesBySTKPUSH(c echo.Context) error {
	log.Println("Received request to fetch sales by category")

	// Create an instance of SalesRequest to hold the parsed JSON
	var request SalesRequest

	// Bind the incoming JSON to the struct
	if err := c.Bind(&request); err != nil {
		log.Println("Failed to parse JSON:", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON")
	}

	// Validate the category name
	if request.CategoryName == "" {
		log.Println("Category name is required")
		return echo.NewHTTPError(http.StatusBadRequest, "Category name is required")
	}
	log.Printf("Category name received: %s", request.CategoryName)

	// Initialize database connection
	db := getDB()
	if db == nil {
		log.Println("Failed to connect to the database")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
	}
	log.Println("Database connection established")

	// Query sales by category name from the sales_by_STKPUSH table
	var sales []models.SalebyCash // Ensure this matches your model
	log.Printf("Querying sales for category: %s", request.CategoryName)
	if err := db.Table("sales_by_STKPUSH").Where("category_name = ?", request.CategoryName).Find(&sales).Error; err != nil {
		log.Printf("Error querying sales by category: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	}

	// Log the number of sales fetched
	log.Printf("Fetched %d sales for category: %s", len(sales), request.CategoryName)

	// Return the fetched sales as JSON
	log.Println("Returning fetched sales as JSON")
	return c.JSON(http.StatusOK, sales)
}
