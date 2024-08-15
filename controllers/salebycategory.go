package controllers

import (
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	models "stock/models"
	"strconv"
)

// Get the database instance
//func getDB() *gorm.DB {
//db := database.GetDB()
//if db == nil {
//	return nil
//}
//return db
//}

// FetchSalesByCategory fetches sales data filtered by category name
func FetchSalesByCategory(c echo.Context) error {
	// Extract category_name from request parameters
	categoryName := c.Param("category_name")
	if categoryName == "" {
		log.Printf("No category name provided in the request")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Category name is required"})
	}
	log.Printf("Received request to fetch sales for category name: %s", categoryName)

	db := getDB()
	if db == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Query sales data from the sale table filtered by category name
	var sales []models.SaleByCategory
	if err := db.Where("category_name = ?", categoryName).Find(&sales).Error; err != nil {
		log.Printf("Error querying sales from database: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	}

	// Check if no sales were found
	if len(sales) == 0 {
		log.Printf("No sales found for category name %s", categoryName)
		return c.JSON(http.StatusNotFound, map[string]string{"error": "No sales found for this category"})
	}

	// Log the number of sales fetched
	log.Printf("Fetched %d sales for category name %s", len(sales), categoryName)

	// Return the fetched Sales as JSON
	return c.JSON(http.StatusOK, sales)
}

// FetchSalesByDate fetches sales data filtered by date
func FetchSalesByDate(c echo.Context) error {
	// Extract date from request parameters
	date := c.Param("date")
	if date == "" {
		log.Printf("No date provided in the request")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Date is required"})
	}
	log.Printf("Received request to fetch sales for date: %s", date)

	db := getDB()
	if db == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Query sales data from the sale table filtered by date
	var sales []models.SaleByCategory
	if err := db.Where("DATE(date) = ?", date).Find(&sales).Error; err != nil {
		log.Printf("Error querying sales from database: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	}

	// Check if no sales were found
	if len(sales) == 0 {
		log.Printf("No sales found for date %s", date)
		return c.JSON(http.StatusNotFound, map[string]string{"error": "No sales found for this date"})
	}

	// Log the number of sales fetched
	log.Printf("Fetched %d sales for date %s", len(sales), date)

	// Return the fetched Sales as JSON
	return c.JSON(http.StatusOK, sales)
}

// FetchSalesByUserID fetches sales data filtered by user ID
func FetchSalesByUserID(c echo.Context) error {
	// Extract user_id from request parameters
	userIDStr := c.Param("user_id")
	if userIDStr == "" {
		log.Printf("No user ID provided in the request")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "User ID is required"})
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		log.Printf("Invalid user ID: %s", userIDStr)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID")
	}

	log.Printf("Received request to fetch sales for user ID: %d", userID)

	db := getDB()
	if db == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Query sales data from the sale table filtered by user ID
	var sales []models.SaleByCategory
	if err := db.Where("user_id = ?", userID).Find(&sales).Error; err != nil {
		log.Printf("Error querying sales from database: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	}

	// Check if no sales were found
	if len(sales) == 0 {
		log.Printf("No sales found for user ID %d", userID)
		return c.JSON(http.StatusNotFound, map[string]string{"error": "No sales found for this user"})
	}

	// Log the number of sales fetched
	log.Printf("Fetched %d sales for user ID %d", len(sales), userID)

	// Return the fetched Sales as JSON
	return c.JSON(http.StatusOK, sales)
}
