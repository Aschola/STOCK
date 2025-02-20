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
// func AddSupplier(c echo.Context) error {
//     log.Println("CreateSupplier - Entry")

//     // Retrieve the organization ID from the context
//     organizationIDRaw := c.Get("organizationID")
//     if organizationIDRaw == nil {
//         log.Println("CreateSupplier - organizationID not found in context")
//         return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
//     }

//     organizationID, ok := organizationIDRaw.(uint)
//     if !ok {
//         log.Println("CreateSupplier - Invalid organizationID")
//         return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
//     }

//     // Bind the incoming request body to the supplier struct
//     var supplier models.Suppliers
//     if err := c.Bind(&supplier); err != nil {
//         log.Printf("CreateSupplier - Bind error: %v", err)
//         return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
//     }

//     // Check if the supplier already exists (based on a unique field like 'name' or 'contact')
//     var existingSupplier models.Suppliers
//     if err := db.GetDB().Where("name = ? AND organization_id = ?", supplier.Name, organizationID).First(&existingSupplier).Error; err == nil {
//         log.Printf("CreateSupplier - Supplier with name %s already exists in organization %d", supplier.Name, organizationID)
//         return c.JSON(http.StatusConflict, echo.Map{"error": "Supplier already exists"})
//     }

//     // Retrieve userID from the context
//     userIDRaw := c.Get("userID")
//     if userIDRaw == nil {
//         log.Println("CreateSupplier - userID not found in context")
//         return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
//     }

//     userID, ok := userIDRaw.(uint)
//     if !ok {
//         log.Println("CreateSupplier - Invalid userID")
//         return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
//     }

//     // Automatically link the supplier with the current user's organization
//     supplier.OrganizationID = organizationID
//     supplier.CreatedBy = userID // Set the user who created the supplier

//     log.Printf("CreateSupplier - New supplier data: %+v", supplier)

//     // Save the new supplier to the database
//     if err := db.GetDB().Create(&supplier).Error; err != nil {
//         log.Printf("CreateSupplier - Create error: %v", err)
//         return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
//     }

