package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	models "stock/models"
	"stock/payment"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// SellProduct processes the sale of a product.
func SellProduct(c echo.Context) error {
	// Define a struct to hold the request data
	type SellProductRequest struct {
		ProductID    int    `json:"product_id"`
		QuantitySold int    `json:"quantity_sold"`
		UserID       string `json:"user_id"`
		Phone        string `json:"phone"`
	}

	// Parse the request body into the struct
	var request SellProductRequest
	if err := c.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}

	// Validate phone number
	if request.Phone == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Phone number is required")
	}

	// Initialize database connection
	db := getDB()
	if db == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Retrieve the product details from active_products
	var product models.Product
	if err := db.Table("active_products").First(&product, request.ProductID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Product not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	}

	// Check if enough quantity is available
	if product.Quantity < request.QuantitySold {
		return echo.NewHTTPError(http.StatusBadRequest, "Insufficient quantity")
	}

	// Calculate total cost and profit
	totalCost := float64(request.QuantitySold) * product.SellingPrice
	profit := (product.SellingPrice - product.BuyingPrice) * float64(request.QuantitySold)

	// Call the payment process before inserting the sale record
	if err := payment.ProcessPayment(request.UserID, totalCost, request.Phone); err != nil {
		return echo.NewHTTPError(http.StatusPaymentRequired, "Payment processing failed: "+err.Error())
	}

	// Prepare the updated quantity
	updatedQuantity := product.Quantity - request.QuantitySold

	// Update the quantity of the product in the 'active_products' table
	if err := db.Table("active_products").Where("product_id = ?", request.ProductID).Update("quantity", updatedQuantity).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	}

	// Create sale record
	sale := models.Sale{
		Name:              product.ProductName,
		UnitSellingPrice:  product.SellingPrice,
		Quantity:          request.QuantitySold,
		UserID:            request.UserID,
		Date:              time.Now(),
		CategoryName:      product.CategoryName,
		TotalSellingPrice: totalCost,
		UnitBuyingPrice:   product.BuyingPrice,
		TotalBuyingPrice:  product.BuyingPrice * float64(request.QuantitySold),
		Profit:            profit,
	}

	if err := db.Table("sales_by_STKPUSH").Create(&sale).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	}

	// Return success response with total cost
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message":       "Sale processed successfully",
		"product_id":    request.ProductID,
		"quantity_sold": request.QuantitySold,
		"remaining_qty": updatedQuantity,
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

	// Query all sales from the sales_by_STKPUSH table
	var sales []models.Sale
	if err := db.Table("sales_by_STKPUSH").Find(&sales).Error; err != nil {
		log.Printf("Error querying sales from database: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
	}

	// Log the number of sales fetched
	log.Printf("Fetched %d sales", len(sales))

	// Return the fetched sales as JSON
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
	if err := db.Table("sales_by_STKPUSH").First(&sale, saleID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("Sale not found with ID: %d", saleID)
			return echo.NewHTTPError(http.StatusNotFound, "Sale not found")
		}
		log.Printf("Error querying sale: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch sale")
	}

	// Log the fetched sale
	log.Printf("Fetched sale: %+v", sale)

	// Return the fetched sale as JSON
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
	if err := db.Table("sales_by_STKPUSH").Create(&sale).Error; err != nil {
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

	// Execute the SQL UPDATE query to update the sale in the database
	if err := db.Table("sales_by_STKPUSH").Model(&sale).Where("sale_id = ?", saleID).Updates(sale).Error; err != nil {
		log.Printf("Error updating sale ID %d: %s", saleID, err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Error updating sale")
	}

	// Log the successful update
	log.Printf("Updated sale ID %d successfully", saleID)

	// Return success response
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Sale updated successfully",
		"sale_id": saleID,
	})
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

	// Execute the SQL DELETE query to remove the sale from the database
	if err := db.Table("sales_by_STKPUSH").Delete(&models.Sale{}, saleID).Error; err != nil {
		log.Printf("Error deleting sale ID %d: %s", saleID, err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Error deleting sale")
	}

	// Log the successful deletion
	log.Printf("Deleted sale ID %d successfully", saleID)

	// Return success response
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Sale deleted successfully",
		"sale_id": saleID,
	})
}
