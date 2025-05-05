package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"stock/models"
	"time"

	//"stock/db"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Helper function to log errors
func logError(message string, err error) {
	if err != nil {
		log.Printf("[ERROR] %s: %v", message, err)
	}
}

func getUserID(c echo.Context) (uint, error) {
	userID, ok := c.Get("userID").(uint)
	if !ok {
		return 0, fmt.Errorf("user ID not found in context")
	}
	return userID, nil
}

// func SellProduct(c echo.Context) error {
// 	log.Println("[INFO] Received request to sell products.")

// 	organizationID, err := getOrganizationID(c)
// 	if err != nil {
// 		return err
// 	}

// 	var payload models.SalePayload
// 	if err := json.NewDecoder(c.Request().Body).Decode(&payload); err != nil {
// 		log.Printf("[ERROR] Error parsing request body: %v", err)
// 		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input data")
// 	}

// 	db := getDB()
// 	if db == nil {
// 		log.Println("[ERROR] Database connection failed")
// 		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
// 	}

// 	tx := db.Begin()
// 	defer func() {
// 		if r := recover(); r != nil {
// 			tx.Rollback()
// 			log.Printf("[ERROR] Unexpected error: %v", r)
// 		}
// 	}()
// 	defer tx.Rollback()

// 	var totalSellingPrice float64
// 	var items []models.SaleItem
// 	saleID := time.Now().Unix()

// 	// Validate products and calculate total price
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

// 		if stock.Quantity < item.QuantitySold {
// 			log.Printf("[ERROR] Insufficient stock for product_id: %d", item.ProductID)
// 			return echo.NewHTTPError(http.StatusBadRequest, "Insufficient stock")
// 		}

// 		totalSellingPrice += float64(item.QuantitySold) * stock.SellingPrice
// 		items = append(items, item)
// 	}

// 	log.Printf("[INFO] Payment Mode: %s", payload.PaymentMode)
// 	if payload.PaymentMode == "cash" {
// 		// Process cash payment immediately
// 		if err := processCashSale(tx, items, saleID, organizationID, totalSellingPrice, payload); err != nil {
// 			return err
// 		}

// 		if err := tx.Commit().Error; err != nil {
// 			log.Printf("[ERROR] Error committing transaction: %v", err)
// 			return echo.NewHTTPError(http.StatusInternalServerError, "Error committing transaction")
// 		}

// 		var company models.CompanySetting
//         if err := db.Where("organization_id = ?", organizationID).First(&company).Error; err != nil {
//             log.Printf("[ERROR] Failed to fetch company settings: %v", err)
//         }

//         // Print using the payload and company settings
//         if err := PrintReceipt(payload, company); err != nil {
//             log.Printf("[ERROR] Failed to print receipt: %v", err)
//         }

// 		return c.JSON(http.StatusOK, map[string]interface{}{
// 			"message": "Sale processed successfully for all items",
// 		})

// 	} else if payload.PaymentMode == "Mpesa" {
// 		// Initiate Mpesa payment
// 		stkRequest := STKPushRequest{
// 			PhoneNumber: payload.PhoneNumber,
// 			Amount:      totalSellingPrice,
// 		}

// 		result, err := InitiateSTKPush(int64(organizationID), stkRequest)
// 		if err != nil {
// 			log.Printf("[ERROR] Mpesa STK Push failed: %v", err)
// 			return echo.NewHTTPError(http.StatusInternalServerError, "Error initiating Mpesa payment")
// 		}

// 		// Create pending sales records without reducing stock
// 		if err := createPendingMpesaSales(tx, items, saleID, organizationID, totalSellingPrice, payload, result.TransactionID); err != nil {
// 			return err
// 		}

// 		if err := tx.Commit().Error; err != nil {
// 			log.Printf("[ERROR] Error committing transaction: %v", err)
// 			return echo.NewHTTPError(http.StatusInternalServerError, "Error committing transaction")
// 		}

// 		return c.JSON(http.StatusAccepted, map[string]interface{}{
// 			"message":       "Payment requested. Awaiting confirmation",
// 			"transactionId": result.TransactionID,
// 		})

// 	} else {
// 		log.Printf("[ERROR] Invalid payment mode received: %s", payload.PaymentMode)
// 		return echo.NewHTTPError(http.StatusBadRequest, "Invalid payment mode")
// 	}
// }
// Update the SellProduct function in the controllers package
// This updated version properly passes all sale details to the receipt printer

// Updated SellProduct function in the controllers package
// This version keeps all function names the same but updates how data is passed

