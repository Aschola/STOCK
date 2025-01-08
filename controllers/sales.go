package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"stock/models"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Helper function to log errors
func logError(message string, err error) {
	if err != nil {
		log.Printf("[ERROR] %s: %v", message, err)
	}
}

// SellProduct processes a sale and updates stock
func SellProduct(c echo.Context) error {
	log.Println("[INFO] Received request to sell a product.")

	var payload models.SalePayload

	// Decode the incoming JSON payload
	if err := json.NewDecoder(c.Request().Body).Decode(&payload); err != nil {
		logError("Error parsing request body", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input data")
	}

	log.Printf("[INFO] Parsed payload: %+v", payload)

	// Retrieve parameters from payload
	productID := payload.ProductID
	quantitySold := payload.QuantitySold
	userID := payload.UserID
	cashReceived := payload.CashReceived

	// Database connection
	db := getDB()
	if db == nil {
		logError("Database connection failed", nil)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Start a database transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("[ERROR] Unexpected error: %v", r)
		}
	}()
	defer tx.Rollback() // Rollback in case of any failure

	// Retrieve stock information
	var stock models.Stock
	if err := tx.First(&stock, "product_id = ?", productID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logError("Stock not found for product_id", err)
			return echo.NewHTTPError(http.StatusNotFound, "Stock not found for this product")
		}
		logError("Error fetching stock details", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching stock details")
	}

	// Check stock availability
	if stock.Quantity < quantitySold {
		log.Printf("[INFO] Insufficient stock for product_id: %d. Available: %d, Requested: %d", productID, stock.Quantity, quantitySold)
		return echo.NewHTTPError(http.StatusBadRequest, "Insufficient stock")
	}

	// Check if cash received is sufficient
	totalSellingPrice := float64(quantitySold) * stock.SellingPrice
	if cashReceived < totalSellingPrice {
		log.Printf("[INFO] Insufficient cash received for product_id: %d. Received: %f, Required: %f", productID, cashReceived, totalSellingPrice)
		return echo.NewHTTPError(http.StatusBadRequest, "Insufficient cash received")
	}

	// Update stock quantity
	stock.Quantity -= quantitySold
	stock.CreatedAt = time.Now() // Update timestamp
	if err := tx.Save(&stock).Error; err != nil {
		logError("Error updating stock", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error updating stock")
	}

	log.Printf("[INFO] Updated stock for product_id: %d, new quantity: %d", productID, stock.Quantity)

	// Calculate totals and profit
	totalCost := float64(quantitySold) * stock.BuyingPrice
	profit := totalSellingPrice - totalCost
	balance := cashReceived - totalSellingPrice

	log.Printf("[INFO] Sale calculations - Total Cost: %f, Total Selling Price: %f, Profit: %f, Balance: %f", totalCost, totalSellingPrice, profit, balance)

	// Record the sale in the database
	sale := models.Sale{
		Name:              "Product " + strconv.Itoa(int(stock.ProductID)), // Placeholder for Name
		UnitBuyingPrice:   stock.BuyingPrice,
		TotalBuyingPrice:  totalCost,
		UnitSellingPrice:  stock.SellingPrice,
		TotalSellingPrice: totalSellingPrice,
		Profit:            profit,
		Quantity:          quantitySold,
		CashReceive:       cashReceived,
		Balance:           balance,
		UserID:            userID,
		Date:              time.Now(),
		CategoryName:      "General", // You can modify or omit this field
	}

	if err := tx.Create(&sale).Error; err != nil {
		logError("Error recording sale", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error recording sale")
	}

	log.Printf("[INFO] Sale recorded successfully: %+v", sale)

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		logError("Error committing transaction", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error committing transaction")
	}

	log.Printf("[INFO] Transaction committed successfully for product_id: %d", productID)

	// Return success response
	return c.JSON(http.StatusOK, map[string]interface{}{
		"balance":             balance,
		"message":             "Cash sale processed successfully",
		"product_id":          productID,
		"quantity_sold":       quantitySold,
		"remaining_qty":       stock.Quantity,
		"total_cost":          totalCost,
		"total_selling_price": totalSellingPrice,
		"user_id":             userID,
		"category_name":       "General",                                       // Or another category
		"product_name":        "Product " + strconv.Itoa(int(stock.ProductID)), // Placeholder for Name
	})
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
			"cash_receive":        sale.CashReceive,
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
			"cash_receive":        sale.CashReceive,
			"balance":             sale.Balance,
		}
		response = append(response, saleData)
	}

	// Return the formatted response
	return c.JSON(http.StatusOK, response)
}
