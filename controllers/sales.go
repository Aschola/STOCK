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

// SellProduct function with Mpesa STK Push integration
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

	// Ensure that the correct payment mode is coming through
	log.Printf("[INFO] Payment Mode: %s", payload.PaymentMode)

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

	// Generate a shortened sale_id using Unix timestamp in seconds (10-digit ID)
	saleID := time.Now().Unix() // Unique ID based on timestamp (in seconds)

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

	// Handle payment mode: "cash" or "mpesa"
	var paymentMode string
	log.Printf("[INFO] Incoming payment mode from payload: %s", payload.PaymentMode)

	// Ensure the payment mode is set to either cash or mpesa
	if payload.PaymentMode == "cash" {
		paymentMode = "cash"
	} else if payload.PaymentMode == "Mpesa" {

		paymentMode = "Mpesa"

		////// If payment mode is Mpesa, initiate STK Push
		// log.Printf("[INFO] Initiating Mpesa STK Push for the sale")
		// err := initiateMpesaSTKPush(totalSellingPrice, payload.CustomerPhoneNumber)
		// if err != nil {
		// 	log.Printf("[ERROR] Error initiating Mpesa STK Push: %v", err)
		// 	return echo.NewHTTPError(http.StatusInternalServerError, "Error initiating Mpesa payment")
		// }
		// log.Printf("[INFO] Mpesa STK Push initiated successfully.")
	} else {
		log.Printf("[ERROR] Invalid payment mode received: %s", payload.PaymentMode)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid payment mode")
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

		// Sale object creation
		sale := models.Sale{
			SaleID:            saleID,         // Use the shortened sale_id generated above
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
			PaymentMode:       paymentMode,           // Ensure correct PaymentMode is used
			UserID:            int64(payload.UserID), // Convert int to int64 here
			Date:              time.Now(),
		}

		// Log the sale object before inserting into the database
		log.Printf("[INFO] Creating sale with payment mode: %s", paymentMode)
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
		"message": "Sale processed successfully for all items",
	})
}

// func SellProduct(c echo.Context) error {
// 	// Log the incoming request
// 	log.Println("[INFO] Received request to sell products.")

// 	// Retrieve organizationID from context
// 	organizationID, err := getOrganizationID(c)
// 	if err != nil {
// 		return err
// 	}

// 	// Parse the incoming request body
// 	var payload models.SalePayload
// 	if err := json.NewDecoder(c.Request().Body).Decode(&payload); err != nil {
// 		log.Printf("[ERROR] Error parsing request body: %v", err)
// 		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input data")
// 	}

// 	// Log the parsed payload
// 	log.Printf("[INFO] Parsed Payload: %+v", payload)

// 	// Ensure that the correct payment mode is coming through
// 	log.Printf("[INFO] Payment Mode: %s", payload.PaymentMode)

// 	// Database connection
// 	db := getDB()
// 	if db == nil {
// 		log.Println("[ERROR] Database connection failed")
// 		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
// 	}

// 	// Start a transaction
// 	tx := db.Begin()
// 	defer func() {
// 		if r := recover(); r != nil {
// 			tx.Rollback()
// 			log.Printf("[ERROR] Unexpected error: %v", r)
// 		}
// 	}()
// 	defer tx.Rollback() // Rollback in case of failure

// 	// Generate a shortened sale_id using Unix timestamp in seconds (10-digit ID)
// 	saleID := time.Now().Unix() // Unique ID based on timestamp (in seconds)

// 	// Calculate total selling price
// 	var totalSellingPrice float64
// 	for _, item := range payload.Items {
// 		// Retrieve product details from the products table
// 		var product models.Product
// 		if err := tx.First(&product, "product_id = ?", item.ProductID).Error; err != nil {
// 			if err == gorm.ErrRecordNotFound {
// 				log.Printf("[ERROR] Product not found for product_id: %d", item.ProductID)
// 				return echo.NewHTTPError(http.StatusNotFound, "Product not found")
// 			}
// 			log.Printf("[ERROR] Error fetching product details: %v", err)
// 			return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching product details")
// 		}

// 		// Retrieve stock information for the product
// 		var stock models.Stock
// 		if err := tx.First(&stock, "product_id = ?", item.ProductID).Error; err != nil {
// 			log.Printf("[ERROR] Error fetching stock details for product_id: %d", item.ProductID)
// 			return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching stock details")
// 		}

// 		// Check stock availability
// 		if stock.Quantity < item.QuantitySold {
// 			log.Printf("[ERROR] Insufficient stock for product_id: %d. Available: %d, Requested: %d", item.ProductID, stock.Quantity, item.QuantitySold)
// 			return echo.NewHTTPError(http.StatusBadRequest, "Insufficient stock")
// 		}

