package controllers

import (
	"fmt"
	"log"
	"net/http"

	"stock/db"
	"strconv"
    "database/sql"


	"github.com/labstack/echo/v4"

	"gorm.io/gorm"
	"stock/models"
)

// AdminCreateStock handles the creation of a stock item
// func CreateStock(c echo.Context) error {
// 	log.Println("CreateStock - Entry")

// 	// Retrieve the user's role and organization ID from the context
// 	roleName, ok := c.Get("roleName").(string)
// 	if !ok {
// 		log.Println("CreateStock - Unauthorized: roleName not found in context")
// 		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
// 	}

// 	organizationIDRaw := c.Get("organizationID")
// 	if organizationIDRaw == nil {
// 		log.Println("CreateStock - Unauthorized: organizationID not found in context")
// 		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
// 	}

// 	organizationID, ok := organizationIDRaw.(uint)
// 	if !ok {
// 		log.Println("CreateStock - Unauthorized: organizationID is not of type uint")
// 		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
// 	}

// 	// Log the role and organization ID (optional)
// 	log.Printf("CreateStock - User Role: %s, OrganizationID: %d", roleName, organizationID)

// 	// Create the stock item by binding the request body to the Stock model
// 	var stock models.Stock
// 	if err := c.Bind(&stock); err != nil {
// 		log.Printf("CreateStock - Bind error: %v", err)
// 		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
// 	}

// 	// Check if the stock already exists in the organization
// 	var existingStock models.Stock
// 	if err := db.GetDB().
// 		Where("product_id = ? AND organization_id = ? AND deleted_at IS NULL", stock.ProductID, organizationID).
// 		First(&existingStock).Error; err == nil {
// 		log.Printf("CreateStock - Stock already exists: ProductID %d in OrganizationID %d", stock.ProductID, organizationID)
// 		return c.JSON(http.StatusConflict, echo.Map{"error": "Stock already exists for this product in the organization"})
// 	} else if err != gorm.ErrRecordNotFound {
// 		log.Printf("CreateStock - Error checking existing stock: %v", err)
// 		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not check for existing stock"})
// 	}

// 	// Set the organization ID of the stock to the user's organization
// 	stock.OrganizationID = organizationID

// 	log.Printf("CreateStock - New stock data: %+v", stock)

// 	// Insert the new stock into the database
// 	if err := db.GetDB().Create(&stock).Error; err != nil {
// 		log.Printf("CreateStock - Create error: %v", err)
// 		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
// 	}

