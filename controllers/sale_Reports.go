package controllers

import (
	"log"
	"net/http"
	"stock/models"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

// Define the Product struct outside the function to reuse it later
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

// Helper function to handle database connection errors
func handleDBError(c echo.Context, err error, message string) error {
	log.Printf("[ERROR] %s: %v", message, err)
	return echo.NewHTTPError(http.StatusInternalServerError, message)
}

// // GetAllSales retrieves all sales records from the sales_by_cash table
// func GetAllSales(c echo.Context) error {
// 	log.Println("[INFO] Received request to fetch all sales records.")

// 	// Retrieve organizationID from context
// 	organizationID, err := getOrganizationID(c)
// 	if err != nil {
// 		return err
// 	}

// 	// Database connection
// 	db := getDB()
// 	if db == nil {
// 		return handleDBError(c, nil, "Failed to connect to the database")
// 	}

// 	// Retrieve all sales records for the given organizationID
// 	var sales []models.Sale
// 	if err := db.Where("organizations_id = ?", organizationID).Find(&sales).Error; err != nil {
// 		return handleDBError(c, err, "Error fetching sales records")
// 	}

// 	// Check if no sales were found
// 	if len(sales) == 0 {
// 		return echo.NewHTTPError(http.StatusNotFound, "No sales records found")
// 	}

// 	// Prepare the response
// 	// Group the sales by sale_id
// 	saleMap := make(map[int64]map[string]interface{})
// 	for _, sale := range sales {
// 		// If the sale_id already exists in the map, append the product to the "products" list
// 		if _, exists := saleMap[sale.SaleID]; exists {
// 			// Append the product to the "products" list
// 			saleMap[sale.SaleID]["products"] = append(saleMap[sale.SaleID]["products"].([]map[string]interface{}), map[string]interface{}{
// 				"category_name":       sale.CategoryName,
// 				"name":                sale.Name,
// 				"profit":              sale.Profit,
// 				"quantity":            sale.Quantity,
// 				"total_buying_price":  sale.TotalBuyingPrice,
// 				"total_selling_price": sale.TotalSellingPrice,
// 				"unit_buying_price":   sale.UnitBuyingPrice,
// 				"unit_selling_price":  sale.UnitSellingPrice,
// 			})
// 		} else {
// 			// If the sale_id doesn't exist in the map, create a new entry with this sale
// 			saleMap[sale.SaleID] = map[string]interface{}{
// 				"sale_id":          sale.SaleID,
// 				"user_id":          sale.UserID,
// 				"cash_received":    sale.CashReceived,
// 				"date":             sale.Date.Format("2006-01-02T15:04:05Z"),
// 				"organizations_id": sale.OrganizationsID, // Adding organizations_id
// 				"products": []map[string]interface{}{
// 					{
// 						"category_name":       sale.CategoryName,
// 						"name":                sale.Name,
// 						"profit":              sale.Profit,
// 						"quantity":            sale.Quantity,
// 						"total_buying_price":  sale.TotalBuyingPrice,
// 						"total_selling_price": sale.TotalSellingPrice,
// 						"unit_buying_price":   sale.UnitBuyingPrice,
// 						"unit_selling_price":  sale.UnitSellingPrice,
// 					},
// 				},
// 			}
// 		}
// 	}

// 	// Convert the map to a slice of sale entries
// 	var response []map[string]interface{}
// 	for _, saleData := range saleMap {
// 		response = append(response, saleData)
// 	}

// 	// Return the formatted response
// 	return c.JSON(http.StatusOK, response)
// }

// GetAllSales retrieves all sales records from the sales_transactions table and includes payment_mode
// GetAllSales retrieves all sales records from the sales_transactions table and includes payment_mode and username
func GetAllSales(c echo.Context) error {
	log.Println("[INFO] Received request to fetch all sales records.")

	// Retrieve organizationID from context
	organizationID, err := getOrganizationID(c)
	if err != nil {
		return err
	}

	// Database connection
	db := getDB()
	if db == nil {
		return handleDBError(c, nil, "Failed to connect to the database")
	}

	// Retrieve all sales records for the given organizationID and join with users table to get usernames
	var sales []models.Sale
	query := db.Table("sales_transactions").
		Select("sales_transactions.*, sales_transactions.payment_mode, users.username").
		Joins("JOIN users ON sales_transactions.user_id = users.id").
		Where("sales_transactions.organizations_id = ?", organizationID)

	// Apply condition for Mpesa payment mode, only fetch completed transactions for Mpesa
	query = query.Where("sales_transactions.payment_mode != ? OR (sales_transactions.payment_mode = ? AND sales_transactions.transaction_status = ?)",
		"Mpesa", "Mpesa", "COMPLETED")

	// Execute the query
	if err := query.Find(&sales).Error; err != nil {
		return handleDBError(c, err, "Error fetching sales records")
	}

	// Check if no sales were found
	if len(sales) == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "No sales records found")
	}

	// Prepare the response, group the sales by sale_id
	saleMap := make(map[int64]map[string]interface{})
	for _, sale := range sales {
		// If the sale_id already exists in the map, append the product to the "products" list
		if _, exists := saleMap[sale.SaleID]; exists {
			// Append the product to the "products" list
			saleMap[sale.SaleID]["products"] = append(saleMap[sale.SaleID]["products"].([]map[string]interface{}), map[string]interface{}{
				"category_name":       sale.CategoryName,
				"name":                sale.Name,
				"profit":              sale.Profit,
				"quantity":            sale.Quantity,
				"total_buying_price":  sale.TotalBuyingPrice,
				"total_selling_price": sale.TotalSellingPrice,
				"unit_buying_price":   sale.UnitBuyingPrice,
				"unit_selling_price":  sale.UnitSellingPrice,
			})
		} else {
			// If the sale_id doesn't exist in the map, create a new entry with this sale
			saleMap[sale.SaleID] = map[string]interface{}{
				"sale_id":          sale.SaleID,
				"user_id":          sale.UserID,
				"username":         sale.Username, // Include the username
				"cash_received":    sale.CashReceived,
				"date":             sale.Date.Format("2006-01-02T15:04:05Z"),
				"organizations_id": sale.OrganizationsID,
				"payment_mode":     sale.PaymentMode, // Adding payment_mode from sales_transactions
				"products": []map[string]interface{}{
					{
						"category_name":       sale.CategoryName,
						"name":                sale.Name,
						"profit":              sale.Profit,
						"quantity":            sale.Quantity,
						"total_buying_price":  sale.TotalBuyingPrice,
						"total_selling_price": sale.TotalSellingPrice,
						"unit_buying_price":   sale.UnitBuyingPrice,
						"unit_selling_price":  sale.UnitSellingPrice,
					},
				},
			}
		}
	}

	// Convert the map to a slice of sale entries
	var response []map[string]interface{}
	for _, saleData := range saleMap {
		response = append(response, saleData)
	}

	// Return the formatted response
	return c.JSON(http.StatusOK, response)
}

