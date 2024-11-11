package controllers

import (
	"log"
	"net/http"
	//"strings"

	"github.com/labstack/echo/v4"
	//"gorm.io/gorm"
	"stock/db"
	"stock/models"
)

// AdminCreateStock creates a new stock item
func CreateStock(c echo.Context) error {
	log.Println("AdminCreateStock - Entry")

	roleName, ok := c.Get("roleName").(string)
	if !ok || roleName != "Admin" {
		log.Println("AdminCreateStock - Permission denied: non-admin trying to create stock")
		return c.JSON(http.StatusForbidden, echo.Map{"error": "Permission denied"})
	}

	var stock models.Stock
	if err := c.Bind(&stock); err != nil {
		log.Printf("AdminCreateStock - Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	log.Printf("AdminCreateStock - New stock data: %+v", stock)

	if err := db.GetDB().Create(&stock).Error; err != nil {
		log.Printf("AdminCreateStock - Create error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	log.Println("AdminCreateStock - Stock item created successfully")
	log.Println("AdminCreateStock - Exit")
	return c.JSON(http.StatusOK, echo.Map{"message": "Stock item created successfully"})
}

// AdminEditStock edits an existing stock item
func EditStock(c echo.Context) error {
	id := c.Param("id")
	log.Printf("AdminEditStock - Entry with ID: %s", id)

	roleName, ok := c.Get("roleName").(string)
	if !ok || roleName != "Admin" {
		log.Println("AdminEditStock - Permission denied: non-admin trying to edit stock")
		return c.JSON(http.StatusForbidden, echo.Map{"error": "Permission denied"})
	}

	var stock models.Stock
	if err := db.GetDB().First(&stock, id).Error; err != nil {
		log.Printf("AdminEditStock - Stock not found: %v", err)
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Stock item not found"})
	}

	log.Printf("AdminEditStock - Current stock details: %+v", stock)

	if err := c.Bind(&stock); err != nil {
		log.Printf("AdminEditStock - Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	if err := db.GetDB().Save(&stock).Error; err != nil {
		log.Printf("AdminEditStock - Save error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	log.Println("AdminEditStock - Stock item updated successfully")
	log.Println("AdminEditStock - Exit")
	return c.JSON(http.StatusOK, stock)
}

// AdminDeleteStock permanently deletes a stock item
func DeleteStock(c echo.Context) error {
	id := c.Param("id")
	log.Printf("AdminDeleteStock - Entry with ID: %s", id)

	roleName, ok := c.Get("roleName").(string)
	if !ok || roleName != "Admin" {
		log.Println("AdminDeleteStock - Permission denied: non-admin trying to delete stock")
		return c.JSON(http.StatusForbidden, echo.Map{"error": "Permission denied"})
	}

	if err := db.GetDB().Unscoped().Delete(&models.Stock{}, id).Error; err != nil {
		log.Printf("AdminDeleteStock - Delete error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to delete stock item"})
	}

	log.Println("AdminDeleteStock - Stock item permanently deleted successfully")
	log.Println("AdminDeleteStock - Exit")
	return c.JSON(http.StatusOK, echo.Map{"message": "Stock item permanently deleted successfully"})
}
func ViewAllStock(c echo.Context) error {
    var stocks []models.Stock

    if err := db.GetDB().Find(&stocks).Error; err != nil {
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not retrieve stock items"})
    }

    return c.JSON(http.StatusOK, echo.Map{"stocks": stocks})
}

func ViewStockByID(c echo.Context) error {
    id := c.Param("id")
    var stock models.Stock

    if err := db.GetDB().First(&stock, id).Error; err != nil {
        return c.JSON(http.StatusNotFound, echo.Map{"error": "Stock item not found"})
    }

    return c.JSON(http.StatusOK, echo.Map{"stock": stock})
}
