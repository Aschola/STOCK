// package db

// import (
// 	"fmt"
// 	"github.com/joho/godotenv"
// 	"gorm.io/driver/mysql"
// 	"gorm.io/gorm"
// 	"log"
// 	"os"
// )

// var dbInstance *gorm.DB

// func Init() {
// 	// Load environment variables from .env file if present
// 	err := godotenv.Load()
// 	if err != nil {
// 		log.Printf("Error loading .env file: %v", err)
// 	}

// 	dbHost := os.Getenv("DB_HOST")
// 	dbPort := os.Getenv("DB_PORT")
// 	dbName := os.Getenv("DB_NAME")
// 	dbUser := os.Getenv("DB_USER")
// 	dbPassword := os.Getenv("DB_PASSWORD")

// 	if dbHost == "" || dbPort == "" || dbName == "" || dbUser == "" || dbPassword == "" {
// 		log.Fatal("Database environment variables are not properly set")
// 	}

// 	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbUser, dbPassword, dbHost, dbPort, dbName)
// 	dbInstance, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
// 	if err != nil {
// 		log.Fatalf("Error connecting to the database: %v", err)
// 	}
// }

// // GetDB returns the GORM database instance
// func GetDB() *gorm.DB {
// 	if dbInstance == nil {
// 		log.Fatal("Database connection is not initialized")
// 	}
// 	return dbInstance
// }
package db

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"stock/models" 
)

var dbInstance *gorm.DB

func Init() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")

	if dbHost == "" || dbPort == "" || dbName == "" || dbUser == "" || dbPassword == "" {
		log.Fatal("Database environment variables are not properly set")
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", dbUser, dbPassword, dbHost, dbPort, dbName)
	dbInstance, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}

	err = dbInstance.AutoMigrate(
		&models.CompanySetting{},
	)
	if err != nil {
		log.Fatalf("Error during migration: %v", err)
	}

	log.Println("Database migrations completed successfully.")
}

// GetDB returns the GORM database instance
func GetDB() *gorm.DB {
	if dbInstance == nil {
		log.Fatal("Database connection is not initialized")
	}
	return dbInstance
}