// 		// Update total selling price for the transaction
// 		totalSellingPrice += float64(item.QuantitySold) * stock.SellingPrice

// 		// Update stock quantity
// 		stock.Quantity -= item.QuantitySold
// 		if err := tx.Save(&stock).Error; err != nil {
// 			log.Printf("[ERROR] Error updating stock for product_id: %d: %v", item.ProductID, err)
// 			return echo.NewHTTPError(http.StatusInternalServerError, "Error updating stock")
// 		}
// 	}

// 	// Handle payment mode: "cash" or "mpesa"
// 	var paymentMode string
// 	log.Printf("[INFO] Incoming payment mode from payload: %s", payload.PaymentMode)

// 	// Ensure the payment mode is set to either cash or mpesa
// 	if payload.PaymentMode == "cash" {
// 		paymentMode = "cash"
// 	} else if payload.PaymentMode == "Mpesa" {
// 		paymentMode = "Mpesa"
// 	} else {
// 		log.Printf("[ERROR] Invalid payment mode received: %s", payload.PaymentMode)
// 		return echo.NewHTTPError(http.StatusBadRequest, "Invalid payment mode")
// 	}

// 	log.Printf("[INFO] Setting payment mode: %s", paymentMode)

// 	// Process each item in the sale and create sale records
// 	for _, item := range payload.Items {
// 		var product models.Product
// 		if err := tx.First(&product, "product_id = ?", item.ProductID).Error; err != nil {
// 			if err == gorm.ErrRecordNotFound {
// 				log.Printf("[ERROR] Product not found for product_id: %d", item.ProductID)
// 				return echo.NewHTTPError(http.StatusNotFound, "Product not found")
// 			}
// 			log.Printf("[ERROR] Error fetching product details: %v", err)
// 			return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching product details")
// 		}

// 		var stock models.Stock
// 		if err := tx.First(&stock, "product_id = ?", item.ProductID).Error; err != nil {
// 			log.Printf("[ERROR] Error fetching stock details for product_id: %d", item.ProductID)
// 			return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching stock details")
// 		}

// 		totalCost := float64(item.QuantitySold) * stock.BuyingPrice
// 		profit := (float64(item.QuantitySold) * stock.SellingPrice) - totalCost
// 		balance := payload.CashReceived - totalSellingPrice

// 		// Sale object creation
// 		sale := models.Sale{
// 			SaleID:            saleID,         // Use the shortened sale_id generated above
// 			OrganizationsID:   organizationID, // Use OrganizationsID to associate with the correct organization
// 			Name:              product.ProductName,
// 			CategoryName:      product.CategoryName,
// 			UnitBuyingPrice:   stock.BuyingPrice,
// 			TotalBuyingPrice:  totalCost,
// 			UnitSellingPrice:  stock.SellingPrice,
// 			TotalSellingPrice: float64(item.QuantitySold) * stock.SellingPrice,
// 			Profit:            profit,
// 			Quantity:          item.QuantitySold,
// 			CashReceived:      payload.CashReceived,
// 			Balance:           balance,
// 			PaymentMode:       paymentMode,           // Ensure correct PaymentMode is used
// 			UserID:            int64(payload.UserID), // Convert int to int64 here
// 			Date:              time.Now(),
// 		}

// 		// Log the sale object before inserting into the database
// 		log.Printf("[INFO] Creating sale with payment mode: %s", paymentMode)
// 		log.Printf("[INFO] Sale Object: %+v", sale)

// 		// Insert the sale into the database
// 		if err := tx.Create(&sale).Error; err != nil {
// 			log.Printf("[ERROR] Error recording sale for product_id: %d: %v", item.ProductID, err)
// 			return echo.NewHTTPError(http.StatusInternalServerError, "Error recording sale")
// 		}

// 		// Log successful sale record
// 		log.Printf("[INFO] Sale recorded successfully for product_id: %d", item.ProductID)
// 	}

// 	// Commit the transaction
// 	if err := tx.Commit().Error; err != nil {
// 		log.Printf("[ERROR] Error committing transaction: %v", err)
// 		return echo.NewHTTPError(http.StatusInternalServerError, "Error committing transaction")
// 	}

// 	// Log successful transaction commit
// 	log.Println("[INFO] Transaction committed successfully")

// 	// Return success response
// 	return c.JSON(http.StatusOK, map[string]interface{}{
// 		"message": "Sale processed successfully for all items",
// 	})
//
