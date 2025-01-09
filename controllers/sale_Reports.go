package controllers

import (
	"log"
	"net/http"
	"stock/models"

	"github.com/labstack/echo/v4"
)

// GetAllSales retrieves all sales records from the sales_by_cash table
func GetAllSales(c echo.Context) error {
	log.Println("[INFO] Received request to fetch all sales records.")

	// Database connection
	db := getDB()
	if db == nil {
		logError("Database connection failed", nil)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Retrieve all sales records
	var sales []models.Sale
	if err := db.Find(&sales).Error; err != nil {
		logError("Error fetching all sales records", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching sales records")
	}

	// Check if no sales were found
	if len(sales) == 0 {
		log.Println("[INFO] No sales records found.")
		return echo.NewHTTPError(http.StatusNotFound, "No sales records found")
	}

	log.Printf("[INFO] Fetched %d sales records", len(sales))

	// Prepare the response
	var response []map[string]interface{}
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
			"cash_receive":        sale.CashReceive,
			"balance":             sale.Balance,
			"user_id":             sale.UserID,
			"date":                sale.Date.Format("2006-01-02T15:04:05Z"), // ISO 8601 format
			"category_name":       sale.CategoryName,
		}
		response = append(response, saleData)
	}

	// Return the formatted response
	return c.JSON(http.StatusOK, response)
}
