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

// Get all Categories from the Categories table
func GetCategories(c echo.Context) error {
	log.Println("Received request to fetch categories")

	// Get the database connection
	db := db.GetDB()

	// Query all categories from the Categories table
	var categories []models.Category
	if err := db.Find(&categories).Error; err != nil {
		log.Printf("Error querying categories from database: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
	}

	log.Printf("Fetched %d categories", len(categories))
	return c.JSON(http.StatusOK, categories)
}

// Get Category by ID from the Categories table
func GetCategoryByID(c echo.Context) error {
	categoryID := c.Param("category_id")
	if categoryID == "" {
		log.Printf("No category ID provided in the request")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Category ID is required"})
	}
	log.Printf("Received request to fetch category with ID: %s", categoryID)

	db := db.GetDB()

	var category models.Category
	if err := db.First(&category, categoryID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("Category with ID %s not found", categoryID)
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Category not found"})
		}
		log.Printf("Error querying category: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
	}

	log.Printf("Fetched category with ID %s: %+v", categoryID, category)
	return c.JSON(http.StatusOK, category)
}

// Create a new Category
func CreateCategories(c echo.Context) error {
	db := db.GetDB()

	tx := db.Begin()
	if tx.Error != nil {
		log.Printf("Error starting transaction: %s", tx.Error)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
	}

	var category models.Category
	if err := json.NewDecoder(c.Request().Body).Decode(&category); err != nil {
		log.Printf("Error decoding JSON: %s", err.Error())
		tx.Rollback()
		return echo.NewHTTPError(http.StatusBadRequest, "Error decoding JSON")
	}
	log.Printf("Received request to create a category: %+v", category)

	if err := tx.Create(&category).Error; err != nil {
		log.Printf("Error inserting category into categories table: %s", err.Error())
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Error inserting category into categories table",
		})
	}

	var existingCategory models.Categories_Only
	if err := tx.Where("category_name = ?", category.CategoryName).First(&existingCategory).Error; err == nil {
		log.Printf("Category with name '%s' already exists in categories_onlies. Skipping insertion.", category.CategoryName)
	} else if err := tx.Create(&models.Categories_Only{
		CategoryName: category.CategoryName,
	}).Error; err != nil {
		log.Printf("Error inserting category into categories_onlies table: %s", err.Error())
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": "Error inserting category into categories_onlies table",
		})
	}

	if err := tx.Commit().Error; err != nil {
		log.Printf("Error committing transaction: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to commit transaction"})
	}

	log.Printf("Category created successfully. category_id: %d, category_name: %s", category.CategoryID, category.CategoryName)
	return c.JSON(http.StatusCreated, category)
}

// Update an existing Category
func UpdateCategory(c echo.Context) error {
	db := db.GetDB()

	categoryID := c.Param("category_id")

	var category models.Category
	if err := c.Bind(&category); err != nil {
		log.Printf("Error binding payload: %s", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, "Error binding payload")
	}

	if err := db.Model(&models.Category{}).Where("category_id = ?", categoryID).Updates(category).Error; err != nil {
		log.Printf("Error updating a category: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Error updating category")
	}

	log.Printf("Category updated successfully")
	return c.JSON(http.StatusOK, "Category updated successfully")
}