func SellProduct(c echo.Context) error {
	log.Println("[INFO] Received request to sell products.")

	organizationID, err := getOrganizationID(c)
	if err != nil {
		return err
	}

	var payload models.SalePayload
	if err := json.NewDecoder(c.Request().Body).Decode(&payload); err != nil {
		log.Printf("[ERROR] Error parsing request body: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input data")
	}

	// Debug the incoming payload
	log.Printf("[DEBUG] Received sale payload with %d items", len(payload.Items))

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

	var totalSellingPrice float64
	var items []models.SaleItem
	saleID := time.Now().Unix()

	// Get the current user ID from the context
	userID, err := getUserID(c)
	payload.UserID = int(userID)

	// Fetch product details and populate items with full information
	for i, item := range payload.Items {
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

		if stock.Quantity < item.QuantitySold {
			log.Printf("[ERROR] Insufficient stock for product_id: %d", item.ProductID)
			return echo.NewHTTPError(http.StatusBadRequest, "Insufficient stock")
		}

		// Update the name and unit price in the payload item with full details
		payload.Items[i].Name = product.ProductName
		payload.Items[i].UnitPrice = stock.SellingPrice

		totalSellingPrice += float64(item.QuantitySold) * stock.SellingPrice
		items = append(items, payload.Items[i])
	}

	// Make sure payload has up-to-date information
	payload.Items = items
	
	// Set CashReceived to match total if not provided
	if payload.CashReceived <= 0 {
		payload.CashReceived = totalSellingPrice
	}

	log.Printf("[INFO] Payment Mode: %s", payload.PaymentMode)
	log.Printf("[DEBUG] Updated sale payload now has %d items with complete details", len(payload.Items))
	
	if payload.PaymentMode == "cash" {
		// Process cash payment immediately
		if err := processCashSale(tx, items, saleID, organizationID, totalSellingPrice, payload); err != nil {
			return err
		}

		if err := tx.Commit().Error; err != nil {
			log.Printf("[ERROR] Error committing transaction: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Error committing transaction")
		}

		// Fetch company details for receipt
		var company models.CompanySetting
		if err := db.Where("organization_id = ?", organizationID).First(&company).Error; err != nil {
			log.Printf("[ERROR] Failed to fetch company settings: %v", err)
		}

		// Get username instead of user ID for the receipt
		var user models.User
		if err := db.First(&user, payload.UserID).Error; err != nil {
			log.Printf("[WARN] Failed to fetch user details: %v, using ID instead", err)
		} else {
			// If we found the user, update the username in the payload
			payload.UserName = user.Username
		}

		// Print using the fully populated payload and company settings
		log.Printf("[DEBUG] Calling PrintReceipt with %d items", len(payload.Items))
		for i, item := range payload.Items {
			log.Printf("[DEBUG] Item %d: %s, Qty: %d, Price: %.2f", 
				i, item.Name, item.QuantitySold, item.UnitPrice)
		}
		
		if err := PrintReceipt(payload, company); err != nil {
			log.Printf("[ERROR] Failed to print receipt: %v", err)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Sale processed successfully for all items",
		})

	} else if payload.PaymentMode == "mpesa" {
		// Initiate Mpesa payment (implementation details not included)
		stkRequest := STKPushRequest{
			PhoneNumber: payload.PhoneNumber,
			Amount:      totalSellingPrice,
		}

		result, err := InitiateSTKPush(int64(organizationID), stkRequest)
		if err != nil {
			log.Printf("[ERROR] Mpesa STK Push failed: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Error initiating Mpesa payment")
		}

		// Create pending sales records without reducing stock
		if err := createPendingMpesaSales(tx, items, saleID, organizationID, totalSellingPrice, payload, result.TransactionID); err != nil {
			return err
		}

		if err := tx.Commit().Error; err != nil {
			log.Printf("[ERROR] Error committing transaction: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Error committing transaction")
		}

		return c.JSON(http.StatusAccepted, map[string]interface{}{
			"message":       "Payment requested. Awaiting confirmation",
			"transactionId": result.TransactionID,
		})

	} else {
		log.Printf("[ERROR] Invalid payment mode received: %s", payload.PaymentMode)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid payment mode")
	}
}

