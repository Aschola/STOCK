package controllers

import (
	"log"
	"net/http"
	"stock/models"
	//"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"fmt"
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




func PostTotalSales(c echo.Context) error {
    organizationID, err := getOrganizationID(c)
    if err != nil {
        return err
    }

    var request struct {
        TotalSellingPrice float64 `json:"total_selling_price"`
    }

    if err := c.Bind(&request); err != nil {
        return echo.NewHTTPError(http.StatusBadRequest, "Invalid request payload")
    }

    db := getDB()
    if db == nil {
        return handleDBError(c, nil, "Failed to connect to the database")
    }

    totalSale := models.TotalSales{
        OrganizationID:    organizationID,
        TotalSellingPrice: request.TotalSellingPrice,
        Date:             time.Now(),
    }

    if err := db.Create(&totalSale).Error; err != nil {
        return handleDBError(c, err, "Failed to insert total sales")
    }

    return c.JSON(http.StatusCreated, totalSale)
}

func GetAllTotalSales(c echo.Context) error {
	// Get organization ID from context
	organizationID, err := getOrganizationID(c)
	if err != nil {
		return err
	}

	// Get current month and year
	now := time.Now()
	month := c.QueryParam("month")
	year := c.QueryParam("year")

	// Use current month and year if not provided
	if month == "" {
		month = fmt.Sprintf("%02d", int(now.Month()))
	}
	if year == "" {
		year = fmt.Sprintf("%d", now.Year())
	}

	// Database connection
	db := getDB()
	if db == nil {
		return handleDBError(c, nil, "Failed to connect to the database")
	}

	// Retrieve sales for the given month and year
	var totalSales []models.TotalSales
	if err := db.Where("organization_id = ? AND MONTH(date) = ? AND YEAR(date) = ?", organizationID, month, year).
		Find(&totalSales).Error; err != nil {
		return handleDBError(c, err, "Error fetching total sales")
	}

	// Check if no data found
	if len(totalSales) == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "No sales found for this month")
	}

	// Return response
	return c.JSON(http.StatusOK, totalSales)
}
