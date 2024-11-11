package main

import (
	"log"
	"os"
	//"stock/controllers"
	"stock/db"
	"stock/routes"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Initialize the database
	db.Init()

	e := echo.New()

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	// Register routes
	routes.RegisterRoutes(e)
	routes.SetupRoutes(e)
	//go controllers.GetProductsfornotification()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(e.Start(":" + port))
}
