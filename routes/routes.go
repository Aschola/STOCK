package routes

import (
	"github.com/labstack/echo/v4"
	"stock/controllers"
	"stock/middlewares" // Import the middleware package
)

// RegisterRoutes initializes all the routes for the Echo server
func RegisterRoutes(e *echo.Echo) {
	// Define CRUD endpoints for categories with admin middleware
	categoryGroup := e.Group("/categories")
	categoryGroup.Use(middleware.AdminMiddleware) // Apply middleware
	categoryGroup.GET("", controllers.GetCategories)
	categoryGroup.GET("/:category_id", controllers.GetCategoryByID)
	categoryGroup.POST("", controllers.CreateCategories)
	categoryGroup.PUT("/:category_id", controllers.UpdateCategory)
	categoryGroup.DELETE("/:id", controllers.DeleteCategoryByID)

	// Define CRUD endpoints for products with admin middleware
	productGroup := e.Group("/products")
	productGroup.Use(middleware.AdminMiddleware) // Apply middleware
	productGroup.GET("", controllers.GetProducts)
	productGroup.GET("/:product_id", controllers.GetProductByID)
	productGroup.POST("", controllers.AddProduct)
	productGroup.PUT("/:product_id", controllers.UpdateProduct)
	productGroup.DELETE("/:product_id", controllers.DeleteProduct)
	productGroup.DELETE("/:product_id/pending-deletion", controllers.MoveProductToPendingDeletion)
	productGroup.PUT("/:product_id/recover", controllers.MoveProductFromPendingDeletion)

	// Define CRUD endpoints for sales
	e.GET("/sales", controllers.GetSales)
	e.GET("/sales/:sale_id", controllers.GetSaleByID)
	e.POST("/sales", controllers.AddSale)
	e.DELETE("/sales/:sale_id", controllers.DeleteSale)
	e.GET("/salebycategory/:category_name", controllers.FetchSalesByCategory)
	e.GET("/salebycategory/:date", controllers.FetchSalesByDate)
	e.GET("/salebycategory/:user_id", controllers.FetchSalesByUserID)
	// Endpoint for selling products
	e.POST("/products/:product_id/sell/:quantity_sold", controllers.SellProduct)
}
