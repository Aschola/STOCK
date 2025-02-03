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

// Helper function to log errors
func logError(message string, err error) {
	if err != nil {
		log.Printf("[ERROR] %s: %v", message, err)
	}
}

// SellProduct function with Mpesa STK Push integration
func SellProduct(c echo.Context) error {
	log.Println("[INFO] Received request to sell products.")

	organizationID, err := getOrganizationID(c)
	if err != nil {
		return err
	}

	// Parse the payload to extract Sale data
	var payload models.SalePayload
	if err := json.NewDecoder(c.Request().Body).Decode(&payload); err != nil {
		log.Printf("[ERROR] Error parsing request body: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input data")
	}

	// Log the parsed payload for debugging
	log.Printf("[INFO] Parsed Payload: %+v", payload)

	// Database connection
	db := getDB()
	if db == nil {
		log.Println("[ERROR] Database connection failed")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
	}

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("[ERROR] Unexpected error: %v", r)
		}
	}()
	defer tx.Rollback()

	saleID := time.Now().Unix()
	var totalSellingPrice float64

	// Process each item in the sale
	for _, item := range payload.Items {
		var product models.Product
		if err := tx.First(&product, "product_id = ?", item.ProductID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				log.Printf("[ERROR] Product not found for product_id: %d", item.ProductID)
				return echo.NewHTTPError(http.StatusNotFound, "Product not found")
			}
			log.Printf("[ERROR] Error fetching product details: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching product details")
		}

		var stock models.Stock
		if err := tx.First(&stock, "product_id = ?", item.ProductID).Error; err != nil {
			log.Printf("[ERROR] Error fetching stock details for product_id: %d", item.ProductID)
			return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching stock details")
		}

		// Ensure there is sufficient stock
		if stock.Quantity < item.QuantitySold {
			log.Printf("[ERROR] Insufficient stock for product_id: %d", item.ProductID)
			return echo.NewHTTPError(http.StatusBadRequest, "Insufficient stock")
		}

		// Calculate total selling price
		totalSellingPrice += float64(item.QuantitySold) * stock.SellingPrice
		stock.Quantity -= item.QuantitySold

		// Update stock after sale
		if err := tx.Save(&stock).Error; err != nil {
			log.Printf("[ERROR] Error updating stock for product_id: %d: %v", item.ProductID, err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Error updating stock")
		}
	}

	// Check if the cash received is enough for the total selling price
	if payload.CashReceived < totalSellingPrice {
		log.Printf("[ERROR] Insufficient cash received. Received: %f, Required: %f", payload.CashReceived, totalSellingPrice)
		return echo.NewHTTPError(http.StatusBadRequest, "Insufficient cash received")
	}

	// Process each item in the sale and create sale records
	for _, item := range payload.Items {
		var product models.Product
		if err := tx.First(&product, "product_id = ?", item.ProductID).Error; err != nil {
			log.Printf("[ERROR] Error fetching product details: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching product details")
		}

		var stock models.Stock
		if err := tx.First(&stock, "product_id = ?", item.ProductID).Error; err != nil {
			log.Printf("[ERROR] Error fetching stock details for product_id: %d", item.ProductID)
			return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching stock details")
		}

		// Calculate cost, profit, and balance
		totalCost := float64(item.QuantitySold) * stock.BuyingPrice
		profit := (float64(item.QuantitySold) * stock.SellingPrice) - totalCost
		balance := payload.CashReceived - totalSellingPrice

		// Sale object creation
		sale := models.Sale{
			SaleID:            saleID,
			OrganizationsID:   organizationID,
			Name:              product.ProductName,
			CategoryName:      product.CategoryName,
			UnitBuyingPrice:   stock.BuyingPrice,
			TotalBuyingPrice:  totalCost,
			UnitSellingPrice:  stock.SellingPrice,
			TotalSellingPrice: int64(float64(item.QuantitySold) * stock.SellingPrice),
			Profit:            profit,
			Quantity:          item.QuantitySold,
			CashReceived:      payload.CashReceived,
			Balance:           balance,
			UserID:            int64(payload.UserID), // Convert int to int64 here
			Date:              time.Now(),
		}

		// Log the sale object before inserting into the database
		log.Printf("[INFO] Sale Object: %+v", sale)

		// Save the sale record
		if err := tx.Create(&sale).Error; err != nil {
			log.Printf("[ERROR] Error recording sale for product_id: %d: %v", item.ProductID, err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Error recording sale")
		}

		log.Printf("[INFO] Sale recorded successfully for product_id: %d", item.ProductID)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("[ERROR] Error committing transaction: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error committing transaction")
	}

	log.Println("[INFO] Transaction committed successfully")

	// Return a successful response
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Cash sale processed successfully for all items",
	})
}
