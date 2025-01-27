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
		log.Println("Error: Failed to get database instance")
	}
	return db
}

// Utility function for error responses with logging
func errorResponse(c echo.Context, statusCode int, message string) error {
	log.Printf("Error: %s", message) // Log the detailed error message
	return echo.NewHTTPError(statusCode, message)
}

// GetProducts fetches all products along with their stock details filtered by organization ID
func GetProducts(c echo.Context) error {
	db := getDB()
	if db == nil {
		log.Println("Error: Failed to get database instance while fetching products")
		return errorResponse(c, http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Retrieve organizationID from context (middleware should have set this)
	organizationID, ok := c.Get("organizationID").(uint)
	if !ok {
		log.Println("Error: Failed to get organizationID from context")
		return errorResponse(c, http.StatusUnauthorized, "Unauthorized")
	}

	// Convert organizationID from uint to int64 for DB compatibility
	orgID := int64(organizationID)

	type ProductWithStock struct {
		models.Product
		Quantity     int     `json:"quantity"`
		BuyingPrice  float64 `json:"buying_price"`
		SellingPrice float64 `json:"selling_price"`
	}

	var products []ProductWithStock
	log.Printf("Fetching products for organization ID: %d...", organizationID)

	// Perform the join and filter by organizations_id
	if err := db.Table("products").
		Select("products.*, stock.quantity, stock.buying_price, stock.selling_price").
		Joins("LEFT JOIN stock ON products.product_id = stock.product_id").
		Where("products.deleted_at IS NULL AND products.organizations_id = ?", orgID).
		Find(&products).Error; err != nil {
		log.Printf("Error fetching products for organization ID %d: %v", organizationID, err)
		return errorResponse(c, http.StatusInternalServerError, "Failed to fetch products")
	}

	log.Printf("Successfully retrieved %d products for organization ID: %d", len(products), organizationID)

	return c.JSON(http.StatusOK, products)
}

// GetProductByID fetches a product by its ID and includes stock details
func GetProductByID(c echo.Context) error {
	productID, err := strconv.Atoi(c.Param("product_id"))
	if err != nil {
		log.Printf("Error: Invalid product ID in URL: %v", err)
		return errorResponse(c, http.StatusBadRequest, "Invalid product ID")
	}

	db := getDB()
	if db == nil {
		log.Println("Error: Failed to get database instance while fetching product by ID")
		return errorResponse(c, http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Retrieve organizationID from context (middleware should have set this)
	organizationID, ok := c.Get("organizationID").(uint)
	if !ok {
		log.Println("Error: Failed to get organizationID from context")
		return errorResponse(c, http.StatusUnauthorized, "Unauthorized")
	}

	// Convert organizationID from uint to int64 for DB compatibility
	orgID := int64(organizationID)

	var prod models.Product
	var stock models.Stock

	// Fetch product details
	if err := db.Table("products").Where("product_id = ? AND deleted_at IS NULL AND organizations_id = ?", productID, orgID).First(&prod).Error; err != nil {
		log.Printf("Error: Product not found for ID %d in organization %d: %v", productID, organizationID, err)
		if err == gorm.ErrRecordNotFound {
			return errorResponse(c, http.StatusNotFound, "Product not found")
		}
		return errorResponse(c, http.StatusInternalServerError, "Failed to fetch product details")
	}

	// Fetch stock details
	if err := db.Table("stock").Where("product_id = ?", productID).First(&stock).Error; err != nil {
		log.Printf("Error: Stock details not found for product ID %d in organization %d: %v", productID, organizationID, err)
		if err == gorm.ErrRecordNotFound {
			return errorResponse(c, http.StatusNotFound, "Stock details not found")
		}
		return errorResponse(c, http.StatusInternalServerError, "Failed to fetch stock details")
	}

	// Return combined product and stock details
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

	log.Printf("Successfully fetched product and stock details for product ID: %d in organization ID: %d", productID, organizationID)
	return c.JSON(http.StatusOK, productWithStockDetails)
}

// AddProduct adds a new product and performs category validation
func AddProduct(c echo.Context) error {
	db := getDB()
	if db == nil {
		log.Println("Error: Failed to get database instance while adding product")
		return errorResponse(c, http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Retrieve organizationID from context (middleware should have set this)
	organizationID, ok := c.Get("organizationID").(uint)
	if !ok {
		log.Println("Error: Failed to get organizationID from context")
		return errorResponse(c, http.StatusUnauthorized, "Unauthorized")
	}

	// Convert organizationID from uint to int64 for DB compatibility
	orgID := int64(organizationID)

	var product models.Product
	if err := json.NewDecoder(c.Request().Body).Decode(&product); err != nil {
		log.Printf("Error: Failed to decode product JSON: %v", err)
		return errorResponse(c, http.StatusBadRequest, "Invalid JSON format")
	}

	// Check if the category exists in the 'categories' table
	var count int64
	if err := db.Table("categories").Where("category_name = ?", product.CategoryName).Count(&count).Error; err != nil {
		log.Printf("Error: Failed to check category existence for '%s': %v", product.CategoryName, err)
		return errorResponse(c, http.StatusInternalServerError, "Error checking category existence")
	}

	if count == 0 {
		log.Printf("Error: Category '%s' does not exist", product.CategoryName)
		return errorResponse(c, http.StatusBadRequest, "Category does not exist")
	}

	// Set the organizationID on the product
	product.OrganizationsID = orgID

	// Handle CreatedAt and UpdatedAt fields
	if product.CreatedAt.IsZero() {
		if createdAtStr := c.FormValue("created_at"); createdAtStr != "" {
			parsedTime, err := time.Parse("2006-01-02 15:04:05", createdAtStr)
			if err != nil {
				log.Printf("Error: Failed to parse created_at: %v", err)
				return errorResponse(c, http.StatusBadRequest, "Invalid created_at format")
			}
			product.CreatedAt = parsedTime
		} else {
			product.CreatedAt = time.Now()
		}
	}

	if product.UpdatedAt.IsZero() {
		product.UpdatedAt = time.Now()
	}

	// Insert the product into the database
	if err := db.Table("products").Create(&product).Error; err != nil {
		log.Printf("Error: Failed to insert product: %v", err)
		return errorResponse(c, http.StatusInternalServerError, "Failed to add product")
	}

	log.Printf("Successfully added product with ID %d for organization ID %d", product.ProductID, organizationID)
	return c.JSON(http.StatusCreated, product)
}

// UpdateProduct updates an existing product
func UpdateProduct(c echo.Context) error {
	db := getDB()
	if db == nil {
		log.Println("Error: Failed to get database instance while updating product")
		return errorResponse(c, http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Retrieve organizationID from context (middleware should have set this)
	organizationID, ok := c.Get("organizationID").(uint)
	if !ok {
		log.Println("Error: Failed to get organizationID from context")
		return errorResponse(c, http.StatusUnauthorized, "Unauthorized")
	}

	// Convert organizationID from uint to int64 for DB compatibility
	orgID := int64(organizationID)

	productID := c.Param("product_id")
	var updatedProduct models.Product
	if err := c.Bind(&updatedProduct); err != nil {
		log.Printf("Error: Failed to bind updated product data: %v", err)
		return errorResponse(c, http.StatusBadRequest, "Failed to parse request body")
	}

	updatedProduct.UpdatedAt = time.Now()

	if err := db.Table("products").
		Where("product_id = ? AND organizations_id = ?", productID, orgID).
		Updates(updatedProduct).Error; err != nil {
		log.Printf("Error: Failed to update product with ID %s for organization ID %d: %v", productID, organizationID, err)
		return errorResponse(c, http.StatusInternalServerError, "Failed to update product")
	}

	log.Printf("Successfully updated product with ID %s for organization ID %d", productID, organizationID)
	return c.JSON(http.StatusOK, map[string]string{"message": "Product updated successfully"})
}

// DeleteProduct checks if the product exists in the stock before deleting it from products
func DeleteProduct(c echo.Context) error {
	// Extract product_id from the URL
	productIDStr := c.Param("product_id")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		log.Printf("Error: Invalid product_id %s", productIDStr)
		return errorResponse(c, http.StatusBadRequest, "Invalid product ID")
	}

	db := getDB()
	if db == nil {
		log.Println("Error: Failed to get database instance while deleting product")
		return errorResponse(c, http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Retrieve organizationID from context (middleware should have set this)
	organizationID, ok := c.Get("organizationID").(uint)
	if !ok {
		log.Println("Error: Failed to get organizationID from context")
		return errorResponse(c, http.StatusUnauthorized, "Unauthorized")
	}

	// Convert organizationID from uint to int64 for DB compatibility
	orgID := int64(organizationID)

	// Check if the product exists in stock
	var stockCount int64
	if err := db.Table("stock").Where("product_id = ? AND organization_id = ?", productID, orgID).Count(&stockCount).Error; err != nil {
		log.Printf("Error: Failed to check stock for product ID %d in organization %d: %v", productID, organizationID, err)
		return errorResponse(c, http.StatusInternalServerError, "Error checking stock")
	}

	if stockCount > 0 {
		// If the product exists in stock, return an error message
		log.Printf("Error: Product ID %d exists in stock and cannot be deleted", productID)
		return errorResponse(c, http.StatusConflict, "Product exists in stock and cannot be deleted")
	}

	// Proceed with deleting the product from the products table
	if err := db.Delete(&models.Product{}, "product_id = ? AND organizations_id = ?", productID, orgID).Error; err != nil {
		log.Printf("Error: Failed to delete product with ID %d for organization ID %d: %v", productID, organizationID, err)
		return errorResponse(c, http.StatusInternalServerError, "Failed to delete product")
	}

	log.Printf("Successfully deleted product with ID %d for organization ID %d", productID, organizationID)
	return c.JSON(http.StatusOK, map[string]string{"message": "Product deleted successfully"})
}

// UpdateProductWithoutStock updates product details (excluding stock information) based on product_id
func UpdateProductWithoutStock(c echo.Context) error {
	// Get the database instance
	db := getDB()
	if db == nil {
		log.Println("Error: Failed to get database instance while updating product")
		return errorResponse(c, http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Retrieve organizationID from context (middleware should have set this)
	organizationID, ok := c.Get("organizationID").(uint)
	if !ok {
		log.Println("Error: Failed to get organizationID from context")
		return errorResponse(c, http.StatusUnauthorized, "Unauthorized")
	}

	// Convert organizationID from uint to int64 for DB compatibility
	orgID := int64(organizationID)

	// Get product_id from the URL parameters
	productIDStr := c.Param("product_id")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		log.Printf("Error: Invalid product_id %s", productIDStr)
		return errorResponse(c, http.StatusBadRequest, "Invalid product ID")
	}

	// Fetch the product by product_id and organizations_id
	var product models.Product
	log.Printf("Fetching product with ID %d for organization ID: %d...", productID, organizationID)

	if err := db.Table("products").
		Where("product_id = ? AND organizations_id = ? AND deleted_at IS NULL", productID, orgID).
		First(&product).Error; err != nil {
		log.Printf("Error fetching product with ID %d for organization ID %d: %v", productID, organizationID, err)
		if err == gorm.ErrRecordNotFound {
			return errorResponse(c, http.StatusNotFound, "Product not found")
		}
		return errorResponse(c, http.StatusInternalServerError, "Failed to fetch product")
	}

	// Bind the request body to the product model
	var updatedProduct models.Product
	if err := c.Bind(&updatedProduct); err != nil {
		log.Printf("Error: Failed to bind updated product data: %v", err)
		return errorResponse(c, http.StatusBadRequest, "Failed to parse request body")
	}

	// Update the product's fields (but not the stock-related fields)
	updatedProduct.UpdatedAt = time.Now() // Set the current timestamp for the update

	// Only update fields that have changed (you can customize this as needed)
	if updatedProduct.ProductName != "" {
		product.ProductName = updatedProduct.ProductName
	}
	if updatedProduct.CategoryName != "" {
		product.CategoryName = updatedProduct.CategoryName
	}
	if updatedProduct.ProductDescription != "" {
		product.ProductDescription = updatedProduct.ProductDescription
	}
	if updatedProduct.ReorderLevel > 0 {
		product.ReorderLevel = updatedProduct.ReorderLevel
	}

	// Save the updated product to the database
	if err := db.Table("products").
		Where("product_id = ? AND organizations_id = ?", productID, orgID).
		Updates(product).Error; err != nil {
		log.Printf("Error: Failed to update product with ID %d for organization ID %d: %v", productID, organizationID, err)
		return errorResponse(c, http.StatusInternalServerError, "Failed to update product")
	}

	log.Printf("Successfully updated product with ID %d for organization ID %d", productID, organizationID)

	// Return the updated product as a JSON response
	return c.JSON(http.StatusOK, product)
}