// GetSalesByDate retrieves sales for a specific date
func GetSalesByDate(c echo.Context) error {
	dateParam := c.Param("date")
	saleDate, err := time.Parse("2006-01-02", dateParam)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD")
	}

	// Database connection
	db := getDB()
	if db == nil {
		return handleDBError(c, nil, "Failed to connect to the database")
	}

	// Retrieve sales for the specific date
	var sales []models.Sale
	if err := db.Where("DATE(date) = ?", saleDate.Format("2006-01-02")).Find(&sales).Error; err != nil {
		return handleDBError(c, err, "Error fetching sales for date")
	}

	// Check if no sales found
	if len(sales) == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "No sales found for this date")
	}

	// Group sales by sale_id
	salesGrouped := make(map[int64]map[string]interface{})
	for _, sale := range sales {
		if _, exists := salesGrouped[sale.SaleID]; !exists {
			salesGrouped[sale.SaleID] = map[string]interface{}{
				"sale_id":       sale.SaleID,
				"cash_received": sale.CashReceived,
				"user_id":       sale.UserID,
				"date":          sale.Date.Format("2006-01-02T15:04:05Z"),
				"products":      []map[string]interface{}{},
			}
		}

		// Add products to the sale
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
		salesGrouped[sale.SaleID]["products"] = append(salesGrouped[sale.SaleID]["products"].([]map[string]interface{}), productData)
	}

	// Prepare the response
	var response []map[string]interface{}
	for _, sale := range salesGrouped {
		response = append(response, sale)
	}

	// Return the response as JSON
	return c.JSON(http.StatusOK, response)
}

