package controllers

import (
	"log"
	"net/http"
	"stock/models"
	"time"

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

// Define the struct outside the function to reuse it later
type Product struct {
	ProductName       string  `json:"name" gorm:"column:name"`
	UnitBuyingPrice   float64 `json:"unit_buying_price"`
	TotalBuyingPrice  float64 `json:"total_buying_price"`
	UnitSellingPrice  float64 `json:"unit_selling_price"`
	TotalSellingPrice float64 `json:"total_selling_price"`
	Profit            float64 `json:"profit"`
	Quantity          int     `json:"quantity"`
	CategoryName      string  `json:"category_name"`
}

func GetSalesBySaleID(c echo.Context) error {
	// Get sale_id from the URL parameter or request
	saleID := c.Param("sale_id")

	// Database connection
	db := getDB()
	if db == nil {
		log.Println("Failed to get database instance")
		return errorResponse(c, http.StatusInternalServerError, "Failed to connect to the database")
	}

	// First, retrieve common details: cash_received, user_id, etc.
	var saleDetails struct {
		SaleID       int64   `json:"sale_id"`
		CashReceived float64 `json:"cash_received"`
		UserID       int64   `json:"user_id"`
		Date         string  `json:"date"`
	}
	if err := db.Table("sales_by_cash").
		Where("sale_id = ?", saleID).
		First(&saleDetails).Error; err != nil {
		log.Printf("Error retrieving sale details: %v", err)
		return errorResponse(c, http.StatusNotFound, "Sale details not found")
	}

	// Now, get the products for that sale
	var products []Product

	// Log the query being executed
	query := db.Table("sales_by_cash").
		Where("sale_id = ?", saleID).
		Select("name, unit_buying_price, total_buying_price, unit_selling_price, total_selling_price, profit, quantity, category_name")
	log.Printf("Executing Query: %s", query.Statement.SQL.String())

	// Fetch products associated with the sale_id
	if err := query.Find(&products).Error; err != nil {
		log.Printf("Error retrieving products for sale ID %d: %v", saleID, err)
		return errorResponse(c, http.StatusNotFound, "Products not found")
	}

	// Log the fetched products to ensure the data is correct
	for _, product := range products {
		log.Printf("Product Name: %s", product.ProductName)
		log.Printf("Product Details: %+v", product)
	}

	// Combine the sale details and products in the response
	response := struct {
		SaleID       int64     `json:"sale_id"`
		CashReceived float64   `json:"cash_received"`
		UserID       int64     `json:"user_id"`
		Date         string    `json:"date"`
		Products     []Product `json:"products"`
	}{
		SaleID:       saleDetails.SaleID,
		CashReceived: saleDetails.CashReceived,
		UserID:       saleDetails.UserID,
		Date:         saleDetails.Date,
		Products:     products, // Now products directly fits here
	}

	// Return the response as JSON
	return c.JSON(http.StatusOK, response)
}

// // GetSalesByDate retrieves all sales for a specific date from the sales_by_cash table
// func GetSalesByDate(c echo.Context) error {
// 	log.Println("[INFO] Received request to fetch sales for a specific date.")

// 	// Get the date parameter from the URL
// 	dateParam := c.Param("date")

// 	// Try to parse the date
// 	saleDate, err := time.Parse("2006-01-02", dateParam) // Format: "YYYY-MM-DD"
// 	if err != nil {
// 		logError("Error parsing date", err)
// 		return echo.NewHTTPError(http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD")
// 	}

// 	// Database connection
// 	db := getDB()
// 	if db == nil {
// 		logError("Database connection failed", nil)
// 		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
// 	}

// 	// Retrieve sales for the specific date
// 	var sales []models.Sale
// 	if err := db.Where("DATE(date) = ?", saleDate.Format("2006-01-02")).Find(&sales).Error; err != nil {
// 		logError("Error fetching sales for date", err)
// 		return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching sales for the date")
// 	}

// 	// Check if no sales were found
// 	if len(sales) == 0 {
// 		log.Printf("[INFO] No sales found for date: %v", saleDate)
// 		return echo.NewHTTPError(http.StatusNotFound, "No sales found for this date")
// 	}

// 	log.Printf("[INFO] Fetched %d sales records for date: %v", len(sales), saleDate)

// 	// Prepare the response
// 	var response []map[string]interface{}
// 	for _, sale := range sales {
// 		saleData := map[string]interface{}{
// 			"sale_id":             sale.SaleID,
// 			"product_name":        sale.Name,
// 			"quantity":            sale.Quantity,
// 			"unit_buying_price":   sale.UnitBuyingPrice,
// 			"total_buying_price":  sale.TotalBuyingPrice,
// 			"unit_selling_price":  sale.UnitSellingPrice,
// 			"user_id":             sale.UserID,
// 			"date":                sale.Date.Format("2006-01-02T15:04:05Z"), // ISO 8601 format
// 			"category_name":       sale.CategoryName,
// 			"total_selling_price": sale.TotalSellingPrice,
// 			"profit":              sale.Profit,
// 			"cash_receive":        sale.CashReceived,
// 			"balance":             sale.Balance,
// 		}
// 		response = append(response, saleData)
// 	}

// 	// Return the formatted response
// 	return c.JSON(http.StatusOK, response)
// }

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

	// Create a map to group sales by sale_id
	salesGrouped := make(map[int64]map[string]interface{})

	// Iterate through each sale and group products by sale_id
	for _, sale := range sales {
		// If the sale ID doesn't exist in the map, create a new entry
		if _, exists := salesGrouped[sale.SaleID]; !exists {
			salesGrouped[sale.SaleID] = map[string]interface{}{
				"sale_id":       sale.SaleID,
				"cash_received": sale.CashReceived,
				"user_id":       sale.UserID,
				"date":          sale.Date.Format("2006-01-02T15:04:05Z"), // ISO 8601 format
				"products":      []map[string]interface{}{},
			}
		}

		// Append the product data to the "products" field of the sale
		productData := map[string]interface{}{
			"name":                sale.Name,
			"unit_buying_price":   sale.UnitBuyingPrice,
			"total_buying_price":  sale.TotalBuyingPrice,
			"unit_selling_price":  sale.UnitSellingPrice,
			"total_selling_price": sale.TotalSellingPrice,
			"profit":              sale.Profit,
			"quantity":            sale.Quantity,
			"category_name":       sale.CategoryName,
		}

		// Add product to the sale
		salesGrouped[sale.SaleID]["products"] = append(salesGrouped[sale.SaleID]["products"].([]map[string]interface{}), productData)
	}

	// Prepare the response from the grouped data
	var response []map[string]interface{}
	for _, sale := range salesGrouped {
		response = append(response, sale)
	}

	// Return the formatted response
	return c.JSON(http.StatusOK, response)
}

