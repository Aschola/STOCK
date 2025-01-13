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
	log.Printf("Error: %s", message)
	return echo.NewHTTPError(statusCode, message)
}

// GetProducts fetches all products along with their stock details
func GetProducts(c echo.Context) error {
	db := getDB()
	if db == nil {
		log.Println("Failed to get database instance while fetching products")
		return errorResponse(c, http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Declare a struct to hold the product and stock details together
	type ProductWithStock struct {
		models.Product
		Quantity     int     `json:"quantity"`
		BuyingPrice  float64 `json:"buying_price"`
		SellingPrice float64 `json:"selling_price"`
	}

	var products []ProductWithStock
	// Log the query details for debugging
	log.Println("Fetching products along with stock details...")

	// Perform the join between the products and stock tables based on product_id
	if err := db.Table("products").
		Select("products.*, stock.quantity, stock.buying_price, stock.selling_price").
		Joins("LEFT JOIN stock ON products.product_id = stock.product_id").
		Where("products.deleted_at IS NULL"). // Ensure no soft-deleted products are fetched
		Find(&products).Error; err != nil {
		log.Printf("Error fetching products: %v", err)
		return errorResponse(c, http.StatusInternalServerError, "Failed to fetch products")
	}

	// Log the number of products retrieved
	log.Printf("Retrieved %d products from the database", len(products))

	// Log the products with their stock details for debugging
	for _, product := range products {
		log.Printf("Product ID: %d, Name: %s, Category: %s, Quantity: %d, Buying Price: %.2f, Selling Price: %.2f",
			product.ProductID, product.ProductName, product.CategoryName, product.Quantity, product.BuyingPrice, product.SellingPrice)
	}

	// Return the products with stock details in the response
	return c.JSON(http.StatusOK, products)
}

// GetProductByID fetches a product by its ID and includes stock details
func GetProductByID(c echo.Context) error {
	// Get product_id from the URL parameter
	productID, err := strconv.Atoi(c.Param("product_id"))
	if err != nil {
		log.Printf("Invalid product ID: %s", c.Param("product_id"))
		return errorResponse(c, http.StatusBadRequest, "Invalid product ID")
	}

	// Get database connection
	db := getDB()
	if db == nil {
		log.Println("Failed to get database instance while fetching product by ID")
		return errorResponse(c, http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Declare variables for product and stock details
	var prod models.Product
	var stock models.Stock

	// Fetch the product details from the 'products' table
	if err := db.Table("products").Where("product_id = ? AND deleted_at IS NULL", productID).First(&prod).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("Product not found with ID: %d", productID)
			return errorResponse(c, http.StatusNotFound, "Product not found")
		}
		log.Printf("Failed to fetch product by ID: %v", err)
		return errorResponse(c, http.StatusInternalServerError, "Failed to fetch product")
	}

	// Fetch the stock details from the 'stock' table
	if err := db.Table("stock").Where("product_id = ?", productID).First(&stock).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("Stock details not found for Product ID: %d", productID)
			return errorResponse(c, http.StatusNotFound, "Stock details not found")
		}
		log.Printf("Failed to fetch stock details for product ID: %v", err)
		return errorResponse(c, http.StatusInternalServerError, "Failed to fetch stock details")
	}

	// Combine the product and stock details into the desired response struct
	productWithStockDetails := struct {
		ProductID          int     `json:"product_id"`
		CategoryName       string  `json:"category_name"`
		ProductName        string  `json:"product_name"`
		ProductDescription string  `json:"product_description"`
		ReorderLevel       int     `json:"reorder_level"`
		CreatedAt          string  `json:"created_at"`
		UpdatedAt          string  `json:"updated_at"`
		Quantity           int     `json:"quantity"`
		BuyingPrice        float64 `json:"buying_price"`
		SellingPrice       float64 `json:"selling_price"`
	}{
		ProductID:          prod.ProductID,
		CategoryName:       prod.CategoryName,
		ProductName:        prod.ProductName,
		ProductDescription: prod.ProductDescription,
		ReorderLevel:       prod.ReorderLevel,
		CreatedAt:          prod.CreatedAt.Format(time.RFC3339), 
		UpdatedAt:          prod.UpdatedAt.Format(time.RFC3339), 
		Quantity:           stock.Quantity,
		BuyingPrice:        stock.BuyingPrice,
		SellingPrice:       stock.SellingPrice,
	}

	// Return the combined product and stock details as JSON
	log.Printf("Fetched product and stock details successfully for Product ID: %d", productID)
	return c.JSON(http.StatusOK, productWithStockDetails)
}

