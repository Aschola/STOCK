// controllers/sales.go
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

func SellProduct(c echo.Context) error {
	// Log the incoming request
	log.Println("[INFO] Received request to sell products.")

	// Retrieve organizationID from context
	organizationID, err := getOrganizationID(c)
	if err != nil {
		return err
	}

	// Parse the incoming request body
	var payload models.SalePayload
	if err := json.NewDecoder(c.Request().Body).Decode(&payload); err != nil {
		log.Printf("[ERROR] Error parsing request body: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input data")
	}

	// Log the parsed payload
	log.Printf("[INFO] Parsed Payload: %+v", payload)

	// Database connection
	db := getDB()
	if db == nil {
		log.Println("[ERROR] Database connection failed")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Start a transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("[ERROR] Unexpected error: %v", r)
		}
	}()
	defer tx.Rollback() // Rollback in case of failure

	// Generate a common sale_id for this transaction
	saleID := time.Now().UnixNano() // Unique ID based on timestamp (in nanoseconds)

	// Calculate total selling price
	var totalSellingPrice float64
	for _, item := range payload.Items {
		// Retrieve product details from the products table
		var product models.Product
		if err := tx.First(&product, "product_id = ?", item.ProductID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				log.Printf("[ERROR] Product not found for product_id: %d", item.ProductID)
				return echo.NewHTTPError(http.StatusNotFound, "Product not found")
			}
			log.Printf("[ERROR] Error fetching product details: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching product details")
		}

		// Retrieve stock information for the product
		var stock models.Stock
		if err := tx.First(&stock, "product_id = ?", item.ProductID).Error; err != nil {
			log.Printf("[ERROR] Error fetching stock details for product_id: %d", item.ProductID)
			return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching stock details")
		}

		// Check stock availability
		if stock.Quantity < item.QuantitySold {
			log.Printf("[ERROR] Insufficient stock for product_id: %d. Available: %d, Requested: %d", item.ProductID, stock.Quantity, item.QuantitySold)
			return echo.NewHTTPError(http.StatusBadRequest, "Insufficient stock")
		}

		// Update total selling price for the transaction
		totalSellingPrice += float64(item.QuantitySold) * stock.SellingPrice

		// Update stock quantity
		stock.Quantity -= item.QuantitySold
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

		totalCost := float64(item.QuantitySold) * stock.BuyingPrice
		profit := (float64(item.QuantitySold) * stock.SellingPrice) - totalCost
		balance := payload.CashReceived - totalSellingPrice

		sale := models.Sale{
			SaleID:            saleID,         // Use the common sale_id for all items in this transaction
			OrganizationsID:   organizationID, // Use OrganizationsID to associate with the correct organization
			Name:              product.ProductName,
			CategoryName:      product.CategoryName,
			UnitBuyingPrice:   stock.BuyingPrice,
			TotalBuyingPrice:  totalCost,
			UnitSellingPrice:  stock.SellingPrice,
			TotalSellingPrice: float64(item.QuantitySold) * stock.SellingPrice,
			Profit:            profit,
			Quantity:          item.QuantitySold,
			CashReceived:      payload.CashReceived,
			Balance:           balance,
			UserID:            int64(payload.UserID), // Convert int to int64 here
			Date:              time.Now(),
		}

		// Log the sale object before inserting into the database
		log.Printf("[INFO] Sale Object: %+v", sale)

		// Insert the sale into the database
		if err := tx.Create(&sale).Error; err != nil {
			log.Printf("[ERROR] Error recording sale for product_id: %d: %v", item.ProductID, err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Error recording sale")
		}

		// Log successful sale record
		log.Printf("[INFO] Sale recorded successfully for product_id: %d", item.ProductID)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("[ERROR] Error committing transaction: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error committing transaction")
	}

	// Log successful transaction commit
	log.Println("[INFO] Transaction committed successfully")

	// Return success response
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Cash sale processed successfully for all items",
	})
}
