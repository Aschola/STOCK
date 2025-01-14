package controllers

import (
	"log"
	"net/http"
	"stock/models"
	"time"

	"strconv"

	"github.com/labstack/echo/v4"
)

// GetAllSales retrieves all sales records from the sales_by_cash table
func GetAllSales(c echo.Context) error {
	log.Println("[INFO] Received request to fetch all sales records.")

	// Log request parameters if any (you can add more detailed logs if necessary)
	log.Printf("[INFO] Request received with parameters: %v", c.QueryParams())

	// Database connection
	db := getDB()
	if db == nil {
		logError("Database connection failed", nil)
		log.Println("[ERROR] Failed to connect to the database")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
	}
	log.Println("[INFO] Successfully connected to the database.")

	// Retrieve all sales records
	var sales []models.Sale
	log.Println("[INFO] Attempting to fetch all sales records from the database.")
	if err := db.Find(&sales).Error; err != nil {
		logError("Error fetching all sales records", err)
		log.Printf("[ERROR] Error fetching sales records: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching sales records")
	}
	log.Printf("[INFO] Successfully fetched %d sales records from the database.", len(sales))

	// Check if no sales were found
	if len(sales) == 0 {
		log.Println("[INFO] No sales records found in the database.")
		return echo.NewHTTPError(http.StatusNotFound, "No sales records found")
	}

	log.Printf("[INFO] Fetched %d sales records from the database.", len(sales))

	// Prepare the response
	var response []map[string]interface{}
	log.Println("[INFO] Formatting sales records for the response.")
	for _, sale := range sales {
		saleData := map[string]interface{}{
			"sale_id":             sale.SaleID,
			"product_name":        sale.Name,
			"unit_buying_price":   sale.UnitBuyingPrice,
			"total_buying_price":  sale.TotalBuyingPrice,
			"unit_selling_price":  sale.UnitSellingPrice,
			"total_selling_price": sale.TotalSellingPrice,
			"profit":              sale.Profit,
			"quantity":            sale.Quantity,
			"cash_receive":        sale.CashReceived,
			"balance":             sale.Balance,
			"user_id":             sale.UserID,
			"date":                sale.Date.Format("2006-01-02T15:04:05Z"), // ISO 8601 format
			"category_name":       sale.CategoryName,
		}
		response = append(response, saleData)
	}

	// Log how many records will be returned in the response
	log.Printf("[INFO] Returning %d sales records in the response.", len(response))

	// Return the formatted response
	return c.JSON(http.StatusOK, response)
}

func GetSalesBySaleID(c echo.Context) error {
	// Extract sale_id from the URL parameter using :sale_id format
	saleID := c.Param("sale_id")
	// Parse the sale_id from the URL into an integer
	parsedSaleID, err := strconv.ParseInt(saleID, 10, 64) // Parse the sale_id as an int64
	if err != nil {
		log.Printf("Failed to parse sale_id: %v", err)
		return errorResponse(c, http.StatusBadRequest, "Invalid sale_id")
	}

	// Fetch all sales with the matching sale_id
	var sales []models.Sale
	db := getDB()
	if db == nil {
		log.Println("Failed to get database instance while fetching sales")
		return errorResponse(c, http.StatusInternalServerError, "Failed to connect to the database")
	}

	if err := db.Table("sales_by_cash").Where("sale_id = ?", parsedSaleID).Find(&sales).Error; err != nil {
		log.Printf("Failed to fetch sales for sale_id: %d, error: %v", parsedSaleID, err)
		return errorResponse(c, http.StatusInternalServerError, "Failed to fetch sales")
	}

	// Return the list of sales
	log.Printf("Fetched %d sales for sale_id: %d", len(sales), parsedSaleID)
	return c.JSON(http.StatusOK, sales)
}

// GetSalesByDate retrieves all sales for a specific date from the sales_by_cash table
func GetSalesByDate(c echo.Context) error {
	log.Println("[INFO] Received request to fetch sales for a specific date.")

	// Get the date parameter from the URL
	dateParam := c.Param("date")

	// Try to parse the date
	saleDate, err := time.Parse("2006-01-02", dateParam) // Format: "YYYY-MM-DD"
	if err != nil {
		logError("Error parsing date", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD")
	}

	// Database connection
	db := getDB()
	if db == nil {
		logError("Database connection failed", nil)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Retrieve sales for the specific date
	var sales []models.Sale
	if err := db.Where("DATE(date) = ?", saleDate.Format("2006-01-02")).Find(&sales).Error; err != nil {
		logError("Error fetching sales for date", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching sales for the date")
	}

	// Check if no sales were found
	if len(sales) == 0 {
		log.Printf("[INFO] No sales found for date: %v", saleDate)
		return echo.NewHTTPError(http.StatusNotFound, "No sales found for this date")
	}

	log.Printf("[INFO] Fetched %d sales records for date: %v", len(sales), saleDate)

	// Prepare the response
	var response []map[string]interface{}
	for _, sale := range sales {
		saleData := map[string]interface{}{
			"sale_id":             sale.SaleID,
			"product_name":        sale.Name,
			"quantity":            sale.Quantity,
			"unit_buying_price":   sale.UnitBuyingPrice,
			"total_buying_price":  sale.TotalBuyingPrice,
			"unit_selling_price":  sale.UnitSellingPrice,
			"user_id":             sale.UserID,
			"date":                sale.Date.Format("2006-01-02T15:04:05Z"), // ISO 8601 format
			"category_name":       sale.CategoryName,
			"total_selling_price": sale.TotalSellingPrice,
			"profit":              sale.Profit,
			"cash_receive":        sale.CashReceived,
			"balance":             sale.Balance,
		}
		response = append(response, saleData)
	}

	// Return the formatted response
	return c.JSON(http.StatusOK, response)
}

// GetSalesByUser retrieves all sales for a specific user_id from the sales_by_cash table
func GetSalesByUser(c echo.Context) error {
	log.Println("[INFO] Received request to fetch sales for a specific user_id.")

	// Get the user_id from the URL parameter
	userID := c.Param("user_id")

	// Database connection
	db := getDB()
	if db == nil {
		logError("Database connection failed", nil)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Retrieve sales for the specific user_id
	var sales []models.Sale
	if err := db.Where("user_id = ?", userID).Find(&sales).Error; err != nil {
		logError("Error fetching sales for user_id", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching sales for the user")
	}

	// Check if no sales found
	if len(sales) == 0 {
		log.Printf("[INFO] No sales found for user_id: %s", userID)
		return echo.NewHTTPError(http.StatusNotFound, "No sales found for this user")
	}

	log.Printf("[INFO] Fetched %d sales records for user_id: %s", len(sales), userID)

	// Prepare the response
	var response []map[string]interface{}
	for _, sale := range sales {
		saleData := map[string]interface{}{
			"sale_id":             sale.SaleID,
			"product_name":        sale.Name,
			"quantity":            sale.Quantity,
			"unit_buying_price":   sale.UnitBuyingPrice,
			"total_buying_price":  sale.TotalBuyingPrice,
			"unit_selling_price":  sale.UnitSellingPrice,
			"user_id":             sale.UserID,
			"date":                sale.Date.Format("2006-01-02T15:04:05Z"), // ISO 8601 format
			"category_name":       sale.CategoryName,
			"total_selling_price": sale.TotalSellingPrice,
			"profit":              sale.Profit,
			"cash_receive":        sale.CashReceived,
			"balance":             sale.Balance,
		}
		response = append(response, saleData)
	}

	// Return the formatted response
	return c.JSON(http.StatusOK, response)
}
