package routes

import (
    "stock/controllers"
    "stock/middlewares"
    "github.com/labstack/echo/v4"
    "stock/models"
)

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

    // Super Admin routes
    superadmin := e.Group("/superadmin")
    superadmin.POST("/signup", controllers.SuperAdminSignup)
    superadmin.Use(middlewares.AuthMiddleware(models.SuperAdminRoleID)) // Ensure SuperAdmin is authorized
    superadmin.POST("/addadmin", controllers.AddAdmin)
    superadmin.POST("/addorganization", controllers.SuperAdminAddOrganization)
    superadmin.POST("/addorganizationadmin", controllers.SuperAdminAddOrganizationAdmin)

    // Admin routes
    adminGroup := e.Group("/admin")
    adminGroup.Use(middlewares.AuthMiddleware(models.AdminRoleID)) // Ensure Admin is authorized
    adminGroup.POST("/adduser", controllers.AdminAddUser)
    adminGroup.GET("/user/:id", controllers.GetUserByID)
    adminGroup.PUT("/user/:id", controllers.EditUser)
    adminGroup.DELETE("/user/:id", controllers.SoftDeleteUser)
    adminGroup.GET("/user", controllers.AdminViewAllUsers)
    adminGroup.GET("/organization/:id", controllers.GetOrganizationByID) 
    adminGroup.GET("/organizations", controllers.GetAllOrganizations)
    adminGroup.GET("/users/active", controllers.GetActiveUsers)   
    adminGroup.PUT("/user/activate", controllers.ActivateUser) 
    adminGroup.PUT("/user/deactivate", controllers.DeactivateUser)   
    adminGroup.GET("/users/inactive", controllers.GetInactiveUsers)   
    adminGroup.GET("/organizations/active", controllers.GetActiveOrganizations)       
    adminGroup.GET("/organizations/inactive", controllers.GetInactiveOrganizations) 
    adminGroup.PUT("/organization/:id/activate", controllers.ActivateOrganization) 
    adminGroup.PUT("/organization/:id/deactivate", controllers.DeactivateOrganization) 
    //adminGroup.DELETE("/organization:id", controllers.AdminDeleteOrganization)


    orgAdminGroup := e.Group("/orgadmin")
    orgAdminGroup.Use(middlewares.AuthMiddleware(models.OrganizationAdminRoleID))
    orgAdminGroup.POST("/adduser", controllers.OrganizationAdminAddUser) 
    orgAdminGroup.PUT("/user/:id", controllers.OrganizationAdminEditUser)
    orgAdminGroup.GET("/users", controllers.OrganizationAdminGetUsers)
    orgAdminGroup.GET("/user/:id", controllers.OrganizationAdminGetUserByID)
    orgAdminGroup.DELETE("/user/:id", controllers.OrganizationAdminSoftDeleteUser)
    orgAdminGroup.PATCH("/users/:id/activate-deactivate", controllers.OrganizationAdminActivateDeactivateUser)
}