// Define the Product struct outside the function to reuse it later

func GetAllSalesReports(c echo.Context) error {
	// Database connection
	db := getDB()
	if db == nil {
		log.Println("Failed to get database instance")
		return errorResponse(c, http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Retrieve all sale records
	var saleDetails []struct {
		SaleID       int64   `json:"sale_id"`
		CashReceived float64 `json:"cash_received"`
		UserID       int64   `json:"user_id"`
		Date         string  `json:"date"`
	}
	if err := db.Table("sales_by_cash").
		Select("sale_id, cash_received, user_id, date").
		Find(&saleDetails).Error; err != nil {
		log.Printf("Error retrieving all sale records: %v", err)
		return errorResponse(c, http.StatusInternalServerError, "Failed to retrieve sales records")
	}

	// Prepare a response array for the report
	var reports []struct {
		SaleID       int64     `json:"sale_id"`
		CashReceived float64   `json:"cash_received"`
		UserID       int64     `json:"user_id"`
		Date         string    `json:"date"`
		Products     []Product `json:"products"`
	}

	// Iterate over all sale records to fetch their products
	for _, sale := range saleDetails {
		// Fetch products for the current sale
		var products []Product

		if err := db.Table("sales_by_cash").
			Where("sale_id = ?", sale.SaleID).
			Select("name, unit_buying_price, total_buying_price, unit_selling_price, total_selling_price, profit, quantity, category_name").
			Find(&products).Error; err != nil {
			log.Printf("Error retrieving products for sale ID %d: %v", sale.SaleID, err)
			continue
		}

		// Ensure that products are not empty (this should not happen as long as the data is consistent)
		if len(products) == 0 {
			log.Printf("No products found for sale ID %d", sale.SaleID)
			continue
		}

		// Add the sale and its products to the report
		reports = append(reports, struct {
			SaleID       int64     `json:"sale_id"`
			CashReceived float64   `json:"cash_received"`
			UserID       int64     `json:"user_id"`
			Date         string    `json:"date"`
			Products     []Product `json:"products"`
		}{
			SaleID:       sale.SaleID,
			CashReceived: sale.CashReceived,
			UserID:       sale.UserID,
			Date:         sale.Date,
			Products:     products,
		})
	}

	// Return the report as JSON
	return c.JSON(http.StatusOK, reports)
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
