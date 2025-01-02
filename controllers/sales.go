package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"stock/models"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// SellProduct handles the sale of a product and updates the stock
func SellProduct(c echo.Context) error {
	// Log the incoming request
	log.Println("Received a request to sell a product.")

	// Create an instance of SalePayload to hold the incoming data
	var payload models.SalePayload

	// Decode the incoming JSON payload into the struct
	if err := json.NewDecoder(c.Request().Body).Decode(&payload); err != nil {
		log.Printf("Error parsing request body: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input data")
	}

	// Log the parsed data for better debugging
	log.Printf("Parsed payload: %+v", payload)

	// Retrieve the product_id, quantity_sold, and other fields from the payload
	productID := payload.ProductID
	quantitySold := payload.QuantitySold
	userID := payload.UserID
	cashReceived := payload.CashReceived

	// Get database instance
	db := getDB()
	if db == nil {
		log.Println("Database connection failed.")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Start a transaction
	tx := db.Begin()
	defer tx.Rollback()

	// Retrieve the Stock_pas entry for the product
	var Stock_pas models.Stock_pas
	if err := tx.First(&Stock_pas, "product_id = ?", productID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("Stock_pas not found for product_id: %d", productID)
			return echo.NewHTTPError(http.StatusNotFound, "Stock_pas not found for this product")
		}
		log.Printf("Error fetching Stock_pas details for product_id: %d, error: %v", productID, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching Stock_pas details")
	}

	// Fetch the product details (including name and category_name)
	var product models.Product
	if err := tx.First(&product, "product_id = ?", productID).Error; err != nil {
		log.Printf("Product not found for product_id: %d, error: %v", productID, err)
		return echo.NewHTTPError(http.StatusNotFound, "Product not found")
	}

	// Log the fetched product details
	log.Printf("Fetched product details: %+v", product)

	// Check Stock_pas availability
	if Stock_pas.Quantity < quantitySold {
		log.Printf("Insufficient Stock_pas for product_id: %d. Available: %d, Requested: %d", productID, Stock_pas.Quantity, quantitySold)
		return echo.NewHTTPError(http.StatusBadRequest, "Insufficient Stock_pas")
	}

	// Update Stock_pas quantity
	Stock_pas.Quantity -= quantitySold
	Stock_pas.UpdatedAt = time.Now() // Update the updated_at timestamp
	if err := tx.Save(&Stock_pas).Error; err != nil {
		log.Printf("Error updating Stock_pas for product_id: %d, error: %v", productID, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error updating Stock_pas")
	}

	// Log Stock_pas update
	log.Printf("Updated Stock_pas for product_id: %d, new quantity: %d", productID, Stock_pas.Quantity)

	// Calculate the total buying price, selling price, and profit
	totalCost := float64(quantitySold) * Stock_pas.BuyingPrice
	totalSellingPrice := float64(quantitySold) * Stock_pas.SellingPrice
	profit := totalSellingPrice - totalCost
	balance := cashReceived - totalSellingPrice

	// Log the calculated values
	log.Printf("Sale calculations - Total Cost: %f, Total Selling Price: %f, Profit: %f, Balance: %f", totalCost, totalSellingPrice, profit, balance)

	// Record the sale in the sales_by_cash table
	sale := models.Sale{
		Name:              product.Name,          // Dynamically fetched product name
		UnitBuyingPrice:   Stock_pas.BuyingPrice, // Assuming this is part of your Stock_pas model
		TotalBuyingPrice:  totalCost,
		UnitSellingPrice:  Stock_pas.SellingPrice,
		TotalSellingPrice: totalSellingPrice,
		Profit:            profit,
		Quantity:          quantitySold,
		CashReceive:       cashReceived,
		Balance:           balance,
		UserID:            userID,
		Date:              time.Now(),
		CategoryName:      product.Category, // Dynamically fetched category name
	}

	if err := tx.Create(&sale).Error; err != nil {
		log.Printf("Error recording sale for product_id: %d, error: %v", productID, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error recording sale")
	}

	// Log the sale record
	log.Printf("Sale recorded successfully: %+v", sale)

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("Error committing transaction for product_id: %d, error: %v", productID, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error committing transaction")
	}

	// Log successful transaction commit
	log.Printf("Transaction committed successfully for product_id: %d", productID)

	// Return a success response with the updated structure
	return c.JSON(http.StatusOK, map[string]interface{}{
		"balance":             balance,
		"message":             "Cash sale processed successfully",
		"product_id":          productID,
		"quantity_sold":       quantitySold,
		"remaining_qty":       Stock_pas.Quantity,
		"total_cost":          totalCost,
		"total_selling_price": totalSellingPrice,
		"user_id":             userID,
		"category_name":       product.Category, // Return dynamically fetched category name
		"product_name":        product.Name,     // Return dynamically fetched product name
	})
}

// GetAllSales retrieves all sales from the sales_by_cash table
func GetAllSales(c echo.Context) error {
	// Log the incoming request
	log.Println("Received request to fetch all sales.")

	// Get database instance
	db := getDB()
	if db == nil {
		log.Println("Database connection failed.")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Retrieve all sales from the sales_by_cash table
	var sales []models.Sale
	if err := db.Find(&sales).Error; err != nil {
		log.Printf("Error fetching sales: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching sales")
	}

	// Log the number of sales fetched
	log.Printf("Fetched %d sales records.", len(sales))

	// Create a slice of formatted sale data to return
	var response []map[string]interface{}
	for _, sale := range sales {
		saleData := map[string]interface{}{
			"sale_id":             sale.SaleID,
			"name":                sale.Name,
			"quantity":            sale.Quantity,
			"unit_buying_price":   sale.UnitBuyingPrice,
			"total_buying_price":  sale.TotalBuyingPrice,
			"unit_selling_price":  sale.UnitSellingPrice,
			"user_id":             sale.UserID,
			"date":                sale.Date.Format("2006-01-02T15:04:05Z"), // Format date as required
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

// GetSalesByUser retrieves sales for a specific user from the sales_by_cash table
func GetSalesByUser(c echo.Context) error {
	// Get the user_id from the URL
	userID := c.Param("user_id")

	// Log the incoming request for this user
	log.Printf("Received request to fetch sales for user_id: %s", userID)

	// Get database instance
	db := getDB()
	if db == nil {
		log.Println("Database connection failed.")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Retrieve sales for the specific user from the sales_by_cash table
	var sales []models.Sale
	if err := db.Where("user_id = ?", userID).Find(&sales).Error; err != nil {
		log.Printf("Error fetching sales for user_id: %s, error: %v", userID, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching sales")
	}

	// Check if no sales found
	if len(sales) == 0 {
		log.Printf("No sales found for user_id: %s", userID)
		return echo.NewHTTPError(http.StatusNotFound, "No sales found for this user")
	}

	// Log the number of sales fetched
	log.Printf("Fetched %d sales records for user_id: %s", len(sales), userID)

	// Create a slice of formatted sale data to return
	var response []map[string]interface{}
	for _, sale := range sales {
		saleData := map[string]interface{}{
			"sale_id":             sale.SaleID,
			"name":                sale.Name,
			"quantity":            sale.Quantity,
			"unit_buying_price":   sale.UnitBuyingPrice,
			"total_buying_price":  sale.TotalBuyingPrice,
			"unit_selling_price":  sale.UnitSellingPrice,
			"user_id":             sale.UserID,
			"date":                sale.Date.Format("2006-01-02T15:04:05-07:00"), // Date format with timezone
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

// GetSalesByDate retrieves all sales for a specific date from the sales_by_cash table
func GetSalesByDate(c echo.Context) error {
	// Log the incoming request
	log.Println("Received request to fetch sales for a specific date.")

	// Get the date from the URL parameter (e.g., /sales/date/2023-12-20)
	dateParam := c.Param("date")

	// Try to parse the date from the URL parameter
	saleDate, err := time.Parse("2006-01-02", dateParam) // Format: "YYYY-MM-DD"
	if err != nil {
		log.Printf("Error parsing date: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD")
	}

	// Get database instance
	db := getDB()
	if db == nil {
		log.Println("Database connection failed.")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Retrieve sales for the specific date from the sales_by_cash table
	var sales []models.Sale
	if err := db.Where("DATE(date) = ?", saleDate.Format("2006-01-02")).Find(&sales).Error; err != nil {
		log.Printf("Error fetching sales for date %v: %v", saleDate, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching sales for the date")
	}

	// Check if no sales are found for the given date
	if len(sales) == 0 {
		log.Printf("No sales found for date %v", saleDate)
		return echo.NewHTTPError(http.StatusNotFound, "No sales found for this date")
	}

	// Log the number of sales fetched
	log.Printf("Fetched %d sales records for date %v", len(sales), saleDate)

	// Create a slice of formatted sale data to return
	var response []map[string]interface{}
	for _, sale := range sales {
		saleData := map[string]interface{}{
			"sale_id":             sale.SaleID,
			"name":                sale.Name,
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
