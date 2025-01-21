package routes

import (
	"stock/controllers"
	"stock/middlewares"
	"stock/models"

	"github.com/labstack/echo/v4"
)

// RegisterRoutes initializes all the routes for the Echo server
func RegisterRoutes(e *echo.Echo) {
	//productGroup := e.Group("/products")
	// productGroup.GET("", controllers.GetProducts)                  // Fetch all products
	// productGroup.GET("/:product_id", controllers.GetProductByID)   // Fetch a single product by ID
	//productGroup.POST("", controllers.AddProduct)               // Add a new product
	//productGroup.PUT("/:product_id", controllers.UpdateProduct) // Update product details
	//productGroup.DELETE("/:product_id", controllers.DeleteProduct) // Delete a product

	// Define CRUD endpoints for categories without admin middleware
	//categoryGroup := e.Group("/categories")
	// categoryGroup.GET("", controllers.GetCategories)
	// categoryGroup.GET("/:category_id", controllers.GetCategoryByID)
	// categoryGroup.POST("", controllers.CreateCategories)
	// categoryGroup.PUT("/:category_id", controllers.UpdateCategory)
	// categoryGroup.DELETE("/:id", controllers.DeleteCategoryByID)
	// categoryGroup.GET("/only", controllers.GetCategoriesOnly)
	//e.GET("/products/category/:id", controllers.GetProductsByCategoryID)

	//e.POST("/categories_only", controllers.CreateCategoryInCategoriesOnly)
	//e.GET("/categories_only/:id", controllers.GetCategoryNameByID)
	//e.POST("/categories_only", controllers.Categories_Only)

	//e.DELETE("/categories_onlies/:id", controllers.DeleteCategoryFromCategoriesOnly)

	// Define CRUD endpoints for sales
	// Define the new route in main.go or wherever you define your routes.

	e.GET("/cash/sales", controllers.GetAllSales)

	//e.GET("/salebycategory/:category_name", controllers.FetchSalesByCategory)
	// Register the route to get sales by date

	e.GET("/salebycategory/:user_id", controllers.FetchSalesByUserID)
	// Endpoint for selling products
	e.POST("/products/:product_id/sell/:quantity_sold", controllers.SellProduct)
	e.GET("/sales/:sale_id", controllers.GetSalesBySaleID)
}

