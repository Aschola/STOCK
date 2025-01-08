package controllers

import (
	"fmt"
	"log"
	"net/http"

	"stock/db"
	"strconv"

	"github.com/labstack/echo/v4"

	//"gorm.io/gorm"
	"stock/models"
)

// AdminCreateStock handles the creation of a stock item
func CreateStock(c echo.Context) error {
	log.Println("AdminCreateStock - Entry")

	roleName, ok := c.Get("roleName").(string)
	if !ok {
		log.Println("AdminCreateStock - Unauthorized: roleName not found in context")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	if roleName != "Admin" {
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

// AdminEditStock handles editing of a stock item
func EditStock(c echo.Context) error {
	log.Println("AdminEditStock - Entry")

	roleName, ok := c.Get("roleName").(string)
	if !ok {
		log.Println("AdminEditStock - Unauthorized: roleName not found in context")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	if roleName != "Admin" {
		log.Println("AdminEditStock - Permission denied: non-admin trying to edit stock")
		return c.JSON(http.StatusForbidden, echo.Map{"error": "Permission denied"})
	}

	id := c.Param("id")
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

// AdminDeleteStock handles permanent deletion of a stock item
func DeleteStock(c echo.Context) error {
	log.Println("AdminDeleteStock - Entry")

	roleName, ok := c.Get("roleName").(string)
	if !ok {
		log.Println("AdminDeleteStock - Unauthorized: roleName not found in context")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	if roleName != "Admin" {
		log.Println("AdminDeleteStock - Permission denied: non-admin trying to delete stock")
		return c.JSON(http.StatusForbidden, echo.Map{"error": "Permission denied"})
	}

	id := c.Param("id")
	if err := db.GetDB().Unscoped().Delete(&models.Stock{}, id).Error; err != nil {
		log.Printf("AdminDeleteStock - Delete error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to delete stock item"})
	}

	log.Println("AdminDeleteStock - Stock item permanently deleted successfully")
	log.Println("AdminDeleteStock - Exit")
	return c.JSON(http.StatusOK, echo.Map{"message": "Stock item permanently deleted successfully"})
}

// AdminViewAllStock retrieves all stock items
func ViewAllStock(c echo.Context) error {
	gormDB := db.GetDB()

	// SQL Query with INNER JOIN to ensure only valid products are matched
	query := `
        SELECT 
            s.id,
            s.product_id,
            p.product_name AS product_name,
            s.quantity,
            s.buying_price,
            s.selling_price,
            s.expiry_date,
            p.product_description AS product_description  
        FROM stock s
        INNER JOIN products p ON s.product_id = p.product_id
        WHERE p.product_id IS NOT NULL -- Ensures that we only get products that exist in the products table
    `

	// Execute the query
	rows, err := gormDB.Raw(query).Rows()
	if err != nil {
		fmt.Printf("Query execution failed: %v\nQuery: %s\n", err, query)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch stock"})
	}
	defer rows.Close()

	var stocks []map[string]interface{}
	for rows.Next() {
		var (
			id                 uint64
			productID          uint64
			productName        string
			quantity           int
			buyingPrice        float64
			sellingPrice       float64
			expiryDate         *string
			productDescription string
		)

		err = rows.Scan(&id, &productID, &productName, &quantity, &buyingPrice, &sellingPrice, &expiryDate, &productDescription)
		if err != nil {
			fmt.Printf("Error scanning row: %v\n", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error reading stock data"})
		}

		stock := map[string]interface{}{
			"id":                  id,
			"product_id":          productID,
			"product_name":        productName,
			"quantity":            quantity,
			"buying_price":        buyingPrice,
			"selling_price":       sellingPrice,
			"expiry_date":         expiryDate,
			"product_description": productDescription,
		}

		stocks = append(stocks, stock)
	}

	return c.JSON(http.StatusOK, stocks)
}

// func ViewAllStock(c echo.Context) error {
// 	log.Println("AdminViewAllStock - Entry")

// 	var stocks []models.Stock
// 	if err := db.GetDB().Find(&stocks).Error; err != nil {
// 		log.Printf("AdminViewAllStock - Retrieve error: %v", err)
// 		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not retrieve stock items"})
// 	}

// 	log.Println("AdminViewAllStock - Stock items retrieved successfully")
// 	log.Println("AdminViewAllStock - Exit")
// 	return c.JSON(http.StatusOK, echo.Map{"stocks": stocks})
// }

// AdminViewStockByID retrieves a stock item by its ID
// func ViewStockByID(c echo.Context) error {
// 	log.Println("AdminViewStockByID - Entry")

// 	id := c.Param("id")
// 	var stock models.Stock
// 	if err := db.GetDB().First(&stock, id).Error; err != nil {
// 		log.Printf("AdminViewStockByID - Stock item not found: %v", err)
// 		return c.JSON(http.StatusNotFound, echo.Map{"error": "Stock item not found"})
// 	}

// 	log.Println("AdminViewStockByID - Stock item retrieved successfully")
// 	log.Println("AdminViewStockByID - Exit")
// 	return c.JSON(http.StatusOK, echo.Map{"stock": stock})
// }
func ViewStockByID(c echo.Context) error {
    log.Println("GetStock - Entry")

    // Get supplier ID from path
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        log.Printf("GetStock - Invalid ID: %v", err)
        return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid stock ID"})
    }

    var stock models.Stock
    if err := db.GetDB().First(&stock, id).Error; err != nil {
        log.Printf("GetStock - Stock not found: %v", err)
        return c.JSON(http.StatusNotFound, echo.Map{"error": "Stock not found"})
    }

    log.Println("GetStock - Stock retrieved successfully")
    log.Println("GetStock - Exit")
    return c.JSON(http.StatusOK, stock)
}
