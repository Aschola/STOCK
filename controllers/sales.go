package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"stock/models"
	"time"
	"fmt"
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

// 	saleID := time.Now().Unix()
// 	var totalSellingPrice float64

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

// 		totalSellingPrice += float64(item.QuantitySold) * stock.SellingPrice
// 	}

// 	log.Printf("[INFO] Payment Mode: %s", payload.PaymentMode)
// 	if payload.PaymentMode == "cash" {
// 		// Process cash payment immediately
// 		for _, item := range payload.Items {
// 			var stock models.Stock
// 			if err := tx.First(&stock, "product_id = ?", item.ProductID).Error; err != nil {
// 				log.Printf("[ERROR] Error fetching stock details for product_id: %d", item.ProductID)
// 				return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching stock details")
// 			}

// 			if stock.Quantity < item.QuantitySold {
// 				log.Printf("[ERROR] Insufficient stock for product_id: %d", item.ProductID)
// 				return echo.NewHTTPError(http.StatusBadRequest, "Insufficient stock")
// 			}

// 			stock.Quantity -= item.QuantitySold
// 			if err := tx.Save(&stock).Error; err != nil {
// 				log.Printf("[ERROR] Error updating stock for product_id: %d: %v", item.ProductID, err)
// 				return echo.NewHTTPError(http.StatusInternalServerError, "Error updating stock")
// 			}

// 			var product models.Product
// 			if err := tx.First(&product, "product_id = ?", item.ProductID).Error; err != nil {
// 				log.Printf("[ERROR] Error fetching product details: %v", err)
// 				return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching product details")
// 			}

// 			totalCost := float64(item.QuantitySold) * stock.BuyingPrice
// 			profit := (float64(item.QuantitySold) * stock.SellingPrice) - totalCost
// 			balance := payload.CashReceived - totalSellingPrice

// 			sale := models.Sale{
// 				SaleID:            saleID,
// 				OrganizationsID:   organizationID,
// 				Name:              product.ProductName,
// 				CategoryName:      product.CategoryName,
// 				UnitBuyingPrice:   stock.BuyingPrice,
// 				TotalBuyingPrice:  totalCost,
// 				UnitSellingPrice:  stock.SellingPrice,
// 				TotalSellingPrice: int64(float64(item.QuantitySold) * stock.SellingPrice),
// 				Profit:            profit,
// 				Quantity:          item.QuantitySold,
// 				CashReceived:      payload.CashReceived,
// 				Balance:           balance,
// 				PaymentMode:       "cash",
// 				UserID:            int64(payload.UserID),
// 				Date:              time.Now(),
// 				TransactionStatus: "COMPLETE",
// 			}

// 			if err := tx.Create(&sale).Error; err != nil {
// 				log.Printf("[ERROR] Error recording sale for product_id: %d: %v", item.ProductID, err)
// 				return echo.NewHTTPError(http.StatusInternalServerError, "Error recording sale")
// 			}
// 		}

// 		if err := tx.Commit().Error; err != nil {
// 			log.Printf("[ERROR] Error committing transaction: %v", err)
// 			return echo.NewHTTPError(http.StatusInternalServerError, "Error committing transaction")
// 		}

// 		return c.JSON(http.StatusOK, map[string]interface{}{
// 			"message": "Sale processed successfully for all items",
// 		})

// 	} else if payload.PaymentMode == "Mpesa" {
// 		stkRequest := STKPushRequest{
// 			PhoneNumber: payload.PhoneNumber,
// 			Amount:      totalSellingPrice,
// 		}

// 		result, err := InitiateSTKPush(int64(organizationID), stkRequest)
// 		if err != nil {
// 			log.Printf("[ERROR] Mpesa STK Push failed: %v", err)
// 			return echo.NewHTTPError(http.StatusInternalServerError, "Error initiating Mpesa payment")
// 		}

// 		for _, item := range payload.Items {
// 			var stock models.Stock
// 			var product models.Product
// 			if err := tx.First(&product, "product_id = ?", item.ProductID).Error; err != nil {
// 				log.Printf("[ERROR] Error fetching product details: %v", err)
// 				return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching product details")
// 			}

