package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"stock/db"
	models "stock/models"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Get the database instance
func getDB() *gorm.DB {
	db := db.GetDB()
	if db == nil {
		log.Println("Failed to get database instance")
	}
	return db
}

// Utility function for error responses
func errorResponse(c echo.Context, statusCode int, message string) error {
	log.Println(message)
	return echo.NewHTTPError(statusCode, message)
}

// GetProducts fetches all products
func GetProducts(c echo.Context) error {
	log.Println("Fetching all products")
	db := getDB()
	if db == nil {
		return errorResponse(c, http.StatusInternalServerError, "Failed to connect to the database")
	}

	var products []models.Product
	if err := db.Table("active_products").Find(&products).Error; err != nil {
		return errorResponse(c, http.StatusInternalServerError, "Failed to fetch products")
	}

	log.Printf("Successfully fetched %d products", len(products))
	return c.JSON(http.StatusOK, products)
}

// GetProductByID fetches a product by its ID
func GetProductByID(c echo.Context) error {
	productID, err := strconv.Atoi(c.Param("product_id"))
	if err != nil {
		log.Printf("Error converting product ID: %v", err)
		return errorResponse(c, http.StatusBadRequest, "Invalid product ID")
	}

	log.Printf("Fetching product with ID: %d", productID)
	db := getDB()
	if db == nil {
		log.Println("Failed to connect to the database")
		return errorResponse(c, http.StatusInternalServerError, "Failed to connect to the database")
	}

	var prod models.Product
	if err := db.Table("active_products").Where("product_id = ?", productID).First(&prod).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("Product not found: %d", productID)
			return errorResponse(c, http.StatusNotFound, "Product not found")
		}
		log.Printf("Error fetching product: %v", err)
		return errorResponse(c, http.StatusInternalServerError, "Failed to fetch product")
	}

	// Create a response struct
	response := struct {
		CategoryName       string  `json:"category_name"`
		ProductName        string  `json:"product_name"`
		ProductCode        string  `json:"product_code"`
		ProductDescription string  `json:"product_description"`
		Quantity           int     `json:"quantity"`
		ReorderLevel       int     `json:"reorder_level"`
		BuyingPrice        float64 `json:"buying_price"`
		SellingPrice       float64 `json:"selling_price"`
		ProductID          int     `json:"product_id"`
	}{
		CategoryName:       prod.CategoryName,
		ProductName:        prod.ProductName,
		ProductCode:        prod.ProductCode,
		ProductDescription: prod.ProductDescription,
		Quantity:           prod.Quantity,
		ReorderLevel:       prod.ReorderLevel,
		BuyingPrice:        prod.BuyingPrice,
		SellingPrice:       prod.SellingPrice,
		ProductID:          prod.ProductID,
	}

	log.Printf("Successfully fetched product: %+v", response)
	return c.JSON(http.StatusOK, response)
}

// AddProduct adds a new product with the current date
func AddProduct(c echo.Context) error {
	log.Println("Adding a new product")
	db := getDB()
	if db == nil {
		return errorResponse(c, http.StatusInternalServerError, "Failed to connect to the database")
	}

	var product models.Product
	if err := json.NewDecoder(c.Request().Body).Decode(&product); err != nil {
		log.Printf("Error decoding JSON: %v", err)
		return errorResponse(c, http.StatusBadRequest, "Error decoding JSON")
	}

	product.DateCreated = time.Now() // Set DateCreated
	product.DateDeleted = nil        // Ensure DateDeleted is nil

	if err := db.Table("active_products").Create(&product).Error; err != nil {
		log.Printf("Error inserting product: %v", err)
		return errorResponse(c, http.StatusInternalServerError, "Error inserting product")
	}

	log.Printf("Successfully added product: %+v", product)
	return c.JSON(http.StatusCreated, product)
}

