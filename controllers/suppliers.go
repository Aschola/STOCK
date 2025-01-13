package controllers

import (
	"log"
	"net/http"
	"stock/models"
	"stock/db"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

// CreateSupplier handles the creation of a new supplier
func AddSupplier(c echo.Context) error {
    log.Println("CreateSupplier - Entry")

    // Retrieve the organization ID from the context
    organizationIDRaw := c.Get("organizationID")
    if organizationIDRaw == nil {
        log.Println("CreateSupplier - organizationID not found in context")
        return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
    }

    organizationID, ok := organizationIDRaw.(uint)
    if !ok {
        log.Println("CreateSupplier - Invalid organizationID")
        return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
    }

    // Bind the incoming request body to the supplier struct
    var supplier models.Suppliers
    if err := c.Bind(&supplier); err != nil {
        log.Printf("CreateSupplier - Bind error: %v", err)
        return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
    }

    // Check if the supplier already exists (based on a unique field like 'name' or 'contact')
    var existingSupplier models.Suppliers
    if err := db.GetDB().Where("name = ? AND organization_id = ?", supplier.Name, organizationID).First(&existingSupplier).Error; err == nil {
        log.Printf("CreateSupplier - Supplier with name %s already exists in organization %d", supplier.Name, organizationID)
        return c.JSON(http.StatusConflict, echo.Map{"error": "Supplier already exists"})
    }

    // Retrieve userID from the context
    userIDRaw := c.Get("userID")
    if userIDRaw == nil {
        log.Println("CreateSupplier - userID not found in context")
        return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
    }

    userID, ok := userIDRaw.(uint)
    if !ok {
        log.Println("CreateSupplier - Invalid userID")
        return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
    }

    // Automatically link the supplier with the current user's organization
    supplier.OrganizationID = organizationID
    supplier.CreatedBy = userID // Set the user who created the supplier

    log.Printf("CreateSupplier - New supplier data: %+v", supplier)

    // Save the new supplier to the database
    if err := db.GetDB().Create(&supplier).Error; err != nil {
        log.Printf("CreateSupplier - Create error: %v", err)
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
    }

    log.Println("CreateSupplier - Supplier created successfully")
    log.Println("CreateSupplier - Exit")
    return c.JSON(http.StatusCreated, supplier)
}

// UpdateSupplier handles updating an existing supplier
func EditSupplier(c echo.Context) error {
    log.Println("UpdateSupplier - Entry")

    roleName, ok := c.Get("roleName").(string)
    if !ok {
        log.Println("UpdateSupplier - Unauthorized: roleName not found in context")
        return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
    }

    if roleName != "Admin" {
        log.Println("UpdateSupplier - Permission denied: non-admin trying to update supplier")
        return c.JSON(http.StatusForbidden, echo.Map{"error": "Permission denied"})
    }

    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        log.Printf("UpdateSupplier - Invalid ID: %v", err)
        return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid supplier ID"})
    }

    // Find existing supplier
    var existingSupplier models.Suppliers
    if err := db.GetDB().First(&existingSupplier, id).Error; err != nil {
        log.Printf("UpdateSupplier - Supplier not found: %v", err)
        return c.JSON(http.StatusNotFound, echo.Map{"error": "Supplier not found"})
    }

    var updateData models.Suppliers
    if err := c.Bind(&updateData); err != nil {
        log.Printf("UpdateSupplier - Bind error: %v", err)
        return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
    }

    // Update fields
    existingSupplier.Name = updateData.Name
    existingSupplier.Phonenumber = updateData.Phonenumber
    existingSupplier.UpdatedAt = time.Now()

    // Save updates
    if err := db.GetDB().Save(&existingSupplier).Error; err != nil {
        log.Printf("UpdateSupplier - Update error: %v", err)
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
    }

    log.Println("UpdateSupplier - Supplier updated successfully")
    log.Println("UpdateSupplier - Exit")
    return c.JSON(http.StatusOK, existingSupplier)
}

// DeleteSupplier handles supplier deletion (soft delete)
func DeleteSupplier(c echo.Context) error {
    log.Println("DeleteSupplier - Entry")

    // Check role authorization
    roleName, ok := c.Get("roleName").(string)
    if !ok {
        log.Println("DeleteSupplier - Unauthorized: roleName not found in context")
        return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
    }

    if roleName != "Admin" {
        log.Println("DeleteSupplier - Permission denied: non-admin trying to delete supplier")
        return c.JSON(http.StatusForbidden, echo.Map{"error": "Permission denied"})
    }

    // Get supplier ID from path
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        log.Printf("DeleteSupplier - Invalid ID: %v", err)
        return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid supplier ID"})
    }

    // Soft delete the supplier
    if err := db.GetDB().Delete(&models.Suppliers{}, id).Error; err != nil {
        log.Printf("DeleteSupplier - Delete error: %v", err)
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
    }

    log.Println("DeleteSupplier - Supplier deleted successfully")
    log.Println("DeleteSupplier - Exit")
    return c.JSON(http.StatusOK, echo.Map{"message": "Supplier deleted successfully"})
}

