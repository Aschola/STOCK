package controllers

import (
	"log"
	"net/http"
	"stock/models"
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

// GetAllSales retrieves all sales records from the sales_by_cash table
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

	// Retrieve all sales records for the given organizationID
	var sales []models.Sale
	if err := db.Where("organizations_id = ?", organizationID).Find(&sales).Error; err != nil {
		return handleDBError(c, err, "Error fetching sales records")
	}

	// Check if no sales were found
	if len(sales) == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "No sales records found")
	}

	// Prepare the response
	// Group the sales by sale_id
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
				"cash_received":    sale.CashReceived,
				"date":             sale.Date.Format("2006-01-02T15:04:05Z"),
				"organizations_id": sale.OrganizationsID, // Adding organizations_id
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

// GetSalesBySaleID retrieves sale and product details for a specific sale ID
func GetSalesBySaleID(c echo.Context) error {
	saleID := c.Param("sale_id")

	// Database connection
	db := getDB()
	if db == nil {
		return handleDBError(c, nil, "Failed to connect to the database")
	}

	// Fetch sale details
	var saleDetails struct {
		SaleID       int64   `json:"sale_id"`
		CashReceived float64 `json:"cash_received"`
		UserID       int64   `json:"user_id"`
		Date         string  `json:"date"`
	}
	if err := db.Table("sales_by_cash").Where("sale_id = ?", saleID).First(&saleDetails).Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Sale details not found")
	}

	// Fetch associated products
	var products []Product
	if err := db.Table("sales_by_cash").
		Where("sale_id = ?", saleID).
		Select("name, unit_buying_price, total_buying_price, unit_selling_price, total_selling_price, profit, quantity, category_name").
		Find(&products).Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "Products not found")
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
		Products:     products,
	}

	// Return the response as JSON
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
