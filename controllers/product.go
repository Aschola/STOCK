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
	db := getDB()
	if db == nil {
		return errorResponse(c, http.StatusInternalServerError, "Failed to connect to the database")
	}

	var products []models.Product
	if err := db.Table("products").Find(&products).Error; err != nil {
		return errorResponse(c, http.StatusInternalServerError, "Failed to fetch products")
	}

	return c.JSON(http.StatusOK, products)
}

// GetProductByID fetches a product by its ID
func GetProductByID(c echo.Context) error {
	productID, err := strconv.Atoi(c.Param("product_id"))
	if err != nil {
		return errorResponse(c, http.StatusBadRequest, "Invalid product ID")
	}

	db := getDB()
	if db == nil {
		return errorResponse(c, http.StatusInternalServerError, "Failed to connect to the database")
	}

	var prod models.Product
	if err := db.Table("products").Where("product_id = ?", productID).First(&prod).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errorResponse(c, http.StatusNotFound, "Product not found")
		}
		return errorResponse(c, http.StatusInternalServerError, "Failed to fetch product")
	}

	return c.JSON(http.StatusOK, prod)
}
func AddProduct(c echo.Context) error {
	db := getDB()
	if db == nil {
		return errorResponse(c, http.StatusInternalServerError, "Failed to connect to the database")
	}

	var product models.Product
	if err := json.NewDecoder(c.Request().Body).Decode(&product); err != nil {
		return errorResponse(c, http.StatusBadRequest, "Error decoding JSON")
	}

	// Set CreatedAt and UpdatedAt if they are not automatically handled
	if product.CreatedAt.IsZero() {
		product.CreatedAt = time.Now()
	}
	if product.UpdatedAt.IsZero() {
		product.UpdatedAt = time.Now()
	}

	// Insert the product and retrieve the auto-generated product_id
	if err := db.Table("products").Create(&product).Error; err != nil {
		return errorResponse(c, http.StatusInternalServerError, "Error inserting product")
	}

	// The product struct now includes the auto-generated product_id
	return c.JSON(http.StatusCreated, product)
}

// UpdateProduct updates an existing product
func UpdateProduct(c echo.Context) error {
	db := getDB()
	if db == nil {
		return errorResponse(c, http.StatusInternalServerError, "Failed to connect to the database")
	}

	productID := c.Param("product_id")
	var updatedProduct models.Product
	if err := c.Bind(&updatedProduct); err != nil {
		return errorResponse(c, http.StatusBadRequest, "Failed to parse request body")
	}

	// Set the UpdatedAt timestamp when updating the product
	updatedProduct.UpdatedAt = time.Now()

	// Update the product in the database
	if err := db.Table("products").Where("product_id = ?", productID).Updates(updatedProduct).Error; err != nil {
		return errorResponse(c, http.StatusInternalServerError, "Failed to update product")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Product updated successfully"})
}

// DeleteProduct deletes a product by ID
func DeleteProduct(c echo.Context) error {
	productID, err := strconv.Atoi(c.Param("product_id"))
	if err != nil {
		return errorResponse(c, http.StatusBadRequest, "Invalid product ID")
	}

	db := getDB()
	if db == nil {
		return errorResponse(c, http.StatusInternalServerError, "Failed to connect to the database")
	}

	if err := db.Table("products").Where("product_id = ?", productID).Delete(&models.Product{}).Error; err != nil {
		return errorResponse(c, http.StatusInternalServerError, "Failed to delete product")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Product deleted successfully"})
}
