package controllers

import (
	"fmt"
	"log"
	"net/http"

	"stock/db"
	"strconv"

	"github.com/labstack/echo/v4"

	"gorm.io/gorm"
	"stock/models"
)

// AdminCreateStock handles the creation of a stock item
func CreateStock(c echo.Context) error {
	log.Println("CreateStock - Entry")

	// Retrieve the user's role and organization ID from the context
	roleName, ok := c.Get("roleName").(string)
	if !ok {
		log.Println("CreateStock - Unauthorized: roleName not found in context")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	organizationIDRaw := c.Get("organizationID")
	if organizationIDRaw == nil {
		log.Println("CreateStock - Unauthorized: organizationID not found in context")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	organizationID, ok := organizationIDRaw.(uint)
	if !ok {
		log.Println("CreateStock - Unauthorized: organizationID is not of type uint")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	// Log the role and organization ID (optional)
	log.Printf("CreateStock - User Role: %s, OrganizationID: %d", roleName, organizationID)

	// Create the stock item by binding the request body to the Stock model
	var stock models.Stock
	if err := c.Bind(&stock); err != nil {
		log.Printf("CreateStock - Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	// Check if the stock already exists in the organization
	var existingStock models.Stock
	if err := db.GetDB().
		Where("product_id = ? AND organization_id = ? AND deleted_at IS NULL", stock.ProductID, organizationID).
		First(&existingStock).Error; err == nil {
		log.Printf("CreateStock - Stock already exists: ProductID %d in OrganizationID %d", stock.ProductID, organizationID)
		return c.JSON(http.StatusConflict, echo.Map{"error": "Stock already exists for this product in the organization"})
	} else if err != gorm.ErrRecordNotFound {
		log.Printf("CreateStock - Error checking existing stock: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not check for existing stock"})
	}

	// Set the organization ID of the stock to the user's organization
	stock.OrganizationID = organizationID

	log.Printf("CreateStock - New stock data: %+v", stock)

	// Insert the new stock into the database
	if err := db.GetDB().Create(&stock).Error; err != nil {
		log.Printf("CreateStock - Create error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	log.Println("CreateStock - Stock item created successfully")
	log.Println("CreateStock - Exit")
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

	// Create a new stock struct to hold only the editable fields
	var updatedStock struct {
		Quantity    int       `json:"quantity"`
		BuyingPrice float64   `json:"buying_price"`
		SellingPrice float64  `json:"selling_price"`
		ExpiryDate  *string   `json:"expiry_date"`
	}

	// Bind the input data to the new struct
	if err := c.Bind(&updatedStock); err != nil {
		log.Printf("AdminEditStock - Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	// Update only the editable fields in the stock item
	stock.Quantity = updatedStock.Quantity
	stock.BuyingPrice = updatedStock.BuyingPrice
	stock.SellingPrice = updatedStock.SellingPrice
	//stock.ExpiryDate = updatedStock.ExpiryDate

	// Save the updated stock item
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

	// SQL Query with INNER JOIN to include supplier's name
	query := `
        SELECT 
            s.id,
            s.product_id,
            p.product_name AS product_name,
            s.quantity,
            s.buying_price,
            s.selling_price,
            s.expiry_date,
            p.product_description AS product_description,
            su.name AS supplier_name  -- Added supplier name
        FROM stock s
        INNER JOIN products p ON s.product_id = p.product_id
        INNER JOIN suppliers su ON su.id = s.supplier_id  -- Joining with suppliers table
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
			supplierName       string  // Added variable for supplier name
		)

		err = rows.Scan(&id, &productID, &productName, &quantity, &buyingPrice, &sellingPrice, &expiryDate, &productDescription, &supplierName)
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
			"supplier_name":       supplierName,  // Added supplier name to the response
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


// func ViewStockByID(c echo.Context) error {
//     log.Println("GetStock - Entry")

//     // Get supplier ID from path
//     id, err := strconv.Atoi(c.Param("id"))
//     if err != nil {
//         log.Printf("GetStock - Invalid ID: %v", err)
//         return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid stock ID"})
//     }

//     var stock models.Stock
//     if err := db.GetDB().First(&stock, id).Error; err != nil {
//         log.Printf("GetStock - Stock not found: %v", err)
//         return c.JSON(http.StatusNotFound, echo.Map{"error": "Stock not found"})
//     }

//     log.Println("GetStock - Stock retrieved successfully")
//     log.Println("GetStock - Exit")
//     return c.JSON(http.StatusOK, stock)
// }
func ViewStockByID(c echo.Context) error {
    log.Println("ViewStockByID - Entry")

    // Get stock ID from path
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        log.Printf("ViewStockByID - Invalid ID: %v", err)
        return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid stock ID"})
    }

    gormDB := db.GetDB()

    // SQL Query to fetch stock details with associated product and supplier details
    query := `
        SELECT 
            s.id,
            s.product_id,
            p.product_name AS product_name,
            s.quantity,
            s.buying_price,
            s.selling_price,
            s.expiry_date,
            p.product_description AS product_description,
            su.name AS supplier_name  -- Added supplier name
        FROM stock s
        INNER JOIN products p ON s.product_id = p.product_id
        INNER JOIN suppliers su ON su.id = s.supplier_id  -- Joining with suppliers table
        WHERE s.id = ?
    `

    var (
        idVal                uint64
        productID            uint64
        productName          string
        quantity             int
        buyingPrice          float64
        sellingPrice         float64
        expiryDate           *string
        productDescription   string
        supplierName         string  // Added variable for supplier name
    )

    // Execute the query
    row := gormDB.Raw(query, id).Row()
    if err := row.Scan(&idVal, &productID, &productName, &quantity, &buyingPrice, &sellingPrice, &expiryDate, &productDescription, &supplierName); err != nil {
        log.Printf("ViewStockByID - Stock not found: %v", err)
        return c.JSON(http.StatusNotFound, echo.Map{"error": "Stock not found"})
    }

    // Construct response
    stock := map[string]interface{}{
        "id":                  idVal,
        "product_id":          productID,
        "product_name":        productName,
        "quantity":            quantity,
        "buying_price":        buyingPrice,
        "selling_price":       sellingPrice,
        "expiry_date":         expiryDate,
        "product_description": productDescription,
        "supplier_name":       supplierName,  // Added supplier name to the response
    }

    log.Println("ViewStockByID - Stock retrieved successfully")
    log.Println("ViewStockByID - Exit")
    return c.JSON(http.StatusOK, stock)
}
