package main

import (
	"github.com/labstack/echo/v4"
	"log"
	"os"
	"stock/db"
	"stock/routes"
)

func main() {
	// Initialize the database
	db.Init() // Changed from InitDB to Init

	// Create a new Echo instance
	e := echo.New()
	// Set up routes
	routes.RegisterRoutes(e) // Only pass the Echo instance
	routes.SetupRoutes(e)

	// Start the server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000" // Default port if not specified
	}
	log.Fatal(e.Start(":" + port))
}