// GetSalesByUserID retrieves all sales records for a specific user ID
func GetSalesByUserID(c echo.Context) error {
	userID := c.Param("user_id") // Retrieve the user_id from the request URL

	// Convert userID to int64 (since user_id is an integer)
	userIDInt, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user_id format")
	}

	// Database connection
	db := getDB()
	if db == nil {
		return handleDBError(c, nil, "Failed to connect to the database")
	}

	// Retrieve all sales records for the given user_id
	var sales []models.Sale
	if err := db.Where("user_id = ?", userIDInt).Find(&sales).Error; err != nil {
		return handleDBError(c, err, "Error fetching sales records for user")
	}

	// Check if no sales were found
	if len(sales) == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "No sales records found for this user")
	}

	// Group the sales by sale_id
	saleMap := make(map[int64]map[string]interface{})
	for _, sale := range sales {
		if _, exists := saleMap[sale.SaleID]; exists {
			// Append the product to the "products" list
			saleMap[sale.SaleID]["products"] = append(saleMap[sale.SaleID]["products"].([]map[string]interface{}), map[string]interface{}{
				"category_name":       sale.CategoryName,
				"name":                sale.Name,
				"profit":              sale.Profit,
				"quantity":            sale.Quantity,
				"total_buying_price":  sale.TotalBuyingPrice,
				"total_selling_price": sale.TotalSellingPrice,
				"unit_buying_price":   sale.UnitBuyingPrice,
				"unit_selling_price":  sale.UnitSellingPrice,
			})
		} else {
			// If the sale_id doesn't exist in the map, create a new entry with this sale
			saleMap[sale.SaleID] = map[string]interface{}{
				"sale_id":          sale.SaleID,
				"user_id":          sale.UserID,
				"cash_received":    sale.CashReceived,
				"date":             sale.Date.Format("2006-01-02T15:04:05Z"),
				"organizations_id": sale.OrganizationsID,
				"products": []map[string]interface{}{
					{
						"category_name":       sale.CategoryName,
						"name":                sale.Name,
						"profit":              sale.Profit,
						"quantity":            sale.Quantity,
						"total_buying_price":  sale.TotalBuyingPrice,
						"total_selling_price": sale.TotalSellingPrice,
						"unit_buying_price":   sale.UnitBuyingPrice,
						"unit_selling_price":  sale.UnitSellingPrice,
					},
				},
			}
		}
	}

	// Convert the map to a slice of sale entries
	var response []map[string]interface{}
	for _, saleData := range saleMap {
		response = append(response, saleData)
	}

	// Return the formatted response
	return c.JSON(http.StatusOK, response)
}

