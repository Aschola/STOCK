package main

import (
	"log"
	"os"
	"stock/controllers"
	"stock/db"
	"stock/routes"

	//"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Initialize the database
	db.Init()

	go controllers.StartReorderLevelNotification(db.GetDB())

	e := echo.New()

	e.POST("/send-sms", controllers.SendSmsHandler)

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{echo.GET, echo.POST, echo.PUT, echo.DELETE},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	// Register routes
	routes.RegisterRoutes(e)
	routes.SetupRoutes(e)

	// Start the test goroutine to send SMS every 30 seconds

	// If you want to call the test function, you can now call it directly from main
	//go runTestSendSmsHandler()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(e.Start(":" + port))
}

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