// Helper function to process cash sales
func processCashSale(tx *gorm.DB, items []models.SaleItem, saleID int64, organizationID uint, totalSellingPrice float64, payload models.SalePayload) error {
	for _, item := range items {
		var stock models.Stock
		if err := tx.First(&stock, "product_id = ?", item.ProductID).Error; err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching stock details")
		}

		stock.Quantity -= item.QuantitySold
		if err := tx.Save(&stock).Error; err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Error updating stock")
		}

		var product models.Product
		if err := tx.First(&product, "product_id = ?", item.ProductID).Error; err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching product details")
		}

		totalCost := float64(item.QuantitySold) * stock.BuyingPrice
		profit := (float64(item.QuantitySold) * stock.SellingPrice) - totalCost
		balance := payload.CashReceived - totalSellingPrice

		sale := models.Sale{
			SaleID:           saleID,
			OrganizationsID:  organizationID,
			Name:             product.ProductName,
			CategoryName:     product.CategoryName,
			UnitBuyingPrice:  stock.BuyingPrice,
			TotalBuyingPrice: totalCost,
			UnitSellingPrice: stock.SellingPrice,
			//TotalSellingPrice: int64(float64(item.QuantitySold) * stock.SellingPrice),
			TotalSellingPrice: float64(item.QuantitySold) * stock.SellingPrice,
			Profit:            profit,
			Quantity:          item.QuantitySold,
			CashReceived:      payload.CashReceived,
			Balance:           balance,
			PaymentMode:       "cash",
			UserID:            int64(payload.UserID),
			Date:              time.Now(),
		}

		if err := tx.Create(&sale).Error; err != nil {
			log.Printf("[ERROR] Error recording sale for product_id: %d: %v", item.ProductID, err)
			log.Printf("[DEBUG] Received Sale Request: %+v", payload)
			return echo.NewHTTPError(http.StatusInternalServerError, "Error recording sale")

		}
	}
	return nil
}

// Helper function to create pending Mpesa sales
func createPendingMpesaSales(tx *gorm.DB, items []models.SaleItem, saleID int64, organizationID uint, totalSellingPrice float64, payload models.SalePayload, transactionID string) error {
	for _, item := range items {
		var product models.Product
		if err := tx.First(&product, "product_id = ?", item.ProductID).Error; err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching product details")
		}

		var stock models.Stock
		if err := tx.First(&stock, "product_id = ?", item.ProductID).Error; err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching stock details")
		}

		totalCost := float64(item.QuantitySold) * stock.BuyingPrice
		profit := (float64(item.QuantitySold) * stock.SellingPrice) - totalCost
		balance := payload.CashReceived - totalSellingPrice

		sale := models.Sale{
			SaleID:           saleID,
			OrganizationsID:  organizationID,
			Name:             product.ProductName,
			CategoryName:     product.CategoryName,
			ProductID:        item.ProductID,
			UnitBuyingPrice:  stock.BuyingPrice,
			TotalBuyingPrice: totalCost,
			UnitSellingPrice: stock.SellingPrice,
			//TotalSellingPrice: int64(float64(item.QuantitySold) * stock.SellingPrice),
			TotalSellingPrice: float64(item.QuantitySold) * stock.SellingPrice,
			Profit:            profit,
			Quantity:          item.QuantitySold,
			CashReceived:      payload.CashReceived,
			Balance:           balance,
			PaymentMode:       "Mpesa",
			UserID:            int64(payload.UserID),
			Date:              time.Now(),
			TransactionID:     transactionID,
			TransactionStatus: "PENDING",
		}

		if err := tx.Create(&sale).Error; err != nil {
			log.Printf("[ERROR] Error recording sale for product_id: %d: %v", item.ProductID, err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Error recording sale")
		}
	}
	return nil
}

// New function to handle Mpesa callback and update stock
func UpdateMpesaTransactionStatus(transactionID string, newStatus string) error {
	db := getDB()
	if db == nil {
		return fmt.Errorf("database connection failed")
	}

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	defer tx.Rollback()

	// Update transaction status in the sales table
	var sales []models.Sale
	if err := tx.Where("transaction_id = ?", transactionID).Find(&sales).Error; err != nil {
		return fmt.Errorf("error fetching sales: %v", err)
	}

	for _, sale := range sales {
		sale.TransactionStatus = newStatus
		if err := tx.Save(&sale).Error; err != nil {
			return fmt.Errorf("error updating sale status: %v", err)
		}

		// If transaction is complete, update stock
		if sale.TransactionStatus == "COMPLETED" {
			var stock models.Stock
			if err := tx.First(&stock, "product_id = ?", sale.ProductID).Error; err != nil {
				return fmt.Errorf("error fetching stock: %v", err)
			}

			stock.Quantity -= sale.Quantity
			if err := tx.Save(&stock).Error; err != nil {
				return fmt.Errorf("error updating stock: %v", err)
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("error committing transaction: %v", err)
	}

	return nil
}