func GetSalesByUsername(c echo.Context) error {
	username := c.Param("username") // Retrieve the username from the request URL

	// Database connection
	db := getDB()
	if db == nil {
		return handleDBError(c, nil, "Failed to connect to the database")
	}

	// Retrieve all sales records for the given username by joining with the users table
	var sales []models.Sale
	if err := db.Table("sales_by_cash").
		Joins("JOIN users u ON sales_by_cash.user_id = u.id").
		Where("u.username = ?", username).
		Find(&sales).Error; err != nil {
		return handleDBError(c, err, "Error fetching sales records for user")
	}

	// Check if no sales were found
	if len(sales) == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "No sales records found for this user")
	}

	// Group the sales by sale_id
	saleMap := make(map[int64]map[string]interface{})
	for _, sale := range sales {
		if _, exists := saleMap[sale.SaleID]; exists {
			// Append the product to the "products" list
			saleMap[sale.SaleID]["products"] = append(saleMap[sale.SaleID]["products"].([]map[string]interface{}), map[string]interface{}{
				"category_name":       sale.CategoryName,
				"name":                sale.Name,
				"profit":              sale.Profit,
				"quantity":            sale.Quantity,
				"total_buying_price":  sale.TotalBuyingPrice,
				"total_selling_price": sale.TotalSellingPrice,
				"unit_buying_price":   sale.UnitBuyingPrice,
				"unit_selling_price":  sale.UnitSellingPrice,
			})
		} else {
			// If the sale_id doesn't exist in the map, create a new entry with this sale
			saleMap[sale.SaleID] = map[string]interface{}{
				"sale_id":          sale.SaleID,
				"user_id":          sale.UserID,
				"cash_received":    sale.CashReceived,
				"date":             sale.Date.Format("2006-01-02T15:04:05Z"),
				"organizations_id": sale.OrganizationsID,
				"products": []map[string]interface{}{
					{
						"category_name":       sale.CategoryName,
						"name":                sale.Name,
						"profit":              sale.Profit,
						"quantity":            sale.Quantity,
						"total_buying_price":  sale.TotalBuyingPrice,
						"total_selling_price": sale.TotalSellingPrice,
						"unit_buying_price":   sale.UnitBuyingPrice,
						"unit_selling_price":  sale.UnitSellingPrice,
					},
				},
			}
		}
	}

	// Convert the map to a slice of sale entries
	var response []map[string]interface{}
	for _, saleData := range saleMap {
		response = append(response, saleData)
	}

	// Return the formatted response
	return c.JSON(http.StatusOK, response)
}

// func GetSalesBySaleID(c echo.Context) error {
// 	saleID := c.Param("sale_id") // Retrieve the sale_id from the request URL

// 	// Database connection
// 	db := getDB()
// 	if db == nil {
// 		return handleDBError(c, nil, "Failed to connect to the database")
// 	}

// 	// Retrieve the sale record for the given sale_id
// 	var sale models.Sale
// 	if err := db.Table("sales_by_cash").
// 		Joins("JOIN users u ON sales_by_cash.user_id = u.id").
// 		Where("sales_by_cash.sale_id = ?", saleID).
// 		First(&sale).Error; err != nil {
// 		return handleDBError(c, err, "Error fetching sale record by sale_id")
// 	}

// 	// Map the result into a custom response structure
// 	saleResponse := map[string]interface{}{
// 		"sale_id":          sale.SaleID,
// 		"user_id":          sale.UserID,
// 		"username":         sale.User.Username, // assuming User model is populated correctly
// 		"cash_received":    sale.CashReceived,
// 		"date":             sale.Date.Format("2006-01-02T15:04:05Z"),
// 		"organizations_id": sale.OrganizationsID,
// 		"products": []map[string]interface{}{
// 			{
// 				"category_name":       sale.CategoryName,
// 				"name":                sale.Name,
// 				"profit":              sale.Profit,
// 				"quantity":            sale.Quantity,
// 				"total_buying_price":  sale.TotalBuyingPrice,
// 				"total_selling_price": sale.TotalSellingPrice,
// 				"unit_buying_price":   sale.UnitBuyingPrice,
// 				"unit_selling_price":  sale.UnitSellingPrice,
// 			},
// 		},
// 	}

// 	// Return the formatted sale response
// 	return c.JSON(http.StatusOK, saleResponse)
// }