//     log.Println("CreateSupplier - Supplier created successfully")
//     log.Println("CreateSupplier - Exit")
//     return c.JSON(http.StatusCreated, supplier)
// }
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

    // Check if the supplier already exists by name
    var existingSupplier models.Suppliers
    if err := db.GetDB().Where("name = ? AND organization_id = ?", supplier.Name, organizationID).First(&existingSupplier).Error; err == nil {
        log.Printf("CreateSupplier - Supplier with name %s already exists in organization %d", supplier.Name, organizationID)
        return c.JSON(http.StatusConflict, echo.Map{"error": "Supplier already exists"})
    }

    // Check if the phone number already exists
    var existingPhone models.Suppliers
    if err := db.GetDB().Where("phone_number = ? AND organization_id = ?", supplier.Phonenumber, organizationID).First(&existingPhone).Error; err == nil {
        log.Printf("CreateSupplier - Supplier with phone number %s already exists in organization %d", supplier.Phonenumber, organizationID)
        return c.JSON(http.StatusConflict, echo.Map{"error": "Phone number already exists"})
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
    supplier.CreatedBy = userID

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
    existingSupplier.CompanyName = updateData.CompanyName
    existingSupplier.Address = updateData.Address
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

// Controller function
// func GetAllSuppliers(c echo.Context) error {
//     log.Println("GetAllSuppliers - Entry")

//     // Retrieve organizationID from the context (from token)
//     organizationIDRaw := c.Get("organizationID")
//     if organizationIDRaw == nil {
//         log.Println("GetAllSuppliers - organizationID is not set in the context")
//         return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
//     }

//     organizationID, ok := organizationIDRaw.(uint)
//     if !ok {
//         log.Println("GetAllSuppliers - organizationID is not of type uint")
//         return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
//     }

//     // Log the organization ID
//     log.Printf("GetAllSuppliers - OrganizationID: %d", organizationID)

//     // Query suppliers linked to the organization (without product_name)
//     query := `
//         SELECT s.id AS supplier_id,
//                s.name AS supplier_name,
//                s.phone_number,
//                s.organization_id,
//                s.address,
//                s.company_name,
//                s.created_at,
//                s.deleted_at
//         FROM suppliers s
//         LEFT JOIN stock st ON st.supplier_id = s.id -- Ensure supplier is included even if they don't have stock
//         WHERE s.organization_id = ? AND s.deleted_at IS NULL
//     `

//     var suppliers []map[string]interface{}

//     // Execute the query and pass the organizationID as a parameter
//     rows, err := db.GetDB().Raw(query, organizationID).Rows()
//     if err != nil {
//         log.Printf("GetAllSuppliers - Query error: %v", err)
//         return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not retrieve suppliers"})
//     }
//     defer rows.Close()

//     // Iterate through rows and scan the data into a slice
//     for rows.Next() {
//         var supplierID uint
//         var supplierName, phoneNumber string
//         var Companyname string
//         var address string
//         var orgID uint
//         var createdAt, deletedAt *string 

//         // Scan row data into variables
//         if err := rows.Scan(&supplierID, &supplierName, &phoneNumber, &Companyname, &address, &orgID, &createdAt, &deletedAt); err != nil {
//             log.Printf("GetAllSuppliers - Row scan error: %v", err)
//             return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Error scanning suppliers data"})
//         }

//         // Construct supplier data
//         supplier := map[string]interface{}{
//             "id":              supplierID,
//             "name":            supplierName,
//             "phone_number":    phoneNumber,
//             "address":         address,
//             "company_name":    Companyname,
//             "organization_id": orgID,
//             "created_at":      createdAt,
//             "deleted_at":      deletedAt,
//         }

//         // Append the supplier to the list
//         suppliers = append(suppliers, supplier)
//     }

//     // Return the suppliers data as a response
//     log.Printf("GetAllSuppliers - Retrieved %d suppliers", len(suppliers))
//     return c.JSON(http.StatusOK, suppliers)
// }


// // GetSupplier retrieves a single supplier by ID
// func GetSupplier(c echo.Context) error {
//     log.Println("GetSupplier - Entry")

//     // Get supplier ID from path
//     id, err := strconv.Atoi(c.Param("id"))
//     if err != nil {
//         log.Printf("GetSupplier - Invalid ID: %v", err)
//         return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid supplier ID"})
//     }

//     // Retrieve organizationID from the context
//     organizationIDRaw := c.Get("organizationID")
//     if organizationIDRaw == nil {
//         log.Println("GetSupplier - organizationID not found in context")
//         return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
//     }

//     organizationID, ok := organizationIDRaw.(uint)
//     if !ok {
//         log.Println("GetSupplier - Invalid organizationID")
//         return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Unauthorized"})
//     }

//     // Query supplier details without product_name
//     query := `
//         SELECT 
//             s.id AS supplier_id,
//             s.name AS supplier_name,
//             s.phone_number,
//             s.organization_id,
//             s.address,
//             s.company_name,
//             s.created_at,
//             s.deleted_at
//         FROM suppliers s
//         WHERE s.id = ? AND s.organization_id = ? AND s.deleted_at IS NULL
//     `

//     var supplierDetails []map[string]interface{}
//     rows, err := db.GetDB().Raw(query, id, organizationID).Rows()
//     if err != nil {
//         log.Printf("GetSupplier - Query error: %v", err)
//         return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not retrieve supplier"})
//     }
//     defer rows.Close()

//     for rows.Next() {
//         var supplierID uint
//         var supplierName, phoneNumber *string
//         var orgID uint
//         var createdAt, deletedAt *string
//         var Companyname string
//         var address string

//         if err := rows.Scan(&supplierID, &supplierName, &phoneNumber, &Companyname, &address, &orgID, &createdAt, &deletedAt); err != nil {
//             log.Printf("GetSupplier - Row scan error: %v", err)
//             return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Error scanning supplier data"})
//         }

//         supplier := map[string]interface{}{
//             "id":              supplierID,
//             "name":            supplierName,
//             "phone_number":    phoneNumber,
//             "organization_id": orgID,
//             "address":         address,
//             "company_name":    Companyname,
//             "created_at":      createdAt,
//             "deleted_at":      deletedAt,
//         }

//         supplierDetails = append(supplierDetails, supplier)
//     }

//     // Check if no suppliers were found
//     if len(supplierDetails) == 0 {
//         log.Println("GetSupplier - No supplier found for the given ID")
//         return c.JSON(http.StatusNotFound, echo.Map{"error": "Supplier not found"})
//     }

//     log.Println("GetSupplier - Supplier retrieved successfully")
//     log.Println("GetSupplier - Exit")
//     return c.JSON(http.StatusOK, supplierDetails)
// }
func GetAllSuppliers(c echo.Context) error {
    log.Println("GetAllSuppliers - Entry")

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

    log.Printf("GetAllSuppliers - OrganizationID: %d", organizationID)

    query := `
        SELECT 
            s.id AS supplier_id,
            s.name AS supplier_name,
            s.phone_number,
            s.organization_id,
            s.company_name,
            s.address,
            s.created_at,
            s.deleted_at,
            s.description
        FROM suppliers s
        LEFT JOIN stock st ON st.supplier_id = s.id
        WHERE s.organization_id = ? AND s.deleted_at IS NULL
        GROUP BY s.id
    `

    var suppliers []map[string]interface{}

    rows, err := db.GetDB().Raw(query, organizationID).Rows()
    if err != nil {
        log.Printf("GetAllSuppliers - Query error: %v", err)
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not retrieve suppliers"})
    }
    defer rows.Close()

    for rows.Next() {
        var (
            supplierID     uint
            supplierName   *string
            phoneNumber    *string
            orgID         uint
            company_name  *string
            address       *string
            createdAt     *string
            deletedAt     *string
            description   *string
        )

        if err := rows.Scan(
            &supplierID,
            &supplierName,
            &phoneNumber,
            &orgID,
            &company_name,
            &address,
            &createdAt,
            &deletedAt,
            &description,
        ); err != nil {
            log.Printf("GetAllSuppliers - Row scan error: %v", err)
            return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Error scanning suppliers data"})
        }

        supplier := map[string]interface{}{
            "id":              supplierID,
            "name":            nilStringToEmpty(supplierName),
            "phone_number":    nilStringToEmpty(phoneNumber),
            "organization_id": orgID,
            "company_name":    nilStringToEmpty(company_name),
            "address":         nilStringToEmpty(address),
            "created_at":      createdAt,
            "deleted_at":      deletedAt,
            "description":     nilStringToEmpty(description),
        }

        suppliers = append(suppliers, supplier)
    }

    log.Printf("GetAllSuppliers - Retrieved %d suppliers", len(suppliers))
    return c.JSON(http.StatusOK, suppliers)
}

func GetSupplier(c echo.Context) error {
    log.Println("GetSupplier - Entry")

    id, err := strconv.Atoi(c.Param("id"))
    if err != nil {
        log.Printf("GetSupplier - Invalid ID: %v", err)
        return c.JSON(http.StatusBadRequest, echo.Map{"error": "Invalid supplier ID"})
    }

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

    query := `
        SELECT 
            s.id AS supplier_id,
            s.name AS supplier_name,
            s.phone_number,
            s.organization_id,
            s.company_name,
            s.address,
            s.created_at,
            s.deleted_at,
            s.description
        FROM suppliers s
        WHERE s.id = ? AND s.organization_id = ? AND s.deleted_at IS NULL
    `

    var supplierDetails []map[string]interface{}
    rows, err := db.GetDB().Raw(query, id, organizationID).Rows()
    if err != nil {
        log.Printf("GetSupplier - Query error: %v", err)
        return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Could not retrieve supplier"})
    }
    defer rows.Close()

    for rows.Next() {
        var (
            supplierID   uint
            supplierName *string
            phoneNumber  *string
            orgID       uint
            company_name *string
            address     *string
            createdAt   *string
            deletedAt   *string
            description *string
        )

        if err := rows.Scan(
            &supplierID,
            &supplierName,
            &phoneNumber,
            &orgID,
            &company_name,
            &address,
            &createdAt,
            &deletedAt,
            &description,
        ); err != nil {
            log.Printf("GetSupplier - Row scan error: %v", err)
            return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Error scanning supplier data"})
        }

        supplier := map[string]interface{}{
            "id":              supplierID,
            "name":            nilStringToEmpty(supplierName),
            "phone_number":    nilStringToEmpty(phoneNumber),
            "organization_id": orgID,
            "company_name":    nilStringToEmpty(company_name),
            "address":         nilStringToEmpty(address),
            "created_at":      createdAt,
            "deleted_at":      deletedAt,
            "description":     nilStringToEmpty(description),
        }

        supplierDetails = append(supplierDetails, supplier)
    }

    if len(supplierDetails) == 0 {
        log.Println("GetSupplier - No supplier found for the given ID")
        return c.JSON(http.StatusNotFound, echo.Map{"error": "Supplier not found"})
    }

    log.Println("GetSupplier - Supplier retrieved successfully")
    return c.JSON(http.StatusOK, supplierDetails[0])
}

// Helper function to convert nil string pointers to empty strings
func nilStringToEmpty(s *string) string {
    if s == nil {
        return ""
    }
    return *s
}