// UpdateProduct updates an existing product
func UpdateProduct(c echo.Context) error {
	productID := c.Param("product_id")
	log.Printf("Updating product with ID: %s", productID)

	db := getDB()
	if db == nil {
		return errorResponse(c, http.StatusInternalServerError, "Failed to connect to the database")
	}

	var updatedProduct models.Product
	if err := c.Bind(&updatedProduct); err != nil {
		log.Printf("Failed to parse request body: %v", err)
		return errorResponse(c, http.StatusBadRequest, "Failed to parse request body")
	}

	// Ensure DateCreated is set to the current date if needed
	updatedProduct.DateCreated = time.Now()

	if err := db.Table("active_products").Where("product_id = ?", productID).Updates(updatedProduct).Error; err != nil {
		log.Printf("Failed to update product: %v", err)
		return errorResponse(c, http.StatusInternalServerError, "Failed to update product")
	}

	log.Printf("Successfully updated product with ID: %s", productID)
	return c.JSON(http.StatusOK, map[string]string{"message": "Product updated successfully"})
}

// MakeProductsInactive moves a product to the inactive_products table
func MakeProductsInactive(c echo.Context) error {
	productID, err := strconv.Atoi(c.Param("product_id"))
	if err != nil {
		log.Printf("Error converting product ID: %v", err)
		return errorResponse(c, http.StatusBadRequest, "Invalid product ID")
	}

	log.Printf("Making product with ID: %d inactive", productID)
	db := getDB()
	if db == nil {
		return errorResponse(c, http.StatusInternalServerError, "Failed to connect to the database")
	}

	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var prod models.Product
	if err := tx.Table("active_products").Where("product_id = ?", productID).First(&prod).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("Product not found: %d", productID)
			return errorResponse(c, http.StatusNotFound, "Product not found")
		}
		log.Printf("Error fetching product: %v", err)
		return errorResponse(c, http.StatusInternalServerError, "Failed to fetch product")
	}

	// Set the current date for the 'date_deleted' field
	now := time.Now()
	prod.DateDeleted = &now // Use the address of now

	// Prepare the product for the inactive_products table
	pendingProduct := struct {
		ProductID          int        `json:"product_id"`
		CategoryName       string     `json:"category_name"`
		ProductName        string     `json:"product_name"`
		ProductCode        string     `json:"product_code"`
		ProductDescription string     `json:"product_description"`
		DateCreated        time.Time  `json:"date_created"`
		Quantity           int        `json:"quantity"`
		ReorderLevel       int        `json:"reorder_level"`
		BuyingPrice        float64    `json:"buying_price"`
		SellingPrice       float64    `json:"selling_price"`
		DateDeleted        *time.Time `json:"date_deleted"`
	}{
		ProductID:          prod.ProductID,
		CategoryName:       prod.CategoryName,
		ProductName:        prod.ProductName,
		ProductCode:        prod.ProductCode,
		ProductDescription: prod.ProductDescription,
		DateCreated:        prod.DateCreated,
		Quantity:           prod.Quantity,
		ReorderLevel:       prod.ReorderLevel,
		BuyingPrice:        prod.BuyingPrice,
		SellingPrice:       prod.SellingPrice,
		DateDeleted:        prod.DateDeleted,
	}

	// Move the product to the inactive_products table
	if err := tx.Table("inactive_products").Create(&pendingProduct).Error; err != nil {
		log.Printf("Failed to move product to inactive products: %v", err)
		return errorResponse(c, http.StatusInternalServerError, "Failed to move product to inactive products: "+err.Error())
	}

	// Delete the product from the active_products table
	if err := tx.Table("active_products").Where("product_id = ?", productID).Delete(&models.Product{}).Error; err != nil {
		log.Printf("Failed to delete product: %v", err)
		return errorResponse(c, http.StatusInternalServerError, "Failed to delete product: "+err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("Failed to complete operation: %v", err)
		return errorResponse(c, http.StatusInternalServerError, "Failed to complete operation: "+err.Error())
	}

	log.Printf("Successfully moved product to inactive products: %d", productID)
	return c.JSON(http.StatusOK, map[string]string{"message": "Product moved to inactive products successfully"})
}