// 			totalCost := float64(item.QuantitySold) * stock.BuyingPrice
// 			profit := (float64(item.QuantitySold) * stock.SellingPrice) - totalCost
// 			balance := payload.CashReceived - totalSellingPrice

// 			sale := models.Sale{
// 				SaleID:            saleID,
// 				OrganizationsID:   organizationID,
// 				Name:              product.ProductName,
// 				CategoryName:      product.CategoryName,
// 				UnitBuyingPrice:   stock.BuyingPrice,
// 				TotalBuyingPrice:  totalCost,
// 				UnitSellingPrice:  stock.SellingPrice,
// 				TotalSellingPrice: int64(float64(item.QuantitySold) * stock.SellingPrice),
// 				Profit:            profit,
// 				Quantity:          item.QuantitySold,
// 				CashReceived:      payload.CashReceived,
// 				Balance:           balance,
// 				PaymentMode:       "Mpesa",
// 				UserID:            int64(payload.UserID),
// 				Date:              time.Now(),
// 				TransactionID:     result.TransactionID,
// 				TransactionStatus: "PENDING",
// 			}

// 			if err := tx.Create(&sale).Error; err != nil {
// 				log.Printf("[ERROR] Error recording sale for product_id: %d: %v", item.ProductID, err)
// 				return echo.NewHTTPError(http.StatusInternalServerError, "Error recording sale")
// 			}
// 		}

// 		if err := tx.Commit().Error; err != nil {
// 			log.Printf("[ERROR] Error committing transaction: %v", err)
// 			return echo.NewHTTPError(http.StatusInternalServerError, "Error committing transaction")
// 		}

// 		return c.JSON(http.StatusAccepted, map[string]interface{}{
// 			"message": "Sale recorded and awaiting Mpesa payment confirmation",
// 			"transactionId": result.TransactionID,
// 		})

// 	} else {
// 		log.Printf("[ERROR] Invalid payment mode received: %s", payload.PaymentMode)
// 		return echo.NewHTTPError(http.StatusBadRequest, "Invalid payment mode")
// 	}
// }
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

	// Validate products and calculate total price
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

		if stock.Quantity < item.QuantitySold {
			log.Printf("[ERROR] Insufficient stock for product_id: %d", item.ProductID)
			return echo.NewHTTPError(http.StatusBadRequest, "Insufficient stock")
		}

		totalSellingPrice += float64(item.QuantitySold) * stock.SellingPrice
		items = append(items, item)
	}

	log.Printf("[INFO] Payment Mode: %s", payload.PaymentMode)
	if payload.PaymentMode == "cash" {
		// Process cash payment immediately
		if err := processCashSale(tx, items, saleID, organizationID, totalSellingPrice, payload); err != nil {
			return err
		}

		if err := tx.Commit().Error; err != nil {
			log.Printf("[ERROR] Error committing transaction: %v", err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Error committing transaction")
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"message": "Sale processed successfully for all items",
		})

	} else if payload.PaymentMode == "Mpesa" {
		// Initiate Mpesa payment
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
			PaymentMode:       "cash",
			UserID:            int64(payload.UserID),
			Date:              time.Now(),
		}

		if err := tx.Create(&sale).Error; err != nil {
			log.Printf("[ERROR] Error recording sale for product_id: %d: %v", item.ProductID, err)
			return echo.NewHTTPError(http.StatusInternalServerError, "Error recording sale")
			log.Printf("[DEBUG] Received Sale Request: %+v", payload)

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
			SaleID:            saleID,
			OrganizationsID:   organizationID,
			Name:              product.ProductName,
			CategoryName:      product.CategoryName,
			ProductID:         item.ProductID,
			UnitBuyingPrice:   stock.BuyingPrice,
			TotalBuyingPrice:  totalCost,
			UnitSellingPrice:  stock.SellingPrice,
			TotalSellingPrice: int64(float64(item.QuantitySold) * stock.SellingPrice),
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

