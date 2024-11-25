package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"stock/db"
	models "stock/models"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

// CustomValidator wraps the validator instance
type CustomValidator struct {
	Validator *validator.Validate
}

// Validate method for Echo's validation middleware
func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.Validator.Struct(i)
}

var validate = validator.New()

// GetCategories fetches all categories from the database.
func GetCategories(c echo.Context) error {
	log.Println("Received request to fetch categories")

	// Get the database connection
	db := db.GetDB()

	// Query all categories
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

	// Query the category
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

// CreateCategories adds a new category to the database.
func CreateCategories(c echo.Context) error {
	log.Println("Received request to create a new category")
	db := db.GetDB()
	tx := db.Begin()

	// Ensure rollback on panic
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Panic occurred: %v. Rolling back transaction.", r)
			tx.Rollback()
		}
	}()

	var category models.Category
	// Decode the JSON request body
	if err := json.NewDecoder(c.Request().Body).Decode(&category); err != nil {
		log.Printf("Error decoding JSON: %s", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, "Error decoding JSON")
	}

	// Validate the category struct
	if err := validate.Struct(category); err != nil {
		log.Printf("Validation failed: %s", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid input")
	}

	log.Printf("Category to be created: %+v", category)

	// Insert into categories table
	if err := insertIntoCategories(tx, category); err != nil {
		tx.Rollback()
		log.Printf("Error inserting into categories: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "Error inserting into categories"})
	}

	// Check if the category exists in new_categories
	exists, err := categoryExistsInNewCategories(tx, category.CategoryName)
	if err != nil {
		tx.Rollback()
		log.Printf("Error checking existence in new_categories: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "Error checking existence in new_categories"})
	}

	// Log existence without preventing insertion
	if exists {
		log.Printf("Category '%s' already exists in new_categories. Inserting into categories anyway.", category.CategoryName)
	}

	// Insert into new_categories table if it doesn't exist
	if !exists {
		if err := insertIntoNewCategories(tx, category.CategoryName); err != nil {
			tx.Rollback()
			log.Printf("Error inserting into new_categories: %s", err.Error())
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "Error inserting into new_categories"})
		}
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("Error committing transaction: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{"error": "Error committing transaction"})
	}

	log.Printf("Category created successfully. ID: %d, Name: %s", category.CategoryID, category.CategoryName)
	return c.JSON(http.StatusCreated, category)
}

// UpdateCategory modifies an existing category in both tables.
func UpdateCategoryName(c echo.Context) error {
	db := db.GetDB()
	categoryID := c.Param("category_id")

	log.Printf("Received request to update category with ID: %s", categoryID)

	var category models.Categories_only // Use Categories_only to bind the input
	if err := c.Bind(&category); err != nil {
		log.Printf("Error binding payload: %s", err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, "Error binding payload")
	}

	// Start a transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Fetch the old category name from the new_categories table
	var oldCategory models.Categories_only
	if err := tx.First(&oldCategory, categoryID).Error; err != nil {
		tx.Rollback()
		log.Printf("Error fetching old category from new_categories: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Error fetching old category")
	}

	// Update the category in the new_categories table
	if err := tx.Model(&models.Categories_only{}).Where("category_id = ?", categoryID).Updates(category).Error; err != nil {
		tx.Rollback()
		log.Printf("Error updating category in new_categories: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Error updating category in new_categories")
	}

	log.Printf("Category updated successfully in new_categories table with new name: %s", category.CategoryName)

	// Now update all matching categories in the categories table
	if err := tx.Model(&models.Category{}).Where("category_name = ?", oldCategory.CategoryName).Updates(models.Category{CategoryName: category.CategoryName}).Error; err != nil {
		tx.Rollback()
		log.Printf("Error updating all matching categories in categories table: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Error updating all matching categories")
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("Error committing transaction: %s", err.Error())
		return echo.NewHTTPError(http.StatusInternalServerError, "Error committing transaction")
	}

	log.Printf("All matching categories updated successfully in categories table to new name: %s", category.CategoryName)
	return c.JSON(http.StatusOK, "Category updated successfully")
}

// DeleteCategoryNameFromNewCategories removes a category_name from the new_categories table
// DeleteCategoryByID removes a category from the new_categories table by category_id
// and also deletes all related data in the categories table.
func DeleteCategoryByID(c echo.Context) error {
	// Get the category_id from the request
	categoryID := c.Param("category_id")
	log.Printf("Received request to delete category with ID: %s", categoryID)

	if categoryID == "" {
		log.Printf("No category_id provided in the request")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Category ID is required"})
	}

	// Get the database connection
	db := db.GetDB()

	// Start a transaction
	tx := db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("Panic occurred: %v", r)
		}
	}()

	// Fetch the category_name associated with the category_id
	var category models.Categories_only
	if err := tx.First(&category, "category_id = ?", categoryID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Printf("Category with ID '%s' not found in new_categories", categoryID)
			tx.Rollback()
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Category not found"})
		}
		log.Printf("Error fetching category: %s", err.Error())
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
	}

	categoryName := category.CategoryName
	log.Printf("Category with ID '%s' found: '%s'", categoryID, categoryName)

	// Delete related rows from the categories table
	categoriesResult := tx.Where("category_name = ?", categoryName).Delete(&models.Category{})
	if categoriesResult.Error != nil {
		log.Printf("Error deleting related categories with category_name '%s': %s", categoryName, categoriesResult.Error.Error())
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
	}
	log.Printf("Deleted %d related rows in categories table for category_name '%s'", categoriesResult.RowsAffected, categoryName)

	// Delete the category from the new_categories table
	newCategoriesResult := tx.Delete(&models.Categories_only{}, "category_id = ?", categoryID)
	if newCategoriesResult.Error != nil {
		log.Printf("Error deleting category with ID '%s' from new_categories: %s", categoryID, newCategoriesResult.Error.Error())
		tx.Rollback()
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
	}

	if newCategoriesResult.RowsAffected == 0 {
		log.Printf("Category with ID '%s' not found in new_categories table", categoryID)
		tx.Rollback()
		return c.JSON(http.StatusNotFound, map[string]string{"error": "Category not found"})
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("Error committing transaction: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error committing transaction"})
	}

	log.Printf("Deleted category with ID '%s' successfully along with related data in categories table", categoryID)
	return c.JSON(http.StatusOK, map[string]string{"message": "Category and related data deleted successfully"})
}

