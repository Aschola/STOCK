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

// GetProducts fetches all products
func GetProducts(c echo.Context) error {
	db := getDB()
	if db == nil {
		log.Println("Failed to get database instance while fetching products")
		return errorResponse(c, http.StatusInternalServerError, "Failed to connect to the database")
	}

	var products []models.Product
	if err := db.Table("products").Where("deleted_at IS NULL").Find(&products).Error; err != nil {
		log.Printf("Failed to fetch products: %v", err)
		return errorResponse(c, http.StatusInternalServerError, "Failed to fetch products")
	}

	return c.JSON(http.StatusOK, products)
}

// GetProductByID fetches a product by its ID
func GetProductByID(c echo.Context) error {
	productID, err := strconv.Atoi(c.Param("product_id"))
	if err != nil {
		log.Printf("Invalid product ID: %s", c.Param("product_id"))
		return errorResponse(c, http.StatusBadRequest, "Invalid product ID")
	}

	db := getDB()
	if db == nil {
		log.Println("Failed to get database instance while fetching product by ID")
		return errorResponse(c, http.StatusInternalServerError, "Failed to connect to the database")
	}

	var prod models.Product
	if err := db.Table("products").Where("product_id = ? AND deleted_at IS NULL", productID).First(&prod).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("Product not found with ID: %d", productID)
			return errorResponse(c, http.StatusNotFound, "Product not found")
		}
		log.Printf("Failed to fetch product by ID: %v", err)
		return errorResponse(c, http.StatusInternalServerError, "Failed to fetch product")
	}

	return c.JSON(http.StatusOK, prod)
}

// AddProduct adds a new product to the database
func AddProduct(c echo.Context) error {
	db := getDB()
	if db == nil {
		log.Println("Failed to get database instance while adding product")
		return errorResponse(c, http.StatusInternalServerError, "Failed to connect to the database")
	}

	var product models.Product
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

	// Step 3: Handle and parse CreatedAt and UpdatedAt
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

	// Step 4: Insert the product into the database
	if err := db.Table("products").Create(&product).Error; err != nil {
		log.Printf("Error inserting product: %v", err)
		return errorResponse(c, http.StatusInternalServerError, "Error inserting product")
	}

	// Step 5: Return the created product (including the auto-generated product_id)
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