func SetupRoutes(e *echo.Echo) {
	// Public routes
	e.POST("admin/signup", controllers.AdminSignup)
	e.POST("/superadmin/login", controllers.SuperAdminLogin)
	e.POST("/superadmin/logout", controllers.SuperAdminLogout)
	e.POST("admin/login", controllers.AdminLogin)
	e.POST("admin/logout", controllers.AdminLogout)
	e.POST("shopkeeper/login", controllers.Login)
	e.POST("shopkeeper/logout", controllers.Logout)
	e.POST("auditor/login", controllers.AuditorLogin)
	e.POST("auditor/logout", controllers.AuditorLogout)
	//e.POST("/addsupplier", controllers.AddSupplier)
	// e.PUT("/editsupplier/:id", controllers.EditSupplier)
	// e.DELETE("/deletesupplier", controllers.DeleteSupplier)
	// e.GET("/viewallsuppliers", controllers.GetAllSuppliers)
	// e.GET("/viewsupplier/:id", controllers.GetSupplier)
	// e.POST("/createstock", controllers.CreateStock)
	// e.DELETE("/deletestock/:id", controllers.DeleteStock)
	// e.PUT("/editstock", controllers.EditStock)
	// e.GET("/viewallstock", controllers.ViewAllStock)
	// e.GET("/viewstock/:id", controllers.ViewStockByID)
	//e.POST("/organizationadmin/login", controllers.OrganizationAdminLogin)
	//e.POST("/organizationadmin/logout", controllers.OrganizationAdminLogout)
	e.POST("/forgot-password", controllers.ForgotPassword)
	e.POST("/reset-password", controllers.ResetPassword)

	// Super Admin routes
	superadmin := e.Group("/superadmin")
	superadmin.POST("/addadmin", controllers.AdminSignup)
	superadmin.Use(middlewares.AuthMiddleware(models.SuperAdminRoleName))
	superadmin.POST("/addadmin", controllers.AddAdmin)
	//superadmin.POST("/addorganization", controllers.SuperAdminAddOrganization)
	//superadmin.POST("/addorganizationadmin", controllers.SuperAdminAddOrganizationAdmin)
	superadmin.PUT("/organization/:id/deactivate", controllers.SoftDeleteOrganization)
	superadmin.PUT("/organization/:id/reactivate", controllers.ReactivateOrganization)
	superadmin.PATCH("/organization/:id", controllers.SoftDeleteOrganization)
	superadmin.GET("/organization/:id", controllers.GetOrganizationByID)
	superadmin.GET("/organizations", controllers.GetAllOrganizations)
	superadmin.PUT("/organization/edit", controllers.EditOrganization)
	superadmin.DELETE("/organization/:id", controllers.DeleteOrganization)
	superadmin.PUT("/organization/:id/activate", controllers.ActivateOrganization)
	superadmin.GET("/organizations/active", controllers.GetActiveOrganizations)
	superadmin.GET("/organizations/inactive", controllers.GetInactiveOrganizations)

	// Admin routes
	adminGroup := e.Group("/admin")
	adminGroup.Use(middlewares.AuthMiddleware(models.AdminRoleName))
	//superadmin.POST("/signup", controllers.AdminSignup)
	adminGroup.POST("/adduser", controllers.AdminAddUser)
	// adminGroup.POST("/createstock", controllers.CreateStock)
	// adminGroup.DELETE("/deletestock/:id", controllers.DeleteStock)
	// adminGroup.PUT("/editstock", controllers.EditStock)
	// adminGroup.GET("/viewallstock", controllers.ViewAllStock)
	// adminGroup.GET("/viewstock/:id", controllers.ViewStockByID)
	adminGroup.GET("/user/:id", controllers.GetUserByID)
	adminGroup.PUT("/user/:id", controllers.EditUser)
	adminGroup.PATCH("/user/:id", controllers.SoftDeleteUser)
	adminGroup.GET("/users", controllers.AdminViewAllUsers)
	adminGroup.GET("/users/active", controllers.GetActiveUsers)
	adminGroup.PUT("/user/:id/reactivate", controllers.ReactivateUser)
	adminGroup.PUT("/user/:id/deactivate", controllers.DeactivateUser)
	adminGroup.GET("/users/inactive", controllers.GetInactiveUsers)
	// adminGroup.POST("/addsupplier", controllers.AddSupplier)
	// adminGroup.PUT("/editsupplier/:id", controllers.EditSupplier)
	// adminGroup.DELETE("/deletesupplier", controllers.DeleteSupplier)
	// adminGroup.GET("/viewallsuppliers", controllers.GetAllSuppliers)
	// adminGroup.GET("/viewsupplier/:id", controllers.GetSupplier)
	adminGroup.DELETE("/user/:id", controllers.DeleteUser)

	//adminGroup.DELETE("/organization:id", controllers.AdminDeleteOrganization)

	//organization admin
	orgAdminGroup := e.Group("/orgadmin")
	orgAdminGroup.Use(middlewares.AuthMiddleware(models.OrganizationAdminRoleName))
	//orgAdminGroup.POST("/adduser", controllers.OrganizationAdminAddUser)
	//orgAdminGroup.PUT("/user/:id", controllers.OrganizationAdminEditUser)
	orgAdminGroup.GET("/users", controllers.OrganizationAdminGetUsers)
	//orgAdminGroup.GET("/user/:id", controllers.OrganizationAdminGetUserByID)
	orgAdminGroup.DELETE("/user/:id", controllers.OrganizationAdminSoftDeleteUser)
	//orgAdminGroup.DELETE("/user/:id", controllers.Delete)
	adminGroup.PUT("/organization/:id/deactivate", controllers.OrgAdminDeactivateOrganization)
	orgAdminGroup.PATCH("/users/:id/activate-deactivate", controllers.OrganizationAdminActivateDeactivateUser)

	organization := e.Group("/organization")
	e.Use(middlewares.AuthMiddleware("Admin", "Shopkeeper", "Auditor"))
	organization.POST("/addsupplier", controllers.AddSupplier)
	organization.PUT("/editsupplier/:id", controllers.EditSupplier)
	organization.DELETE("/deletesupplier", controllers.DeleteSupplier)
	organization.GET("/viewallsuppliers", controllers.GetAllSuppliers)
	organization.GET("/viewsupplier/:id", controllers.GetSupplier)
	organization.POST("/createstock", controllers.CreateStock)
	organization.DELETE("/deletestock/:id", controllers.DeleteStock)
	organization.PUT("/editstock", controllers.EditStock)
	organization.GET("/viewallstock", controllers.ViewAllStock)
	organization.GET("/viewstock/:id", controllers.ViewStockByID)

	organization.GET("/products/:product_id", controllers.GetProductByID)
	organization.GET("/products", controllers.GetProducts)
	organization.DELETE("/products/:product_id", controllers.DeleteProduct)
	organization.POST("/products", controllers.AddProduct) // Add a new product
	organization.PUT("/products/:product_id", controllers.UpdateProduct)

	organization.GET("", controllers.GetCategories)
	//organization.GET("/:category_id", controllers.GetCategoryByID)

	//organization.GET("/only", controllers.GetCategoriesOnly)
	organization.GET("/categories_only", controllers.GetCategories)
	organization.PUT("/categories_only/:id", controllers.UpdateCategory)
	organization.GET("/categories_only/:id", controllers.GetCategoryNameByID)
	organization.POST("/categories_only", controllers.CreateCategory)
	organization.DELETE("/categories_only/:id", controllers.DeleteCategoryByID)
	organization.GET("/categories_only/products/:id", controllers.GetProductsByCategoryID)
	//sale endpoints
	organization.GET("/sales/:sale_id", controllers.GetSalesBySaleID)
	organization.POST("/sell", controllers.SellProduct)
	//organization.GET("/cash/salesbyuser_id/:user_id", controllers.GetSalesByUser)
	organization.GET("/sales/reports_by_date_and_by_sales_ids/:date", controllers.GetSalesByDate)
	organization.GET("/sales/reports_by_sales_ids", controllers.GetAllSalesReports)
}
