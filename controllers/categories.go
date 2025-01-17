package controllers

import (
	"log"
	"net/http"
	"stock/db"
	models "stock/models"

	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// Update an existing category in the categories_onlies table and update related products
func UpdateCategory(c echo.Context) error {
	// Retrieve the category ID from the URL parameter
	categoryID := c.Param("id")
	log.Printf("Received request to update category with ID: %s", categoryID)

	// Retrieve the database connection
	db := db.GetDB()

	// Struct to hold the category data
	var category models.Categories_Only

	// Bind the incoming JSON request data to the category struct
	if err := c.Bind(&category); err != nil {
		log.Printf("Error binding payload: %s", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, "Error binding payload")
	}

	// Check if the category name is provided
	if category.CategoryName == "" {
		log.Printf("Category name is required")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Category name is required"})
	}

	// Query for the category with the given category_id
	var existingCategory models.Categories_Only
	if err := db.First(&existingCategory, categoryID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("Category with ID %s not found in categories_onlies", categoryID)
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "Category not found",
			})
		}
		log.Printf("Error querying category: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Internal Server Error",
		})
	}

	// Store the old category name before update
	oldCategoryName := existingCategory.CategoryName
	log.Printf("Old category name: %s", oldCategoryName)

	// Update the category in categories_onlies
	if err := db.Model(&existingCategory).Where("category_id = ?", categoryID).Updates(models.Categories_Only{
		CategoryName: category.CategoryName,
	}).Error; err != nil {
		log.Printf("Error updating category in categories_onlies: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Error updating category in categories_onlies",
		})
	}

	// Debugging: Check if we are updating the products with the correct query
	log.Printf("Attempting to update products with old category name: %s", oldCategoryName)

	// Now, update the category_name in products where it matches the old category_name
	result := db.Model(&models.Product{}).Where("category_name = ?", oldCategoryName).Updates(map[string]interface{}{
		"category_name": category.CategoryName,
	})

	if result.Error != nil {
		log.Printf("Error updating products with new category_name: %s", result.Error.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Error updating related products in products table",
		})
	}

	// Check how many rows were updated
	if result.RowsAffected > 0 {
		log.Printf("Successfully updated %d products", result.RowsAffected)
	} else {
		log.Printf("No products updated. This might be due to a mismatch in category_name.")
	}

	// Log success and return the updated category
	log.Printf("Category and related products updated successfully with ID: %s", categoryID)
	return c.JSON(http.StatusOK, category)
}

// Fetch Categories from categories_onlies
func GetCategories(c echo.Context) error {
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
func CreateCategory(c echo.Context) error {
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

// DeleteCategoryByID deletes a category from the categories_onlies table by category_id
func DeleteCategoryByID(c echo.Context) error {
	// Retrieve category_id from URL parameters
	categoryID := c.Param("id")
	log.Printf("Received request to delete category with ID: %s", categoryID)

	// Get the database connection
	db := db.GetDB()

	// Delete the category from categories_onlies table by category_id
	result := db.Delete(&models.Categories_Only{}, "category_id = ?", categoryID)

	// Handle possible errors
	if result.Error != nil {
		log.Printf("Error deleting category from categories_onlies: %s", result.Error.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
	}

	// If no rows were affected, return a not found error
	if result.RowsAffected == 0 {
		log.Printf("Category with ID %s not found in categories_onlies", categoryID)
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Category not found"})
	}

	// Log success and return a success message
	log.Printf("Successfully deleted category with ID %s from categories_onlies", categoryID)
	return c.JSON(http.StatusOK, map[string]string{"message": "Category deleted successfully"})
}

// GetProductsByCategoryID retrieves all products associated with the category_id
func GetProductsByCategoryID(c echo.Context) error {
	// Retrieve category_id from URL parameters
	categoryID := c.Param("id")
	log.Printf("Received request to fetch products for category with ID: %s", categoryID)

	// Get the database connection
	db := db.GetDB()

	// Step 1: Get category_name using category_id
	var category models.Categories_Only
	if err := db.Where("category_id = ?", categoryID).First(&category).Error; err != nil {
		// If there's an error or no category found
		log.Printf("Error retrieving category name for category_id %s: %s", categoryID, err.Error())
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Category not found"})
	}

	// Step 2: Fetch products by category_name
	var products []models.Product
	result := db.Where("category_name = ?", category.CategoryName).Find(&products)

	// Handle possible errors
	if result.Error != nil {
		log.Printf("Error retrieving products for category %s: %s", category.CategoryName, result.Error.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
	}

	// If no products found for the category, return a not found error
	if result.RowsAffected == 0 {
		log.Printf("No products found for category %s", category.CategoryName)
		return c.JSON(http.StatusNotFound, map[string]string{"error": "No products found for this category"})
	}

	// Return the list of products as a JSON response
	log.Printf("Successfully retrieved products for category %s", category.CategoryName)
	return c.JSON(http.StatusOK, products)
}
