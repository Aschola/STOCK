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
	type SellProductRequest struct {
		ProductID    int    `json:"product_id"`
		QuantitySold int    `json:"quantity_sold"`
		UserID       string `json:"user_id"`
		Phone        string `json:"phone"`
	}

	var request SellProductRequest
	if err := c.Bind(&request); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}

	if request.Phone == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Phone number is required")
	}

	db := getDB()
	if db == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
	}

	var product models.Product
	if err := db.Table("active_products").First(&product, request.ProductID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return echo.NewHTTPError(http.StatusNotFound, "Product not found")
		}
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	}

	if product.Quantity < request.QuantitySold {
		return echo.NewHTTPError(http.StatusBadRequest, "Insufficient quantity")
	}

	totalCost := float64(request.QuantitySold) * product.SellingPrice
	profit := (product.SellingPrice - product.BuyingPrice) * float64(request.QuantitySold)

	// Call ProcessPayment with the userID and totalCost
	if err := payment.ProcessPayment(request.UserID, totalCost); err != nil {
		return echo.NewHTTPError(http.StatusPaymentRequired, "Payment processing failed: "+err.Error())
	}

	updatedQuantity := product.Quantity - request.QuantitySold

	if err := db.Table("active_products").Where("product_id = ?", request.ProductID).Update("quantity", updatedQuantity).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Internal Server Error")
	}

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

	combinedSale := models.CombinedSale{
		SaleID:            sale.SaleID,
		Name:              product.ProductName,
		Quantity:          request.QuantitySold,
		UnitBuyingPrice:   product.BuyingPrice,
		TotalBuyingPrice:  product.BuyingPrice * float64(request.QuantitySold),
		UnitSellingPrice:  product.SellingPrice,
		TotalSellingPrice: totalCost,
		Profit:            profit,
		CashReceive:       0.0,
		Balance:           0.0,
		UserID:            request.UserID,
		Date:              time.Now(),
		CategoryName:      product.CategoryName,
		SaleType:          "STKPUSH",
		TotalCost:         totalCost,
		ProductID:         request.ProductID,
	}

	if err := db.Table("combined_sales").Create(&combinedSale).Error; err != nil {
		log.Printf("Error inserting into combined_sales: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Error inserting into combined_sales")
	}

	log.Printf("Inserted into combined_sales: %+v", combinedSale)

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

	db := getDB()
	if db == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
	}

	var sales []models.Sale
	if err := db.Table("sales_by_STKPUSH").Find(&sales).Error; err != nil {
		log.Printf("Error querying sales from database: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
	}

	log.Printf("Fetched %d sales", len(sales))
	return c.JSON(http.StatusOK, sales)
}

// GetSaleByID fetches a single sale by its ID from the database.
func GetSaleByID(c echo.Context) error {
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

	var sale models.Sale
	if err := db.Table("sales_by_STKPUSH").First(&sale, saleID).Error; err != nil {
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

// AddSale adds a new sale record to the database.
func AddSale(c echo.Context) error {
	db := getDB()
	if db == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
	}

	var sale models.Sale
	if err := json.NewDecoder(c.Request().Body).Decode(&sale); err != nil {
		log.Printf("Error decoding JSON: %s", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, "Error decoding JSON")
	}

	log.Printf("Received request to create a sale: %+v", sale)

	if err := db.Table("sales_by_STKPUSH").Create(&sale).Error; err != nil {
		log.Printf("Error inserting sale: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Error inserting sale")
	}

	log.Printf("Created new sale with ID: %d", sale.SaleID)

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"message": "Sale created successfully",
		"sale_id": sale.SaleID,
	})
}

// UpdateSale updates an existing sale record in the database.
func UpdateSale(c echo.Context) error {
	saleID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("Invalid sale ID: %s", c.Param("id"))
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid sale ID")
	}

	db := getDB()
	if db == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
	}

	var sale models.Sale
	if err := json.NewDecoder(c.Request().Body).Decode(&sale); err != nil {
		log.Printf("Error decoding JSON: %s", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, "Error decoding JSON")
	}

	log.Printf("Received request to update sale ID %d: %+v", saleID, sale)

	if err := db.Table("sales_by_STKPUSH").Model(&sale).Where("sale_id = ?", saleID).Updates(sale).Error; err != nil {
		log.Printf("Error updating sale ID %d: %s", saleID, err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Error updating sale")
	}

	log.Printf("Updated sale ID %d successfully", saleID)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Sale updated successfully",
		"sale_id": saleID,
	})
}

// DeleteSale deletes a sale record from the database.
func DeleteSale(c echo.Context) error {
	saleID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		log.Printf("Invalid sale ID: %s", c.Param("id"))
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid sale ID")
	}

	db := getDB()
	if db == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
	}

	if err := db.Table("sales_by_STKPUSH").Delete(&models.Sale{}, saleID).Error; err != nil {
		log.Printf("Error deleting sale ID %d: %s", saleID, err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Error deleting sale")
	}

	log.Printf("Deleted sale ID %d successfully", saleID)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Sale deleted successfully",
		"sale_id": saleID,
	})
}
