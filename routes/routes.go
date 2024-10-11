package routes

import (
	"stock/controllers"
	"stock/middlewares"
	"stock/models"

	"github.com/labstack/echo/v4"
)

// RegisterRoutes initializes all the routes for the Echo server
func RegisterRoutes(e *echo.Echo) {

	e.GET("/categories", controllers.GetCategories)
	e.GET("/categories/:category_id", controllers.GetCategoryByID)
	e.POST("/categories", controllers.CreateCategories)
	e.PUT("/categories/:category_id", controllers.UpdateCategory)
	e.DELETE("/categories/:id", controllers.DeleteCategoryByID)

	// Define CRUD endpoints for products without middleware
	productGroup := e.Group("/products")
	productGroup.GET("", controllers.GetProducts)
	productGroup.GET("/:product_id", controllers.GetProductByID)
	productGroup.POST("", controllers.AddProduct)
	productGroup.PUT("/:product_id", controllers.UpdateProduct)
	productGroup.DELETE("/:product_id", controllers.MakeProductsInactive)

	e.PUT("/products/inactive-products/:product_id", controllers.RestoreProductFromInactiveTablee)
	e.DELETE("/products/inactive-products/:product_id", controllers.DeleteProductFromInactiveTablee)
	e.GET("/products/inactive-products", controllers.GetAllInactiveProductss)

	// Define CRUD endpoints for sales
	e.GET("/sales", controllers.GetSales)
	e.GET("/sales/:sale_id", controllers.GetSaleByID)
	e.POST("/sales", controllers.AddSale)
	e.DELETE("/sales/:sale_id", controllers.DeleteSale)

	// Endpoint for selling products
	//e.POST("/products/:product_id/sell/:quantity_sold", controllers.SellProduct)
	e.GET("/sales", controllers.GetSalesByCategory)
	e.GET("/sales/user", controllers.GetSalesByUserID)
	e.GET("/sales/date", controllers.GetSalesByDate)
	e.POST("/sell-product", controllers.SellProduct)

	// Cash sales routes
	e.POST("/cash/sell", controllers.SellProductByCash)   // Sell a product by cash
	e.GET("/cash/sales", controllers.GetCashSales)        // Get all cash sales
	e.GET("/cash/sales/:id", controllers.GetCashSaleByID) // Get a cash sale by ID
	e.POST("/cash/sales", controllers.AddSaleByCash)      // Add a new cash sale
	e.DELETE("/cash/sales/:id", controllers.DeleteSaleByCash)
}

func SetupRoutes(e *echo.Echo) {
	// Public routes
	e.POST("/superadmin/login", controllers.SuperAdminLogin)
	e.POST("/superadmin/logout", controllers.SuperAdminLogout)
	e.POST("admin/login", controllers.AdminLogin)
	e.POST("admin/logout", controllers.AdminLogout)
	e.POST("/login", controllers.Login)
	e.POST("/logout", controllers.Logout)
	e.POST("auditor/login", controllers.AuditorLogin)
	e.POST("auditor/logout", controllers.AuditorLogout)
	e.POST("/organizationadmin/login", controllers.OrganizationAdminLogin)
	e.POST("/organizationadmin/logout", controllers.OrganizationAdminLogout)
	e.POST("/organization/logout", controllers.OrganizationLogout)
	e.POST("/organization/login", controllers.OrganizationLogin)

	// Super Admin routes
	superadmin := e.Group("/superadmin")
	superadmin.POST("/signup", controllers.SuperAdminSignup)
	superadmin.Use(middlewares.AuthMiddleware(models.SuperAdminRoleName)) 
	superadmin.POST("/addadmin", controllers.AddAdmin)
	superadmin.POST("/addorganization", controllers.SuperAdminAddOrganization)
	superadmin.POST("/addorganizationadmin", controllers.SuperAdminAddOrganizationAdmin)
	superadmin.PUT("/organization/:id/deactivate", controllers.SoftDeleteOrganization)
	superadmin.PUT("/organization/:id/reactivate", controllers.ReactivateOrganization)
	superadmin.PATCH("/organization/:id/deactivate", controllers.SoftDeleteOrganization)
    superadmin.PUT("/organization/:id/reactivate", controllers.ReactivateOrganization)


	// Admin routes
	adminGroup := e.Group("/admin")
	adminGroup.Use(middlewares.AuthMiddleware(models.AdminRoleName)) 
	adminGroup.POST("/adduser", controllers.AdminAddUser)
	adminGroup.GET("/user/:id", controllers.GetUserByID)
	adminGroup.PUT("/user/:id", controllers.EditUser)
	adminGroup.PATCH("/user/:id", controllers.SoftDeleteUser)
	adminGroup.GET("/user", controllers.AdminViewAllUsers)
	adminGroup.GET("/organization/:id", controllers.GetOrganizationByID)
	adminGroup.GET("/organizations", controllers.GetAllOrganizations)
	adminGroup.GET("/users/active", controllers.GetActiveUsers)
	//adminGroup.PUT("/user/activate", controllers.ActivateUser)
	adminGroup.PUT("/user/deactivate", controllers.DeactivateUser)
	adminGroup.GET("/users/inactive", controllers.GetInactiveUsers)
	adminGroup.GET("/organizations/active", controllers.GetActiveOrganizations)
	adminGroup.GET("/organizations/inactive", controllers.GetInactiveOrganizations)
	adminGroup.PUT("/organization/:id/activate", controllers.ActivateOrganization)
	//adminGroup.DELETE("/organization:id", controllers.AdminDeleteOrganization)

	//organization admin
	orgAdminGroup := e.Group("/orgadmin")
	orgAdminGroup.Use(middlewares.AuthMiddleware(models.OrganizationAdminRoleName))
	orgAdminGroup.POST("/adduser", controllers.OrganizationAdminAddUser)
	orgAdminGroup.PUT("/user/:id", controllers.OrganizationAdminEditUser)
	orgAdminGroup.GET("/users", controllers.OrganizationAdminGetUsers)
	orgAdminGroup.GET("/user/:id", controllers.OrganizationAdminGetUserByID)
	orgAdminGroup.DELETE("/user/:id", controllers.OrganizationAdminSoftDeleteUser)
	adminGroup.PUT("/organization/:id/deactivate", controllers.OrgAdminDeactivateOrganization)
	orgAdminGroup.PATCH("/users/:id/activate-deactivate", controllers.OrganizationAdminActivateDeactivateUser)
}
