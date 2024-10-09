package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	models "stock/models"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// CashSaleRequest represents the request data for a cash sale.
type CashSaleRequest struct {
	ProductID    int     `json:"product_id"`
	QuantitySold int     `json:"quantity_sold"`
	UserID       string  `json:"user_id"`
	CashReceived float64 `json:"cash_received"` // Add this field
}

func SellProductByCash(c echo.Context) error {
	var request CashSaleRequest

	if err := c.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}

	// Validate input
	if request.QuantitySold <= 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Quantity sold must be greater than zero")
	}

	db := getDB()
	if db == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Begin a transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Retrieve the product details
	var product models.Product
	if err := tx.Table("active_products").First(&product, request.ProductID).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Product not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	}

	// Check if enough quantity is available
	if product.Quantity < request.QuantitySold {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusBadRequest, "Insufficient quantity")
	}

	// Calculate total cost, total selling price, profit
	totalCost := float64(request.QuantitySold) * product.BuyingPrice
	totalSellingPrice := float64(request.QuantitySold) * product.SellingPrice
	profit := totalSellingPrice - totalCost

	// Update the quantity of the product in the 'active_products' table
	updatedQuantity := product.Quantity - request.QuantitySold
	if err := tx.Table("active_products").Where("product_id = ?", request.ProductID).Update("quantity", updatedQuantity).Error; err != nil {
		tx.Rollback()
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	}

	// Calculate balance
	balance := request.CashReceived - totalSellingPrice

	// Create cash sale record
	cashSale := models.SalebyCash{
		Name:              product.ProductName,
		UnitBuyingPrice:   product.BuyingPrice,
		Quantity:          request.QuantitySold,
		UserID:            request.UserID,
		Date:              time.Now(),
		CategoryName:      product.CategoryName,
		TotalBuyingPrice:  totalCost,
		UnitSellingPrice:  product.SellingPrice,
		TotalSellingPrice: totalSellingPrice,
		Profit:            profit,
		CashReceive:       request.CashReceived,
		Balance:           balance,
	}

	// Insert the sale into the sales_by_cash table
	if err := tx.Table("sales_by_cash").Create(&cashSale).Error; err != nil {
		tx.Rollback()
		log.Printf("Error inserting cash sale: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to commit transaction")
	}

	// Return success response with total cost and balance
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":             "Cash sale processed successfully",
		"product_id":          request.ProductID,
		"quantity_sold":       request.QuantitySold,
		"remaining_qty":       updatedQuantity,
		"total_cost":          totalCost,
		"total_selling_price": totalSellingPrice,
		"balance":             balance,
	})
}

// GetCashSales fetches all cash sales from the database.
func GetCashSales(c echo.Context) error {
	log.Println("Received request to fetch cash sales")

	db := getDB()
	if db == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
	}

	var sales []models.SalebyCash
	if err := db.Table("sales_by_cash").Find(&sales).Error; err != nil {
		log.Printf("Error querying cash sales from database: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	}

	log.Printf("Fetched %d cash sales", len(sales))
	return c.JSON(http.StatusOK, sales)
}

// GetCashSaleByID fetches a single sale by its ID from the sales_by_cash table.
func GetCashSaleByID(c echo.Context) error {
	log.Println("Received request to fetch sale by ID")

	saleID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("Invalid sale ID: %s", c.Param("id"))
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid sale ID")
	}

	db := getDB()
	if db == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
	}

	var sale models.SalebyCash
	if err := db.Table("sales_by_cash").First(&sale, saleID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("Sale not found with ID: %d", saleID)
			return echo.NewHTTPError(http.StatusNotFound, "Sale not found")
		}
		log.Printf("Error querying sale: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch sale")
	}

	log.Printf("Fetched sale: %+v", sale)
	return c.JSON(http.StatusOK, sale)
}

// AddSaleByCash adds a new sale record to the sales_by_cash table.
func AddSaleByCash(c echo.Context) error {
	db := getDB()
	if db == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
	}

	var sale models.SalebyCash
	if err := json.NewDecoder(c.Request().Body).Decode(&sale); err != nil {
		log.Printf("Error decoding JSON: %s", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, "Error decoding JSON")
	}

	log.Printf("Received request to create a sale: %+v", sale)

	if err := db.Table("sales_by_cash").Create(&sale).Error; err != nil {
		log.Printf("Error inserting sale: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Error inserting sale")
	}

	log.Printf("Created new sale with ID: %d", sale.SaleID)
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "Sale created successfully",
		"sale_id": sale.SaleID,
	})
}

// DeleteSaleByCash deletes a sale record from the sales_by_cash table.
func DeleteSaleByCash(c echo.Context) error {
	saleID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("Invalid sale ID: %s", c.Param("id"))
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid sale ID")
	}

	db := getDB()
	if db == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
	}

	if err := db.Table("sales_by_cash").Delete(&models.SalebyCash{}, saleID).Error; err != nil {
		log.Printf("Error deleting sale ID %d: %s", saleID, err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Error deleting sale")
	}

	log.Printf("Deleted sale ID %d successfully from sales_by_cash", saleID)
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Sale deleted successfully",
	})
}
