package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"os"
	"stock/db"
	"stock/routes"
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

	port := os.Getenv("PORT")
	if port == "" {
		port = "80"
	}

	log.Fatal(e.Start(":" + port))
}
