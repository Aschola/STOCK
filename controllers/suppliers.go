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

    roleName, ok := c.Get("roleName").(string)
    if !ok {
        log.Println("CreateSupplier - Unauthorized: roleName not found in context")
        return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
    }

    if roleName != "Admin" {
        log.Println("CreateSupplier - Permission denied: non-admin trying to create supplier")
        return c.JSON(http.StatusForbidden, echo.Map{"error": "Permission denied"})
    }

    // Retrieve organizationID from the context (i.e. the organization that the admin is linked to)
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

    // Bind request body to supplier struct
    var supplier models.Suppliers
    if err := c.Bind(&supplier); err != nil {
        log.Printf("CreateSupplier - Bind error: %v", err)
        return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
    }

    supplier.CreatedBy = c.Get("userID").(uint)  // Set the user that created the supplier
    supplier.OrganizationID = organizationID    // Set the supplier's organization to the logged-in admin's organization

    log.Printf("CreateSupplier - New supplier data: %+v", supplier)

    // Save to database
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

    // Query suppliers linked to the organization, excluding deleted suppliers
    var suppliers []models.Suppliers
    if err := db.GetDB().
        Where("organization_id = ? AND deleted_at IS NULL", organizationID).
        Find(&suppliers).Error; err != nil {
        log.Printf("GetAllSuppliers - Query error: %v", err)
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not retrieve suppliers"})
    }

    // Return each supplier individually, not as an array
    for _, supplier := range suppliers {
        if err := c.JSON(http.StatusOK, supplier); err != nil {
            log.Printf("GetAllSuppliers - Error returning supplier: %v", err)
            return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Error returning suppliers"})
        }
    }

    log.Println("GetAllSuppliers - Exit")
    return nil
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

    var supplier models.Suppliers
    if err := db.GetDB().First(&supplier, id).Error; err != nil {
        log.Printf("GetSupplier - Supplier not found: %v", err)
        return c.JSON(http.StatusNotFound, echo.Map{"error": "Supplier not found"})
    }

    // Check if the supplier belongs to the correct organization
    if supplier.OrganizationID != organizationID {
        log.Println("GetSupplier - Supplier does not belong to this organization")
        return c.JSON(http.StatusForbidden, echo.Map{"error": "Supplier not found in your organization"})
    }

    log.Println("GetSupplier - Supplier retrieved successfully")
    log.Println("GetSupplier - Exit")
    return c.JSON(http.StatusOK, supplier)  // Single JSON object response
}
