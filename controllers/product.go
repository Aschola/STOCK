package controllers

import (
	"encoding/json"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"log"
	"net/http"
	"stock/db"
	models "stock/models"
	"strconv"
)

import (
	"time"
)

// Convert the ISO 8601 date to TIMESTAMP format
func convertToTimestampFormat(isoDate string) (string, error) {
	// Parse the ISO 8601 date
	t, err := time.Parse(time.RFC3339, isoDate)
	if err != nil {
		return "", err
	}
	// Format the time to 'YYYY-MM-DD HH:MM:SS'
	return t.Format("2006-01-02 15:04:05"), nil
}

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
func MoveProductFromPendingDeletion(c echo.Context) error {
	productID, err := strconv.Atoi(c.Param("product_id"))
	if err != nil {
		return errorResponse(c, http.StatusBadRequest, "Invalid product ID")
	}

	db := getDB()
	if db == nil {
		return errorResponse(c, http.StatusInternalServerError, "Failed to connect to the database")
	}

	tx := db.Begin()
	if tx.Error != nil {
		return errorResponse(c, http.StatusInternalServerError, "Error starting transaction")
	}
	defer tx.Rollback()

	var prod models.Product
	if err := tx.Table("pending_deletion_products").Where("product_id = ?", productID).First(&prod).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errorResponse(c, http.StatusNotFound, "Product not found in pending deletion")
		}
		return errorResponse(c, http.StatusInternalServerError, "Failed to fetch product")
	}

	// Convert date to TIMESTAMP format
	formattedDate, err := convertToTimestampFormat(prod.Date)
	if err != nil {
		return errorResponse(c, http.StatusBadRequest, "Invalid date format")
	}
	prod.Date = formattedDate
	if err := tx.Table("products").Create(&prod).Error; err != nil {
		return errorResponse(c, http.StatusInternalServerError, "Failed to move product back to products")
	}

	if err := tx.Table("pending_deletion_products").Where("product_id = ?", productID).Delete(&models.Product{}).Error; err != nil {
		return errorResponse(c, http.StatusInternalServerError, "Failed to delete product from pending deletion")
	}

	if err := tx.Commit().Error; err != nil {
		return errorResponse(c, http.StatusInternalServerError, "Failed to complete operation")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Product moved back to products successfully"})
}
func MoveProductToPendingDeletion(c echo.Context) error {
	productID, err := strconv.Atoi(c.Param("product_id"))
	if err != nil {
		return errorResponse(c, http.StatusBadRequest, "Invalid product ID")
	}

	db := getDB()
	if db == nil {
		return errorResponse(c, http.StatusInternalServerError, "Failed to connect to the database")
	}

	tx := db.Begin()
	if tx.Error != nil {
		return errorResponse(c, http.StatusInternalServerError, "Error starting transaction")
	}
	defer tx.Rollback()

	var prod models.Product
	if err := tx.Table("products").Where("product_id = ?", productID).First(&prod).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return errorResponse(c, http.StatusNotFound, "Product not found")
		}
		return errorResponse(c, http.StatusInternalServerError, "Failed to fetch product")
	}

	// Convert date to TIMESTAMP format
	formattedDate, err := convertToTimestampFormat(prod.Date)
	if err != nil {
		return errorResponse(c, http.StatusBadRequest, "Invalid date format")
	}
	prod.Date = formattedDate

	if err := tx.Table("pending_deletion_products").Create(&prod).Error; err != nil {
		return errorResponse(c, http.StatusInternalServerError, "Failed to move product to pending deletion")
	}

	if err := tx.Table("products").Where("product_id = ?", productID).Delete(&models.Product{}).Error; err != nil {
		return errorResponse(c, http.StatusInternalServerError, "Failed to delete product")
	}

	if err := tx.Commit().Error; err != nil {
		return errorResponse(c, http.StatusInternalServerError, "Failed to complete operation")
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "Product moved to pending deletion successfully"})
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

	// Assuming `product.Date` is in ISO 8601 format and needs conversion
	formattedDate, err := convertToTimestampFormat(product.Date)
	if err != nil {
		return errorResponse(c, http.StatusBadRequest, "Invalid date format")
	}
	product.Date = formattedDate

	if err := db.Table("products").Create(&product).Error; err != nil {
		return errorResponse(c, http.StatusInternalServerError, "Error inserting product")
	}

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

	// Convert the ISO 8601 date format to MySQL TIMESTAMP format
	formattedDate, err := convertToTimestampFormat(updatedProduct.Date)
	if err != nil {
		return errorResponse(c, http.StatusBadRequest, "Invalid date format")
	}
	updatedProduct.Date = formattedDate

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
