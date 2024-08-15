package db

import (
    "fmt"
    "log"
    "os"
    "stock/models"
    "github.com/joho/godotenv"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
)

var db *gorm.DB

// Init initializes the database connection and migrates models
func Init() {
    // Load .env file
    if err := godotenv.Load(); err != nil {
        log.Fatalf("Error loading .env file: %v", err)
    }

    // Get database connection details from environment variables
    dbUser := os.Getenv("DB_USER")
    dbPassword := os.Getenv("DB_PASSWORD")
    dbHost := os.Getenv("DB_HOST")
    dbPort := os.Getenv("DB_PORT")
    dbName := os.Getenv("DB_NAME")

    // Create the DSN (Data Source Name) for connecting to the database
    dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbUser, dbPassword, dbHost, dbPort, dbName)

    // Open a connection to the database
    conn, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatalf("Failed to connect to database: %v", err)
    }

    db = conn

    // Auto migrate the models
    if err := db.AutoMigrate(&models.User{}, &models.Organization{}); err != nil {
        log.Fatalf("Failed to migrate database: %v", err)
    }

    // Seed initial data if needed
    // You can implement this function based on your application's requirements
    seedInitialData()
}

// GetDB returns the database connection instance
func GetDB() *gorm.DB {
    return db
}

// Seed initial data if needed
func seedInitialData() {
    // Implement this function to seed initial data
}
