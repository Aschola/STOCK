package main

import (
    "stock/db"
    "stock/routes"
    "github.com/labstack/echo/v4"
)

func main() {
    // Initialize the database
    db.Init()

    // Create a new Echo instance
    e := echo.New()

    // Set up routes
    routes.SetupRoutes(e) // Only pass the Echo instance

    // Start the server
    e.Logger.Fatal(e.Start(":8080"))
}
