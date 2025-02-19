package controllers

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"database/sql"
	"stock/db"
	"strconv"

	"github.com/labstack/echo/v4"

	"stock/models"

	//"gorm.io/gorm"
)

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

	// var existingStock models.Stock
	// if err := db.GetDB().
	// 	Where("product_id = ? AND organization_id = ? AND deleted_at IS NULL", stock.ProductID, organizationID).
	// 	First(&existingStock).Error; err == nil {
	// 	log.Printf("CreateStock - Stock already exists: ProductID %d in OrganizationID %d", stock.ProductID, organizationID)
	// 	return c.JSON(http.StatusConflict, echo.Map{"error": "Stock already exists for this product in the organization"})
	// } else if err != gorm.ErrRecordNotFound {
	// 	log.Printf("CreateStock - Error checking existing stock: %v", err)
	// 	return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not check for existing stock"})
	// }

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

    query := `
        SELECT 
            s.id,
            s.product_id,
            p.product_name AS product_name,
            s.quantity,
            s.original_quantity,
            s.buying_price,
            s.selling_price,
            s.created_at,
            s.username,
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
            OriginalQuantity  int64
            buyingPrice       float64
            sellingPrice      float64
            created_at        time.Time
            username          string
            expiryDate        sql.NullString
            productDescription string
            supplierName      sql.NullString
        )

        err = rows.Scan(&id, &productID, &productName, &quantity, &OriginalQuantity, &buyingPrice, 
            &sellingPrice, &created_at, &username, &expiryDate, &productDescription, &supplierName)
        if err != nil {
            fmt.Printf("Error scanning row: %v\n", err)
            return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error reading stock data"})
        }

        stock := map[string]interface{}{
            "id":                  id,
            "product_id":          productID,
            "product_name":        productName,
            "quantity":            quantity,
            "original_quantity":   OriginalQuantity,
            "buying_price":        buyingPrice,
            "selling_price":       sellingPrice,
            "created_at":          created_at.Format("2006-01-02 15:04:05"),
            "username":             username,
            "product_description": productDescription,
            "supplier_name":       nil,
        }

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

    query := `
        SELECT 
            s.id,
            s.product_id,
            p.product_name AS product_name,
            s.quantity,
            s.original_quantity,
            s.buying_price,
            s.selling_price,
            s.created_at,
            s.expiry_date,
            s.username,
            p.product_description AS product_description,
            su.name AS supplier_name
        FROM stock s
        LEFT JOIN products p ON s.product_id = p.product_id
        LEFT JOIN suppliers su ON su.id = s.supplier_id
        WHERE s.id = ? 
        AND s.organization_id = ?
    `

    row := gormDB.Raw(query, id, organizationID).Row()
    if row == nil {
        log.Printf("ViewStockByID - Stock not found for ID: %d", id)
        return c.JSON(http.StatusNotFound, echo.Map{"error": "Stock not found"})
    }

    var (
        idVal               uint64
        productID          uint64
        productName        string
        quantity          int
        OriginalQuantity  int64
        buyingPrice       float64
        sellingPrice      float64
        created_at        time.Time
        expiryDate        *string
        username           string
        productDescription string
        supplierName      *string
    )

    err = row.Scan(&idVal, &productID, &productName, &quantity, &OriginalQuantity, &buyingPrice, 
        &sellingPrice, &created_at, &expiryDate, &username, &productDescription, &supplierName)
    if err != nil {
        log.Printf("ViewStockByID - Error scanning row: %v", err)
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Error reading stock data"})
    }

    stock := map[string]interface{}{
        "id":                  idVal,
        "product_id":          productID,
        "product_name":        productName,
        "quantity":            quantity,
        "original_quantity":   OriginalQuantity,
        "buying_price":        buyingPrice,
        "selling_price":       sellingPrice,
        "created_at":          created_at.Format("2006-01-02 15:04:05"),
        "expiry_date":         expiryDate,
        "username":            username,
        "product_description": productDescription,
        "supplier_name":       supplierName,
    }

    log.Println("ViewStockByID - Stock retrieved successfully")
    return c.JSON(http.StatusOK, stock)
}

// func ViewTotalPurchasedStock(c echo.Context) error {
//     gormDB := db.GetDB()

//     // Set isolation level
//     err := gormDB.Exec("SET SESSION TRANSACTION ISOLATION LEVEL READ COMMITTED").Error
//     if err != nil {
//         fmt.Printf("Failed to set isolation level: %v\n", err)
//         return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database configuration error"})
//     }

//     organizationID, ok := c.Get("organizationID").(uint)
//     if !ok {
//         log.Println("ViewTotalPurchasedStock - Failed to get organizationID from context")
//         return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
//     }

//     log.Printf("ViewTotalPurchasedStock - OrganizationID: %d", organizationID)

//     query := `
//         SELECT 
//             s.product_id,
//             p.product_name AS product_name,
//             SUM(s.original_quantity) AS total_purchased,  -- Changed to original_quantity
//             s.buying_price,
//             s.selling_price,
//             s.username,
//             p.product_description AS product_description,
//             su.name AS supplier_name,
//             s.created_at
//         FROM stock s
//         LEFT JOIN products p ON s.product_id = p.product_id
//         LEFT JOIN suppliers su ON su.id = s.supplier_id
//         WHERE p.product_id IS NOT NULL
//         AND s.organization_id = ?
//         GROUP BY s.product_id, p.product_name, s.buying_price, s.selling_price, p.product_description, su.name
//         ORDER BY p.product_name ASC
//     `