// GetAllSuppliers retrieves all suppliers
// func GetAllSuppliers(c echo.Context) error {
//     log.Println("GetAllSuppliers - Entry")

//     var suppliers []models.Suppliers

//     // Get all suppliers, excluding soft-deleted ones
//     if err := db.GetDB().Find(&suppliers).Error; err != nil {
//         log.Printf("GetAllSuppliers - Query error: %v", err)
//         return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
//     }

//     log.Printf("GetAllSuppliers - Retrieved %d suppliers", len(suppliers))
//     log.Println("GetAllSuppliers - Exit")
//     return c.JSON(http.StatusOK, suppliers)
// }

func GetAllSuppliers(c echo.Context) error {
    log.Println("GetAllSuppliers - Entry")

    // Retrieve organizationID from the context (from token)
    organizationIDRaw := c.Get("organizationID")
    if organizationIDRaw == nil {
        log.Println("GetAllSuppliers - organizationID is not set in the context")
        return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
    }

    organizationID, ok := organizationIDRaw.(uint)
    if !ok {
        log.Println("GetAllSuppliers - organizationID is not of type uint")
        return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
    }

    // Log the organization ID
    log.Printf("GetAllSuppliers - OrganizationID: %d", organizationID)

    // Query suppliers, stock, and product details (product_name) linked to the organization
    query := `
        SELECT s.id, s.name, s.organization_id, s.created_at, s.updated_at, p.product_name 
        FROM suppliers s
        LEFT JOIN stock st ON st.supplier_id = s.id
        LEFT JOIN products p ON p.product_id = st.product_id
        WHERE s.organization_id = ? AND s.deleted_at IS NULL
    `
    var suppliers []map[string]interface{}

    rows, err := db.GetDB().Raw(query, organizationID).Rows()
    if err != nil {
        log.Printf("GetAllSuppliers - Query error: %v", err)
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not retrieve suppliers"})
    }
    defer rows.Close()

    for rows.Next() {
        var supplierID uint
        var supplierName string
        var organizationID uint
        var createdAt, updatedAt *string
        var productName string

        if err := rows.Scan(&supplierID, &supplierName, &organizationID, &createdAt, &updatedAt, &productName); err != nil {
            log.Printf("GetAllSuppliers - Row scan error: %v", err)
            return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Error scanning suppliers data"})
        }

        supplier := map[string]interface{}{
            "id":             supplierID,
            "name":           supplierName,
            "organization_id": organizationID,
            "created_at":     createdAt,
            "updated_at":     updatedAt,
            "product_name":   productName,
        }

        suppliers = append(suppliers, supplier)
    }

    // Return suppliers as an array
    log.Printf("GetAllSuppliers - Retrieved %d suppliers", len(suppliers))
    log.Println("GetAllSuppliers - Exit")
    return c.JSON(http.StatusOK, suppliers)
}

// GetSupplier retrieves a single supplier by ID
func GetSupplier(c echo.Context) error {
    log.Println("GetSupplier - Entry")

    // Get supplier ID from path
    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        log.Printf("GetSupplier - Invalid ID: %v", err)
        return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid supplier ID"})
    }

    // Retrieve organizationID from the context
    organizationIDRaw := c.Get("organizationID")
    if organizationIDRaw == nil {
        log.Println("GetSupplier - organizationID not found in context")
        return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
    }

    organizationID, ok := organizationIDRaw.(uint)
    if !ok {
        log.Println("GetSupplier - Invalid organizationID")
        return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
    }

    // Query supplier, stock, and product details (product_name) linked to the supplier
    query := `
        SELECT s.id, s.name, s.organization_id, s.created_at, s.updated_at, p.product_name 
        FROM suppliers s
        LEFT JOIN stock st ON st.supplier_id = s.id
        LEFT JOIN products p ON p.product_id = st.product_id
        WHERE s.id = ? AND s.organization_id = ? AND s.deleted_at IS NULL
    `
    var supplierDetails map[string]interface{}
    row := db.GetDB().Raw(query, id, organizationID).Row()

    // Scan the result into the map
    var supplierID uint
    var supplierName string
    var createdAt, updatedAt *string
    var productName string

    if err := row.Scan(&supplierID, &supplierName, &organizationID, &createdAt, &updatedAt, &productName); err != nil {
        log.Printf("GetSupplier - Supplier not found: %v", err)
        return c.JSON(http.StatusNotFound, echo.Map{"error": "Supplier not found"})
    }

    // Prepare the response with supplier and product details
    supplierDetails = map[string]interface{}{
        "id":             supplierID,
        "name":           supplierName,
        "organization_id": organizationID,
        "created_at":     createdAt,
        "updated_at":     updatedAt,
        "product_name":   productName,
    }

    log.Println("GetSupplier - Supplier retrieved successfully")
    log.Println("GetSupplier - Exit")
    return c.JSON(http.StatusOK, supplierDetails)
}