// 	log.Println("CreateStock - Stock item created successfully")
// 	log.Println("CreateStock - Exit")
// 	return c.JSON(http.StatusOK, echo.Map{"message": "Stock item added successfully"})
// }
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

	log.Printf("CreateStock - User Role: %s, OrganizationID: %d", roleName, organizationID)

	// Bind request body to the Stock model
	var stock models.Stock
	if err := c.Bind(&stock); err != nil {
		log.Printf("CreateStock - Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	// ✅ Check if the supplier exists first
	var existingSupplier models.Suppliers
	if err := db.GetDB().
		Where("id = ?", stock.SupplierID).
		First(&existingSupplier).Error; err != nil {
		log.Printf("CreateStock - Supplier with ID %d not found", stock.SupplierID)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid supplier ID"})
	}

	// ✅ Check if the stock already exists in the organization
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

	// ✅ Set the organization ID to the user's organization
	stock.OrganizationID = organizationID

	log.Printf("CreateStock - New stock data: %+v", stock)

	// ✅ Save only the stock item (without recreating the supplier)
	if err := db.GetDB().Create(&stock).Error; err != nil {
		log.Printf("CreateStock - Create error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	log.Println("CreateStock - Stock item created successfully")
	log.Println("CreateStock - Exit")
	return c.JSON(http.StatusOK, echo.Map{"message": "Stock item added successfully"})
}

// AdminEditStock handles editing of a stock item
func EditStock(c echo.Context) error {
	id := c.Param("id")
	log.Printf("EditStock - Entry with ID: %s", id)

	var stock models.Stock
	if err := db.GetDB().First(&stock, id).Error; err != nil {
		log.Printf("EditStock - First error: %v", err)
		return c.JSON(http.StatusNotFound, echo.Map{"error": "Stock not found"})
	}

	log.Printf("EditStock - Current stock details: %+v", stock)

	if err := c.Bind(&stock); err != nil {
		log.Printf("EditStock - Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	if err := db.GetDB().Save(&stock).Error; err != nil {
		log.Printf("EditStock - Save error: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	log.Println("EditStock - Stock updated successfully")
	log.Println("EditStock - Exit")
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
// func ViewAllStock(c echo.Context) error {
//     gormDB := db.GetDB()

//     // Ensure the database connection is properly configured to reflect recent changes
//     err := gormDB.Exec("SET SESSION TRANSACTION ISOLATION LEVEL READ COMMITTED").Error
//     if err != nil {
//         fmt.Printf("Failed to set isolation level: %v\n", err)
//         return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database configuration error"})
//     }

//     // Retrieve organizationID from context
//     organizationID, ok := c.Get("organizationID").(uint)
//     if !ok {
//         log.Println("ViewAllStock - Failed to get organizationID from context")
//         return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
//     }

//     log.Printf("ViewAllStock - OrganizationID: %d", organizationID)

//     // SQL Query with LEFT JOIN to include all stock items specific to an organization
//     query := `
//         SELECT 
//             s.id,
//             s.product_id,
//             p.product_name AS product_name,
//             s.quantity,
//             s.buying_price,
//             s.selling_price,
//             s.expiry_date,
//             p.product_description AS product_description,
//             su.name AS supplier_name
//         FROM stock s
//         LEFT JOIN products p ON s.product_id = p.product_id
//         LEFT JOIN suppliers su ON su.id = s.supplier_id
//         WHERE p.product_id IS NOT NULL  -- Ensures that we only get products that exist in the products table
//         AND s.organization_id = ?      -- Filter by organization_id
//     `

//     // Execute the query, passing the organizationID as a parameter
//     rows, err := gormDB.Raw(query, organizationID).Rows()
//     if err != nil {
//         fmt.Printf("Query execution failed: %v\nQuery: %s\n", err, query)
//         return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch stock"})
//     }
//     defer rows.Close()

//     var stocks []map[string]interface{}
//     for rows.Next() {
//         var (
//             id                 uint64
//             productID          uint64
//             productName        string
//             quantity           int
//             buyingPrice        float64
//             sellingPrice       float64
//             expiryDate         *string
//             productDescription string
//             supplierName       *string  // Supplier name may be null
//         )

//         err = rows.Scan(&id, &productID, &productName, &quantity, &buyingPrice, &sellingPrice, &expiryDate, &productDescription, &supplierName)
//         if err != nil {
//             fmt.Printf("Error scanning row: %v\n", err)
//             return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error reading stock data"})
//         }

//         stock := map[string]interface{}{
//             "id":                  id,
//             "product_id":          productID,
//             "product_name":        productName,
//             "quantity":            quantity,
//             "buying_price":        buyingPrice,
//             "selling_price":       sellingPrice,
//             "expiry_date":         expiryDate,
//             "product_description": productDescription,
//             "supplier_name":       supplierName,  // Supplier name might be null, hence *string
//         }

//         stocks = append(stocks, stock)
//     }

//     return c.JSON(http.StatusOK, stocks)
// }
func ViewAllStock(c echo.Context) error {
    gormDB := db.GetDB()

    // Set isolation level
    err := gormDB.Exec("SET SESSION TRANSACTION ISOLATION LEVEL READ COMMITTED").Error
    if err != nil {
        fmt.Printf("Failed to set isolation level: %v\n", err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database configuration error"})
    }

    organizationID, ok := c.Get("organizationID").(uint)
    if !ok {
        log.Println("ViewAllStock - Failed to get organizationID from context")
        return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
    }

    log.Printf("ViewAllStock - OrganizationID: %d", organizationID)

    // Modify the query to format the date as YYYYMMDD
    query := `
        SELECT 
            s.id,
            s.product_id,
            p.product_name AS product_name,
            s.quantity,
            s.buying_price,
            s.selling_price,
            DATE_FORMAT(s.expiry_date, '%Y-%m-%d') as expiry_date,
            p.product_description AS product_description,
            su.name AS supplier_name
        FROM stock s
        LEFT JOIN products p ON s.product_id = p.product_id
        LEFT JOIN suppliers su ON su.id = s.supplier_id
        WHERE p.product_id IS NOT NULL
        AND s.organization_id = ?
        ORDER BY s.id DESC
    `

    rows, err := gormDB.Raw(query, organizationID).Rows()
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
            quantity          int
            buyingPrice       float64
            sellingPrice      float64
            expiryDate        sql.NullString  
            productDescription string
            supplierName      sql.NullString  
        )

        err = rows.Scan(&id, &productID, &productName, &quantity, &buyingPrice, 
            &sellingPrice, &expiryDate, &productDescription, &supplierName)
        if err != nil {
            fmt.Printf("Error scanning row: %v\n", err)
            return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error reading stock data"})
        }

        // Create the stock map with proper null handling
        stock := map[string]interface{}{
            "id":                  id,
            "product_id":          productID,
            "product_name":        productName,
            "quantity":            quantity,
            "buying_price":        buyingPrice,
            "selling_price":       sellingPrice,
            "product_description": productDescription,
            "supplier_name":       nil,  
        }

        // Handle nullable fields
        if expiryDate.Valid {
            stock["expiry_date"] = expiryDate.String
        } else {
            stock["expiry_date"] = nil
        }

        if supplierName.Valid {
            stock["supplier_name"] = supplierName.String
        }

        stocks = append(stocks, stock)
    }

    return c.JSON(http.StatusOK, stocks)
}
func ViewStockByID(c echo.Context) error {
    log.Println("ViewStockByID - Entry")

    // Get stock ID from path
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        log.Printf("ViewStockByID - Invalid ID: %v", err)
        return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid stock ID"})
    }

    gormDB := db.GetDB()

    // Ensure the database connection is properly configured to reflect recent changes
    err = gormDB.Exec("SET SESSION TRANSACTION ISOLATION LEVEL READ COMMITTED").Error
    if err != nil {
        fmt.Printf("Failed to set isolation level: %v\n", err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database configuration error"})
    }

    // Retrieve organizationID from context
    organizationID, ok := c.Get("organizationID").(uint)
    if !ok {
        log.Println("ViewStockByID - Failed to get organizationID from context")
        return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
    }

    log.Printf("ViewStockByID - OrganizationID: %d", organizationID)

    // SQL Query with LEFT JOIN to include stock item details specific to an organization
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
            su.name AS supplier_name
        FROM stock s
        LEFT JOIN products p ON s.product_id = p.product_id
        LEFT JOIN suppliers su ON su.id = s.supplier_id
        WHERE s.id = ? 
        AND s.organization_id = ?  -- Filter by organization_id
    `

    // Execute the query, passing the stock ID and organization ID as parameters
    row := gormDB.Raw(query, id, organizationID).Row()
    if row == nil {
        log.Printf("ViewStockByID - Stock not found for ID: %d", id)
        return c.JSON(http.StatusNotFound, echo.Map{"error": "Stock not found"})
    }

    var (
        idVal                uint64
        productID            uint64
        productName          string
        quantity             int
        buyingPrice          float64
        sellingPrice         float64
        expiryDate           *string
        productDescription   string
        supplierName         *string  // Supplier name may be null
    )

    // Scan the result row
    err = row.Scan(&idVal, &productID, &productName, &quantity, &buyingPrice, &sellingPrice, &expiryDate, &productDescription, &supplierName)
    if err != nil {
        log.Printf("ViewStockByID - Error scanning row: %v", err)
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Error reading stock data"})
    }

    // Construct the response for the stock found
    stock := map[string]interface{}{
        "id":                  idVal,
        "product_id":          productID,
        "product_name":        productName,
        "quantity":            quantity,
        "buying_price":        buyingPrice,
        "selling_price":       sellingPrice,
        "expiry_date":         expiryDate,
        "product_description": productDescription,
        "supplier_name":       supplierName,
    }

    log.Println("ViewStockByID - Stock retrieved successfully")
    return c.JSON(http.StatusOK, stock)
}