//     rows, err := gormDB.Raw(query, organizationID).Rows()
//     if err != nil {
//         fmt.Printf("Query execution failed: %v\nQuery: %s\n", err, query)
//         return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch total purchased stock"})
//     }
//     defer rows.Close()

//     var purchasedStocks []map[string]interface{}
//     for rows.Next() {
//         var (
//             productID          uint64
//             productName        string
//             totalPurchased     int
//             buyingPrice        float64
//             sellingPrice       float64
//             username            string
//             productDescription string
//             supplierName       sql.NullString
//             createdAt          time.Time
//         )

//         err = rows.Scan(&productID, &productName, &totalPurchased, &buyingPrice,
//             &sellingPrice, &username, &productDescription, &supplierName, &createdAt)
//         if err != nil {
//             fmt.Printf("Error scanning row: %v\n", err)
//             return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error reading purchased stock data"})
//         }

//         stock := map[string]interface{}{
//             "product_id":          productID,
//             "product_name":        productName,
//             "total_purchased":     totalPurchased,
//             "buying_price":        buyingPrice,
//             "selling_price":       sellingPrice,
//             "username":            username,
//             "product_description": productDescription,
//             "supplier_name":       nil,
//             "created_at":          createdAt.Format("2006-01-02 15:04:05"),
//         }

//         if supplierName.Valid {
//             stock["supplier_name"] = supplierName.String
//         }

//         purchasedStocks = append(purchasedStocks, stock)
//     }

