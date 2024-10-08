package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	models "stock/models"
	"strconv"
	"time"

	"stock/payment"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Get the database instance
//
//	func getDB() *gorm.DB {
//	    db := database.GetDB()
//	    if db == nil {
//	        return nil
//	    }
//	    return db
//	}
func SellProduct(c echo.Context) error {
	productIDStr := c.Param("product_id")
	quantitySoldStr := c.Param("quantity_sold")
	userID := c.QueryParam("user_id")

	// Convert parameters to integer values
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		log.Printf("Invalid product ID: %s", productIDStr)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid product ID")
	}

	quantitySold, err := strconv.Atoi(quantitySoldStr)
	if err != nil {
		log.Printf("Invalid quantity sold: %s", quantitySoldStr)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid quantity sold")
	}

	// Initialize database connection
	db := getDB()
	if db == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Start a transaction
	tx := db.Begin()
	if tx.Error != nil {
		log.Printf("Error starting transaction: %s", tx.Error.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	}
	defer tx.Rollback()

	// Retrieve the product details
	var product models.Product
	if err := tx.First(&product, productID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("Product not found with ID: %d", productID)
			return echo.NewHTTPError(http.StatusNotFound, "Product not found")
		}
		log.Printf("Error querying product details: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	}

	// Check if enough quantity is available
	if product.Quantity < quantitySold {
		log.Printf("Insufficient quantity for product ID %d: Available %d, Requested %d", productID, product.Quantity, quantitySold)
		return echo.NewHTTPError(http.StatusBadRequest, "Insufficient quantity")
	}

	// Prepare the updated quantity
	updatedQuantity := product.Quantity - quantitySold

	// Calculate total cost and profit
	totalCost := float64(quantitySold) * product.SellingPrice
	profit := (product.SellingPrice - product.BuyingPrice) * float64(quantitySold)

	// Call the payment process before inserting the sale record
	if err := payment.ProcessPayment(userID, totalCost); err != nil {
		log.Printf("Payment processing failed: %s", err.Error())
		return echo.NewHTTPError(http.StatusPaymentRequired, "Payment processing failed")
	}

	// Update the quantity of the product in the 'products' table
	if err := tx.Model(&models.Product{}).Where("product_id = ?", productID).Update("quantity", updatedQuantity).Error; err != nil {
		log.Printf("Error updating product quantity: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	}

	// Create sale record
	sale := models.Sale{
		Name:             product.ProductName,
		UnitPrice:        product.SellingPrice,
		Quantity:         quantitySold,
		UserID:           userID,
		Date:             time.Now(),
		CategoryName:     product.CategoryName,
		TotalCost:        totalCost,
		UnitBuyingPrice:  product.BuyingPrice,
		TotalBuyingPrice: product.BuyingPrice * float64(quantitySold),
		Profit:           profit,
	}

	if err := tx.Create(&sale).Error; err != nil {
		log.Printf("Error inserting sale record: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("Error committing transaction: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	}

	// Log successful sale
	log.Printf("Sold %d units of product ID %d. Remaining quantity: %d", quantitySold, productID, updatedQuantity)

	// Return success response with total cost
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":       "Sale processed successfully",
		"product_id":    strconv.Itoa(productID),
		"quantity_sold": strconv.Itoa(quantitySold),
		"remaining_qty": strconv.Itoa(updatedQuantity),
		"total_cost":    totalCost,
	})
}

// GetSales fetches all sales from the database.
func GetSales(c echo.Context) error {
	log.Println("Received request to fetch sales")

	// Initialize database connection
	db := getDB()
	if db == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Query all sales from the Sales table
	var sales []models.Sale
	if err := db.Find(&sales).Error; err != nil {
		log.Printf("Error querying sales from database: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
	}

	// Log the number of Sales fetched
	log.Printf("Fetched %d sales", len(sales))

	// Return the fetched Sales as JSON
	return c.JSON(http.StatusOK, sales)
}

// GetSaleByID fetches a single sale by its ID from the database.
func GetSaleByID(c echo.Context) error {
	log.Println("Received request to fetch sale by ID")

	// Extract sale ID from path parameter
	saleID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("Invalid sale ID: %s", c.Param("id"))
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid sale ID")
	}

	// Initialize database connection
	db := getDB()
	if db == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Query the sale by sale ID
	var sale models.Sale
	if err := db.First(&sale, saleID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("Sale not found with ID: %d", saleID)
			return echo.NewHTTPError(http.StatusNotFound, "Sale not found")
		}
		log.Printf("Error querying sale: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch sale")
	}

	// Log the fetched sale
	log.Printf("Fetched sale: %+v", sale)

	// Return the fetched Sale as JSON
	return c.JSON(http.StatusOK, sale)
}

// AddSale adds a new sale record to the database.
func AddSale(c echo.Context) error {
	// Initialize database connection
	db := getDB()
	if db == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Parse JSON manually from request body into Sale struct
	var sale models.Sale
	if err := json.NewDecoder(c.Request().Body).Decode(&sale); err != nil {
		log.Printf("Error decoding JSON: %s", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, "Error decoding JSON")
	}

	// Log the received sale details
	log.Printf("Received request to create a sale: %+v", sale)

	// Execute the SQL INSERT query to add the sale to the database
	if err := db.Create(&sale).Error; err != nil {
		log.Printf("Error inserting sale: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Error inserting sale")
	}

	// Log the creation of the new sale
	log.Printf("Created new sale with ID: %d", sale.SaleID)

	// Return success response
	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "Sale created successfully",
		"sale_id": sale.SaleID,
	})
}

// UpdateSale updates an existing sale record in the database.
func UpdateSale(c echo.Context) error {
	// Extract sale ID from path parameter
	saleID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("Invalid sale ID: %s", c.Param("id"))
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid sale ID")
	}

	// Initialize database connection
	db := getDB()
	if db == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Parse JSON from request body into Sale struct
	var sale models.Sale
	if err := json.NewDecoder(c.Request().Body).Decode(&sale); err != nil {
		log.Printf("Error decoding JSON: %s", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, "Error decoding JSON")
	}

	// Log the update details
	log.Printf("Received request to update sale ID %d: %+v", saleID, sale)

	// Execute SQL UPDATE query to modify the sale in the database
	if err := db.Model(&models.Sale{}).Where("sale_id = ?", saleID).Updates(sale).Error; err != nil {
		log.Printf("Error updating sale: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Error updating sale")
	}

	// Log the update success
	log.Printf("Updated sale with ID: %d", saleID)

	// Return success response
	return c.JSON(http.StatusOK, map[string]string{"message": "Sale updated successfully"})
}

// DeleteSale deletes a sale record from the database.
func DeleteSale(c echo.Context) error {
	// Extract sale ID from path parameter
	saleID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("Invalid sale ID: %s", c.Param("id"))
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid sale ID")
	}

	// Initialize database connection
	db := getDB()
	if db == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Execute SQL DELETE query to remove the sale from the database
	if err := db.Delete(&models.Sale{}, saleID).Error; err != nil {
		log.Printf("Error deleting sale: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Error deleting sale")
	}

	// Log the deletion success
	log.Printf("Deleted sale with ID: %d", saleID)

	// Return success response
	return c.JSON(http.StatusOK, map[string]string{"message": "Sale deleted successfully"})
}
