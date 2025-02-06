// package main

// import (
// 	"log"
// 	"os"
// 	"stock/controllers"
// 	"stock/db"
// 	"stock/routes"

// 	//"time"

// 	"github.com/labstack/echo/v4"
// 	"github.com/labstack/echo/v4/middleware"
// )

// func main() {
// 	// Initialize the database
// 	db.Init()

// 	// Create an Echo instance
// 	e := echo.New()

// 	// Log all registered routes
// 	for _, route := range e.Routes() {
// 		log.Printf("Registered route: %s %s", route.Method, route.Path)
// 	}

// 	// Register routes
// 	e.POST("/send-sms", controllers.SendSmsHandler)
// 	e.POST("/mpesa/callback", controllers.HandleMpesaCallback)

// 	// CORS configuration
// 	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
// 		AllowOrigins: []string{"*"},
// 		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},
// 		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
// 	}))
	
// 	// Middleware for body size limit
// 	e.Use(middleware.BodyLimit("2M"))

// 	// Add middleware to parse and log request body
// 	e.Use(middleware.BodyDump(func(c echo.Context, reqBody, resBody []byte) {
// 		log.Printf("Request Body: %s", string(reqBody))
// 	}))

// 	// Register additional routes (if needed)
// 	routes.RegisterRoutes(e)
// 	routes.SetupRoutes(e)

// 	// Set the port from environment variable or use default (8000)
// 	port := os.Getenv("PORT")
// 	if port == "" {
// 		port = "8000"
// 	}

// 	log.Fatal(e.Start(":" + port))
// }


// package main

// import (
// 	"fmt"
// 	"log"
// 	"os"
// 	"stock/db"
// 	"stock/routes"

// 	"github.com/golang-migrate/migrate/v4"
// 	_ "github.com/golang-migrate/migrate/v4/database/mysql"
// 	_ "github.com/golang-migrate/migrate/v4/source/file"
// 	"github.com/labstack/echo/v4"
// 	"github.com/labstack/echo/v4/middleware"
// )

// // runMigrations applies all pending migrations
// func runMigrations() error {
//     // Load database configuration
//     dbHost := os.Getenv("DB_HOST")
//     dbPort := os.Getenv("DB_PORT")
//     dbName := os.Getenv("DB_NAME")
//     dbUser := os.Getenv("DB_USER")
//     dbPassword := os.Getenv("DB_PASSWORD")

//     if dbHost == "" || dbPort == "" || dbName == "" || dbUser == "" || dbPassword == "" {
//         return fmt.Errorf("database environment variables are not properly set")
//     }

//     dsn := fmt.Sprintf("mysql://%s:%s@tcp(%s:%s)/%s?multiStatements=true", dbUser, dbPassword, dbHost, dbPort, dbName)

//     // Debug: Print the migrations path
//     migrationsPath := "file:///home/schola/Downloads/STOCK/migrations"
//     log.Printf("Using migrations path: %s", migrationsPath)

//     m, err := migrate.New(migrationsPath, dsn)
//     if err != nil {
//         return fmt.Errorf("failed to initialize migration: %v", err)
//     }

//     err = m.Up()
//     if err != nil && err != migrate.ErrNoChange {
//         return fmt.Errorf("migration failed: %v", err)
//     }

//     log.Println("Migrations applied successfully!")
//     return nil
// }

// func main() {
// 	// Initialize the database
// 	db.Init()

// 	// Run migrations
// 	err := runMigrations()
// 	if err != nil {
// 		log.Fatalf("Migration error: %v", err)
// 	}

// 	e := echo.New()

// 	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
// 		AllowOrigins: []string{"*"},
// 		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},
// 		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
// 	}))

// 	// Register routes
// 	routes.RegisterRoutes(e)
// 	routes.SetupRoutes(e)

// 	port := os.Getenv("PORT")
// 	if port == "" {
// 		port = "8080"
// 	}

// 	log.Fatal(e.Start(":" + port))
// }

package main

import (
	"log"
	"os"
	"stock/controllers"
	"stock/db"
	"stock/routes"

	"github.com/joho/godotenv" // Import .env loader
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("[ERROR] Error loading .env file:", err)
	} else {
		log.Println("[INFO] .env file loaded successfully")
	}

	// Initialize the database
	db.Init()

	// Initialize email configuration
	err = controllers.InitEmailConfig()
	if err != nil {
		log.Fatalf("[ERROR] Failed to initialize email config: %v", err)
	}

	//go controllers.CheckAndInsertMissingOrganizations(db.GetDB())
	//go controllers.StartReorderLevelNotification(db.GetDB())

	e := echo.New()

	e.POST("/send-sms", controllers.SendSmsHandler)
	e.POST("/mpesa/callback", controllers.HandleMpesaCallback)
	e.POST("/reset-password", controllers.ResetPassword)


	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	// Register routes
	routes.RegisterRoutes(e)
	routes.SetupRoutes(e)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Println("[INFO] Server starting on port", port)
	e.Start(":" + port)
}