// categoryExistsInNewCategories checks if a category exists in the new_categories table.
func categoryExistsInNewCategories(db *gorm.DB, categoryName string) (bool, error) {
	var category models.Categories_only
	err := db.Where("category_name = ?", categoryName).First(&category).Error
	if err == nil {
		log.Printf("Category '%s' exists in new_categories.", categoryName)
		return true, nil // Category exists
	}
	if err == gorm.ErrRecordNotFound {
		log.Printf("Category '%s' does not exist in new_categories.", categoryName)
		return false, nil // Category does not exist
	}
	return false, err // Unexpected error
}

// insertIntoCategories adds a category to the categories table.
func insertIntoCategories(db *gorm.DB, category models.Category) error {
	log.Printf("Inserting category: %+v", category)
	return db.Create(&category).Error
}

// insertIntoNewCategories adds a new category to the new_categories table.
func insertIntoNewCategories(db *gorm.DB, categoryName string) error {
	newCategory := models.Categories_only{CategoryName: categoryName}
	log.Printf("Inserting into new_categories: %+v", newCategory)
	return db.Create(&newCategory).Error
}

// GetCategoriesOnly fetches all categories from the new_categories table.
func GetCategoriesOnly(c echo.Context) error {
	// Get the database connection
	db := db.GetDB()

	// Query all categories from the new_categories table
	var categories []models.CategoriesOnly
	if err := db.Table("new_categories").Find(&categories).Error; err != nil {
		log.Printf("Error querying categories from new_categories: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
	}

	log.Printf("Fetched %d categories from new_categories", len(categories))
	return c.JSON(http.StatusOK, categories)
}

// AddingCategoriesOnly adds a new category to the new_categories table.
func AddingCategoriesOnly(c echo.Context) error {
	// Get the database connection
	db := db.GetDB()
	// Define a struct to hold the request body
	type CategoryRequest struct {
		CategoryName string `json:"category_name" validate:"required"`
	}

	// Bind and validate the request body
	var categoryRequest CategoryRequest
	if err := c.Bind(&categoryRequest); err != nil {
		log.Printf("Error binding request body: %s", err.Error())
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	if err := c.Validate(categoryRequest); err != nil {
		log.Printf("Validation error: %s", err.Error())
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Validation failed"})
	}

	// Prepare the new category model
	category := models.CategoriesOnly{
		CategoryName: categoryRequest.CategoryName,
	}

	// Insert the new category into the database
	if err := db.Table("new_categories").Create(&category).Error; err != nil {
		log.Printf("Error inserting category: %s", err.Error())
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Internal Server Error"})
	}

	log.Printf("Successfully added category: %s", category.CategoryName)
	return c.JSON(http.StatusCreated, map[string]string{"message": "Category added successfully"})
}
