package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"

	"log"
	models "stock/models"
)

// RestoreProductFromInactiveTable restores a product from the inactive_products table
func RestoreProductFromInactiveTable(c echo.Context) error {
	productID, err := strconv.Atoi(c.Param("product_id"))
	if err != nil {
		log.Printf("Error converting product ID: %v", err)
		return errorResponse(c, http.StatusBadRequest, "Invalid product ID")
	}

	db := getDB()
	if db == nil {
		log.Println("Failed to connect to the database")
		return errorResponse(c, http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Fetch the product from inactive_products
	var pendingProduct models.Product
	if err := db.Table("inactive_products").Where("product_id = ?", productID).First(&pendingProduct).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("Product not found in inactive products: %d", productID)
			return errorResponse(c, http.StatusNotFound, "Product not found in inactive products")
		}
		log.Printf("Error fetching product from inactive products: %v", err)
		return errorResponse(c, http.StatusInternalServerError, "Failed to fetch product from inactive products")
	}

	// Set the current date for the 'date_restored' field
	now := time.Now()

	// Create the restored product
	restoredProduct := models.Product{
		ProductID:          pendingProduct.ProductID,
		CategoryName:       pendingProduct.CategoryName,
		ProductName:        pendingProduct.ProductName,
		ProductCode:        pendingProduct.ProductCode,
		ProductDescription: pendingProduct.ProductDescription,
		DateCreated:        pendingProduct.DateCreated,
		Quantity:           pendingProduct.Quantity,
		ReorderLevel:       pendingProduct.ReorderLevel,
		BuyingPrice:        pendingProduct.BuyingPrice,
		SellingPrice:       pendingProduct.SellingPrice,
		DateDeleted:        pendingProduct.DateDeleted, // Retain DateDeleted
		DateRestored:       &now,                       // Set the DateRestored
	}

	// Insert the restored product into the active_products table
	if err := db.Table("active_products").Create(&restoredProduct).Error; err != nil {
		log.Printf("Failed to restore product: %v", err)
		return errorResponse(c, http.StatusInternalServerError, "Failed to restore product: "+err.Error())
	}

	log.Printf("Successfully restored product: %+v", restoredProduct)

	// Delete the product from inactive_products after successful insertion
	if err := db.Table("inactive_products").Where("product_id = ?", productID).Delete(&models.Product{}).Error; err != nil {
		log.Printf("Failed to remove product from inactive products: %v", err)
		return errorResponse(c, http.StatusInternalServerError, "Failed to remove product from inactive products: "+err.Error())
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Product restored successfully"})
}

// DeleteProductFromInactiveTable handles the deletion of a product from the inactive_products table
func DeleteProductFromInactiveTable(c echo.Context) error {
	// Extract product ID from the URL parameters
	productIDStr := c.Param("product_id")
	productID, err := strconv.Atoi(productIDStr)
	if err != nil {
		log.Printf("Error converting product ID: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid product ID")
	}

	// Initialize database connection
	db := getDB()
	if db == nil {
		log.Println("Failed to connect to the database")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Attempt to delete the product from the 'inactive_products' table
	var pendingProduct models.Product
	if err := db.Table("inactive_products").Where("product_id = ?", productID).Delete(&pendingProduct).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("Product not found in inactive products: %d", productID)
			return echo.NewHTTPError(http.StatusNotFound, "Product not found in inactive products")
		}
		log.Printf("Failed to delete product from inactive products: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to delete product from inactive products")
	}

	log.Printf("Successfully deleted product from inactive products: %d", productID)

	// Return a success response
	return c.JSON(http.StatusOK, map[string]string{"message": "Product deleted from inactive products successfully"})
}

// GetAllInactiveProducts retrieves all products from the inactive_products table
func GetAllInactiveProducts(c echo.Context) error {
	// Initialize database connection
	db := getDB()
	if db == nil {
		log.Println("Failed to connect to the database")
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to connect to the database")
	}

	// Query all products from the inactive_products table
	var products []models.Product
	if err := db.Table("inactive_products").Find(&products).Error; err != nil {
		log.Printf("Failed to fetch products from inactive products: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch products from inactive products")
	}

	log.Printf("Successfully fetched %d inactive products", len(products))

	// Return the fetched products as JSON
	return c.JSON(http.StatusOK, products)
}
