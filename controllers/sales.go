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

	// Verify the phone number has been parsed correctly
	log.Printf("[INFO] Parsed phone number: %d", payload.PhoneNumber)

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

	// Determine payment mode
	// log.Printf("[INFO] Payment Mode: %s", payload.PaymentMode)
	// var paymentMode string
	// if payload.PaymentMode == "cash" {
	// 	paymentMode = "cash"
	// } else if payload.PaymentMode == "Mpesa" {
	// 	paymentMode = "Mpesa"

	// 	// If payment mode is Mpesa, initiate STK Push before committing the sale
	// 	log.Printf("[INFO] Initiating Mpesa STK Push for KES %.2f to phone number: %d", totalSellingPrice, payload.PhoneNumber)

	// 	// Create the STK Push request
	// 	stkRequest := STKPushRequest{
	// 		PhoneNumber: payload.PhoneNumber,

	// 	}

	// 	// Initiate the STK Push
	// 	result, err := InitiateSTKPush(int64(totalSellingPrice), stkRequest)
	// 	if err != nil {
	// 		log.Printf("[ERROR] Mpesa STK Push failed: %v", err)
	// 		return echo.NewHTTPError(http.StatusInternalServerError, "Error initiating Mpesa payment")
	// 	}

	// 	// You might want to store the result for future reference
	// 	log.Printf("[INFO] STK Push initiated successfully: %+v", result)
	// }
	// Determine payment mode
log.Printf("[INFO] Payment Mode: %s", payload.PaymentMode)
var paymentMode string
if payload.PaymentMode == "cash" {
	paymentMode = "cash"
} else if payload.PaymentMode == "Mpesa" {
	paymentMode = "Mpesa"

	// If payment mode is Mpesa, initiate STK Push before committing the sale
	log.Printf("[INFO] Initiating Mpesa STK Push for KES %.2f to phone number: %d", totalSellingPrice, payload.PhoneNumber)

	// Create the STK Push request
	stkRequest := STKPushRequest{
		PhoneNumber: payload.PhoneNumber,
		Amount: (totalSellingPrice),
		
	}

	// Initiate the STK Push
	result, err := InitiateSTKPush(int64(organizationID), stkRequest)
	//result, err := InitiateSTKPush(int64(totalSellingPrice), stkRequest)
	if err != nil {
		log.Printf("[ERROR] Mpesa STK Push failed: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error initiating Mpesa payment")
	}

	log.Printf("[INFO] Mpesa STK Push successful. Response: %+v", result)
} else {
	log.Printf("[ERROR] Invalid payment mode received: %s", payload.PaymentMode)
	return echo.NewHTTPError(http.StatusBadRequest, "Invalid payment mode")
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
			PaymentMode:       paymentMode,
			UserID:            int64(payload.UserID),
			Date:              time.Now(),
		}

		// Log the sale creation info
		log.Printf("[INFO] Creating sale with payment mode: %s", paymentMode)
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
		"message": "Sale processed successfully for all items",
	})
}