// Delete a Category by ID
func DeleteCategoryByID(c echo.Context) error {
	categoryID := c.Param("id")
	log.Printf("Received request to delete category with ID: %s", categoryID)

	db := db.GetDB()

	result := db.Delete(&models.Category{}, categoryID)
	if result.Error != nil {
		log.Printf("Error deleting category: %s", result.Error.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
	}

	if result.RowsAffected == 0 {
		log.Printf("Category with ID %s not found", categoryID)
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Category not found"})
	}

	log.Printf("Deleted category with ID %s", categoryID)
	return c.JSON(http.StatusOK, map[string]string{"message": "Category deleted successfully"})
}

// Fetch Categories from categories_onlies
func GetCategoriesOnly(c echo.Context) error {
	log.Println("Received request to fetch categories from categories_onlies")

	// Get the database connection
	db := db.GetDB()

	// Query all categories from the Categories_Only table (categories_onlies)
	var categoriesOnlies []models.Categories_Only
	if err := db.Find(&categoriesOnlies).Error; err != nil {
		log.Printf("Error querying categories from categories_onlies: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
	}

	// Log the number of categories fetched
	log.Printf("Fetched %d categories from categories_onlies", len(categoriesOnlies))

	// Return the fetched categories as JSON
	return c.JSON(http.StatusOK, categoriesOnlies)
}

// CreateCategoryInCategoriesOnly adds a new category to the categories_onlies table
func CreateCategoryInCategoriesOnly(c echo.Context) error {
	// Retrieve the database connection
	db := db.GetDB()

	// Struct to hold the category data
	var category models.Categories_Only

	// Bind the incoming JSON request data to the category struct
	if err := c.Bind(&category); err != nil {
		log.Printf("Error binding payload: %s", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, "Error binding payload")
	}

	// Validate if the category name is provided
	if category.CategoryName == "" {
		log.Printf("Category name is required")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Category name is required"})
	}

	// Check if the category name already exists in the categories_onlies table
	var existingCategory models.Categories_Only
	if err := db.Where("category_name = ?", category.CategoryName).First(&existingCategory).Error; err == nil {
		// Category already exists, return a message
		log.Printf("Category with name '%s' already exists", category.CategoryName)
		return c.JSON(http.StatusConflict, map[string]string{
			"error": "Category name already exists. Please choose a different name.",
		})
	}

	// Insert the new category into the categories_onlies table
	if err := db.Create(&category).Error; err != nil {
		log.Printf("Error inserting category into categories_onlies table: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Error inserting category into categories_onlies table",
		})
	}

	// Log success and return the newly created category as JSON
	log.Printf("Category created successfully. category_id: %d, category_name: %s", category.CategoryID, category.CategoryName)
	return c.JSON(http.StatusCreated, category)
}

// GetCategoryByID retrieves a single category from categories_onlies based on the provided category_id
func GetCategoryNameByID(c echo.Context) error {
	// Get the category_id from the URL parameter
	categoryID := c.Param("id")

	// Log the received category ID
	log.Printf("Received request to fetch category with ID: %s", categoryID)

	// Retrieve the database connection
	db := db.GetDB()

	// Query for the category with the given category_id
	var category models.Categories_Only
	if err := db.First(&category, categoryID).Error; err != nil {
		// If no category found, return a 404 error
		if err.Error() == "record not found" {
			log.Printf("Category with ID %s not found", categoryID)
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Category not found",
			})
		}
		// If there is any other error, return a 500 error
		log.Printf("Error querying category: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Internal Server Error",
		})
	}

	// Log success and return the found category
	log.Printf("Fetched category with ID %s: %+v", categoryID, category)
	return c.JSON(http.StatusOK, category)
}

// EditCategoryNames updates the category name for a given category_id
func EditCategoryNames(c echo.Context) error {
	// Retrieve category_id from URL parameters
	categoryID := c.Param("id")
	log.Printf("Received request to update category with ID: %s", categoryID)

	// Retrieve the database connection
	db := db.GetDB()

	// Define a variable to hold the category data from the request body
	var category models.Categories_Only

	// Bind the request body to the category struct
	if err := c.Bind(&category); err != nil {
		log.Printf("Error binding payload: %s", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, "Error binding payload")
	}

	// Check if the category exists in the database
	var existingCategory models.Categories_Only
	if err := db.First(&existingCategory, categoryID).Error; err != nil {
		// If the category is not found, return an error
		if err.Error() == "record not found" {
			log.Printf("Category with ID %s not found", categoryID)
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Category not found",
			})
		}
		// If there is another error, return a 500 error
		log.Printf("Error querying category: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Internal Server Error",
		})
	}

	// Category exists, now update the category_name field if provided
	if category.CategoryName != "" {
		// Update the category name in the database
		if err := db.Model(&existingCategory).Where("category_id = ?", categoryID).Update("category_name", category.CategoryName).Error; err != nil {
			log.Printf("Error updating category: %s", err.Error())
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Error updating category",
			})
		}
	}

	// Log success and return the updated category
	log.Printf("Category with ID %s updated successfully", categoryID)
	return c.JSON(http.StatusOK, map[string]string{
		"message":       "Category updated successfully",
		"category_name": category.CategoryName,
	})
}