func AddProduct(c echo.Context) error {
	db := getDB()
	if db == nil {
		log.Println("Failed to get database instance while adding product")
		return errorResponse(c, http.StatusInternalServerError, "Failed to connect to the database")
	}

	var product models.Product
	// Decode the request body into the product struct
	if err := json.NewDecoder(c.Request().Body).Decode(&product); err != nil {
		log.Printf("Error decoding JSON for new product: %v", err)
		return errorResponse(c, http.StatusBadRequest, "Error decoding JSON")
	}

	// Step 1: Check if the category exists in the 'categories_onlies' table
	var count int64
	if err := db.Table("categories_onlies").Where("category_name = ?", product.CategoryName).Count(&count).Error; err != nil {
		log.Printf("Error checking category existence: %v", err)
		return errorResponse(c, http.StatusInternalServerError, "Error checking category existence")
	}

	// Step 2: If no category is found, return an error response
	if count == 0 {
		log.Printf("Category does not exist: %s", product.CategoryName)
		return errorResponse(c, http.StatusBadRequest, "Category does not exist")
	}

	// Step 3: Ensure organizations_id is provided and valid
	if product.OrganizationsID == 0 {
		log.Println("organizations_id is required")
		return errorResponse(c, http.StatusBadRequest, "organizations_id is required")
	}

	// Step 4: Handle and parse CreatedAt and UpdatedAt
	if product.CreatedAt.IsZero() {
		// If CreatedAt is provided in the request, we can handle the custom format
		if createdAtStr := c.FormValue("created_at"); createdAtStr != "" {
			// Try parsing it in the custom format
			parsedTime, err := time.Parse("2006-01-02 15:04:05", createdAtStr)
			if err != nil {
				log.Printf("Error parsing created_at: %v", err)
				return errorResponse(c, http.StatusBadRequest, "Invalid created_at format")
			}
			product.CreatedAt = parsedTime
		} else {
			// Default to current time if not provided
			product.CreatedAt = time.Now()
		}
	}

	if product.UpdatedAt.IsZero() {
		product.UpdatedAt = time.Now()
	}

	// Step 5: Insert the product into the database
	if err := db.Table("products").Create(&product).Error; err != nil {
		log.Printf("Error inserting product: %v", err)
		return errorResponse(c, http.StatusInternalServerError, "Error inserting product")
	}

	// Step 6: Return the created product (including the auto-generated product_id)
	log.Printf("Product added successfully: %v", product)
	return c.JSON(http.StatusCreated, product)
}

// UpdateProduct updates an existing product
func UpdateProduct(c echo.Context) error {
	db := getDB()
	if db == nil {
		log.Println("Failed to get database instance while updating product")
		return errorResponse(c, http.StatusInternalServerError, "Failed to connect to the database")
	}

	productID := c.Param("product_id")
	var updatedProduct models.Product
	if err := c.Bind(&updatedProduct); err != nil {
		log.Printf("Failed to parse request body for product update: %v", err)
		return errorResponse(c, http.StatusBadRequest, "Failed to parse request body")
	}

	// Set the UpdatedAt timestamp when updating the product
	updatedProduct.UpdatedAt = time.Now()

	// Update the product in the database
	if err := db.Table("products").Where("product_id = ? AND deleted_at IS NULL", productID).Updates(updatedProduct).Error; err != nil {
		log.Printf("Failed to update product with ID %s: %v", productID, err)
		return errorResponse(c, http.StatusInternalServerError, "Failed to update product")
	}

	log.Printf("Product updated successfully: %s", productID)
	return c.JSON(http.StatusOK, map[string]string{"message": "Product updated successfully"})
}

