package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"stock/db"
	models "stock/models"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func GetCategories(c echo.Context) error {
	log.Println("Received request to fetch categories")

	db := db.GetDB()

	var categories []models.Category
	if err := db.Find(&categories).Error; err != nil {
		log.Printf("Error querying categories from database: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
	}

	log.Printf("Fetched %d categories", len(categories))

	return c.JSON(http.StatusOK, categories)
}

func GetCategoryByID(c echo.Context) error {
	categoryID := c.Param("category_id")
	if categoryID == "" {
		log.Printf("No category ID provided in the request")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Category ID is required"})
	}
	log.Printf("Received request to fetch category with ID: %s", categoryID)

	// Get the database connection
	db := db.GetDB()

	// Query the category from the Categories table
	var category models.Category
	if err := db.First(&category, categoryID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("Category with ID %s not found", categoryID)
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Category not found"})
		}
		log.Printf("Error querying category: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
	}

	// Log the fetched Category
	log.Printf("Fetched category with ID %s: %+v", categoryID, category)

	// Return the fetched Category as JSON
	return c.JSON(http.StatusOK, category)
}

func CreateCategories(c echo.Context) error {
	// Get the database connection
	db := db.GetDB()

	// Parse JSON manually from request body
	var category models.Category
	if err := json.NewDecoder(c.Request().Body).Decode(&category); err != nil {
		log.Printf("Error decoding JSON: %s", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, "Error decoding JSON")
	}
	log.Printf("Received request to create a category: %+v", category)

	// Execute the SQL INSERT query to add the category to the database
	if err := db.Create(&category).Error; err != nil {
		log.Printf("Error inserting a category: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Error inserting a category",
		})
	}

	// Log the successful creation and the category details
	log.Printf("Category created successfully. category_id: %d, category_name: %s, product_name: %s, product_description: %s",
		category.CategoryID, category.CategoryName, category.ProductName, category.ProductDescription)

	// Return the created category as JSON with status 201 Created
	return c.JSON(http.StatusCreated, category)
}

func UpdateCategory(c echo.Context) error {

	db := db.GetDB()

	categoryID := c.Param("category_id")

	// Bind the request payload to a new Category struct
	var category models.Category
	if err := c.Bind(&category); err != nil {
		log.Printf("Error binding payload: %s", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, "Error binding payload")
	}

	if err := db.Model(&models.Category{}).Where("category_id = ?", categoryID).Updates(category).Error; err != nil {
		log.Printf("Error updating a category: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Error updating category")
	}

	// Return success message
	log.Printf("Category updated successfully")
	return c.JSON(http.StatusOK, "Category updated successfully")
}

func DeleteCategoryByID(c echo.Context) error {
	// Extract category_id from request parameters
	categoryID := c.Param("id")
	log.Printf("Received request to delete category with ID: %s", categoryID)

	// Get the database connection
	db := db.GetDB()

	// Delete the category
	result := db.Delete(&models.Category{}, categoryID)
	if result.Error != nil {
		log.Printf("Error deleting category: %s", result.Error.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
	}

	// Check if any rows were affected
	if result.RowsAffected == 0 {
		log.Printf("Category with ID %s not found", categoryID)
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Category not found"})
	}

	// Log successful deletion
	log.Printf("Deleted category with ID %s", categoryID)

	// Return success response
	return c.JSON(http.StatusOK, map[string]string{"message": "Category deleted successfully"})
}