//     return c.JSON(http.StatusOK, purchasedStocks)
// }
func ViewTotalPurchasedStock(c echo.Context) error {
    gormDB := db.GetDB()
    
    // Set isolation level
    err := gormDB.Exec("SET SESSION TRANSACTION ISOLATION LEVEL READ COMMITTED").Error
    if err != nil {
        fmt.Printf("Failed to set isolation level: %v\n", err)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Database configuration error"})
    }
    
    organizationID, ok := c.Get("organizationID").(uint)
    if !ok {
        log.Println("ViewTotalPurchasedStock - Failed to get organizationID from context")
        return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
    }
    
    log.Printf("ViewTotalPurchasedStock - OrganizationID: %d", organizationID)
    
    query := `
        SELECT 
            s.product_id,
            p.product_name AS product_name,
            COALESCE(SUM(s.original_quantity), 0) AS total_purchased,  -- Handle NULL with COALESCE
            s.buying_price,
            s.selling_price,
            s.username,
            p.product_description AS product_description,
            su.name AS supplier_name,
            s.created_at
        FROM stock s
        LEFT JOIN products p ON s.product_id = p.product_id
        LEFT JOIN suppliers su ON su.id = s.supplier_id
        WHERE p.product_id IS NOT NULL
        AND s.organization_id = ?
        GROUP BY s.product_id, p.product_name, s.buying_price, s.selling_price, s.username, p.product_description, su.name, s.created_at
        ORDER BY p.product_name ASC
    `
    
    rows, err := gormDB.Raw(query, organizationID).Rows()
    if err != nil {
        fmt.Printf("Query execution failed: %v\nQuery: %s\n", err, query)
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to fetch total purchased stock"})
    }
    defer rows.Close()
    
    var purchasedStocks []map[string]interface{}
    for rows.Next() {
        var (
            productID          uint64
            productName        string
            totalPurchased     int
            buyingPrice        float64
            sellingPrice       float64
            username           string
            productDescription string
            supplierName       sql.NullString
            createdAt          time.Time
        )
        
        err = rows.Scan(&productID, &productName, &totalPurchased, &buyingPrice,
            &sellingPrice, &username, &productDescription, &supplierName, &createdAt)
        if err != nil {
            fmt.Printf("Error scanning row: %v\n", err)
            return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Error reading purchased stock data"})
        }
        
        stock := map[string]interface{}{
            "product_id":          productID,
            "product_name":        productName,
            "total_purchased":     totalPurchased,
            "buying_price":        buyingPrice,
            "selling_price":       sellingPrice,
            "username":            username,
            "product_description": productDescription,
            "supplier_name":       nil,
            "created_at":          createdAt.Format("2006-01-02 15:04:05"),
        }
        
        if supplierName.Valid {
            stock["supplier_name"] = supplierName.String
        }
        
        purchasedStocks = append(purchasedStocks, stock)
    }
    
    return c.JSON(http.StatusOK, purchasedStocks)
}
func AddPurchases(c echo.Context) error {
	log.Println("AddPurchases - Entry")

	// Retrieve the user's role and organization ID from the context
	roleName, ok := c.Get("roleName").(string)
	if !ok {
		log.Println("AddPurchases - Unauthorized: roleName not found in context")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	organizationIDRaw := c.Get("organizationID")
	if organizationIDRaw == nil {
		log.Println("AddPurchases - Unauthorized: organizationID not found in context")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	organizationID, ok := organizationIDRaw.(uint)
	if !ok {
		log.Println("AddPurchases - Unauthorized: organizationID is not of type uint")
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
	}

	log.Printf("AddPurchases - User Role: %s, OrganizationID: %d", roleName, organizationID)

	// Bind request body to the Stock model (we use the same Stock struct for purchases)
	var stock models.Stock
	if err := c.Bind(&stock); err != nil {
		log.Printf("AddPurchases - Bind error: %v", err)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}

	var existingSupplier models.Suppliers
	if err := db.GetDB().
		Where("id = ?", stock.SupplierID).
		First(&existingSupplier).Error; err != nil {
		log.Printf("AddPurchases - Supplier with ID %d not found", stock.SupplierID)
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid supplier ID"})
	}

	// ✅ Set the OrganizationID to the user's organization (stock/purchase entry association)
	stock.OrganizationID = organizationID


	// stock.OriginalQuantity = stock.Quantity 

	if err := db.GetDB().Create(&stock).Error; err != nil {
		log.Printf("AddPurchases - Error adding purchase to stock: %v", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Failed to add purchase to stock"})
	}

	log.Println("AddPurchases - Purchase added successfully to stock")
	log.Println("AddPurchases - Exit")
	return c.JSON(http.StatusOK, echo.Map{"message": "Purchase added to stock successfully"})
}