// DeleteProduct soft deletes a product by setting DeletedAt timestamp
func DeleteProduct(c echo.Context) error {
	productID, err := strconv.Atoi(c.Param("product_id"))
	if err != nil {
		log.Printf("Invalid product ID for deletion: %s", c.Param("product_id"))
		return errorResponse(c, http.StatusBadRequest, "Invalid product ID")
	}

	db := getDB()
	if db == nil {
		log.Println("Failed to get database instance while deleting product")
		return errorResponse(c, http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Set the DeletedAt timestamp for soft delete
	currentTime := time.Now()
	if err := db.Table("products").Where("product_id = ?", productID).Update("deleted_at", currentTime).Error; err != nil {
		log.Printf("Failed to soft delete product with ID %d: %v", productID, err)
		return errorResponse(c, http.StatusInternalServerError, "Failed to delete product")
	}

	log.Printf("Product soft deleted successfully: %d", productID)
	return c.JSON(http.StatusOK, map[string]string{"message": "Product soft deleted successfully"})
}

// SoftDeleting soft deletes a product by setting the deleted_at timestamp
func SoftDeleting(c echo.Context) error {
	productID, err := strconv.Atoi(c.Param("product_id"))
	if err != nil {
		log.Printf("Invalid product ID for deletion: %s", c.Param("product_id"))
		return errorResponse(c, http.StatusBadRequest, "Invalid product ID")
	}

	db := getDB()
	if db == nil {
		log.Println("Failed to get database instance while soft deleting product")
		return errorResponse(c, http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Set the DeletedAt timestamp for soft delete
	currentTime := time.Now()
	if err := db.Table("products").Where("product_id = ?", productID).Update("deleted_at", currentTime).Error; err != nil {
		log.Printf("Failed to soft delete product with ID %d: %v", productID, err)
		return errorResponse(c, http.StatusInternalServerError, "Failed to soft delete product")
	}

	log.Printf("Product soft deleted successfully: %d", productID)
	return c.JSON(http.StatusOK, map[string]string{"message": "Product soft deleted successfully"})
}

// SoftDeleteProduct handles the soft deletion of a product by updating the DeletedAt field
func SoftDeleteProduct(c echo.Context) error {
	db := getDB()
	if db == nil {
		log.Println("Failed to get database instance while soft deleting product")
		return errorResponse(c, http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Get the product ID from the URL parameter
	productID := c.Param("product_id")

	// Set the DeletedAt timestamp for soft delete
	currentTime := time.Now()

	// Soft delete the product by setting the DeletedAt timestamp
	if err := db.Table("products").Where("product_id = ?", productID).Update("deleted_at", currentTime).Error; err != nil {
		log.Printf("Failed to soft delete product with ID %s: %v", productID, err)
		return errorResponse(c, http.StatusInternalServerError, "Failed to soft delete product")
	}

	log.Printf("Product soft deleted successfullyfffffffffff: %s", productID)
	return c.JSON(http.StatusOK, map[string]string{"message": "Product soft deleted successfully"})
}

func GetSoftDeletedProducts(c echo.Context) error {
	db := getDB()
	if db == nil {
		log.Println("Failed to get database instance while fetching soft-deleted products")
		return errorResponse(c, http.StatusInternalServerError, "Failed to connect to the database")
	}

	var products []models.Product
	// Fetch products with non-NULL DeletedAt (soft-deleted products)
	if err := db.Table("products").Where("deleted_at IS NOT NULL").Find(&products).Error; err != nil {
		log.Printf("Failed to fetch soft-deleted products: %v", err)
		return errorResponse(c, http.StatusInternalServerError, "Failed to fetch soft-deleted products")
	}

	return c.JSON(http.StatusOK, products)
}

func ActivateProduct(c echo.Context) error {
	db := getDB()
	if db == nil {
		log.Println("Failed to get database instance while activating product")
		return errorResponse(c, http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Get the product ID from the URL
	productID := c.Param("product_id")
	if productID == "" {
		log.Println("Product ID is required")
		return errorResponse(c, http.StatusBadRequest, "Product ID is required")
	}

	// Update the deleted_at field to NULL to activate the product
	if err := db.Table("products").Where("product_id = ? AND deleted_at IS NOT NULL", productID).Update("deleted_at", nil).Error; err != nil {
		log.Printf("Failed to activate product with ID %s: %v", productID, err)
		return errorResponse(c, http.StatusInternalServerError, "Failed to activate product")
	}

	// Return a success message
	log.Printf("Product with ID %s activated successfully", productID)
	return c.JSON(http.StatusOK, map[string]string{"message": "Product activated successfully"})
}
