package controllers

import (
	"fmt"
	"log"
	"time"

	"golang.org/x/crypto/bcrypt"
	"stock/db"
	"stock/models"
)
func SeedUser() {
	database := db.GetDB()

	users := []models.User{
		{
			Username:       "rooot",
			Email:          "support@infinitytechafrica.com",
			FullName:       "Super Admin",
			Organization:   "",
			RoleName:       "Superadmin",
			Password:       "Newvera@764", 
			//OrganizationID: 1, 
			IsActive:       true,
			CreatedBy:      1, 
			Phonenumber:    0720000000,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
		{
			Username:       "rooot",
			Email:          "support@infinitytechafrica.com",
			FullName:       "Super Admin",
			Organization:   "",
			RoleName:       "Superadmin",
			Password:       "Newvera@764",
			//OrganizationID: 1, 
			IsActive:       true,
			//CreatedBy:      1, 
			Phonenumber:    0720000000,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		},
	}

	for _, user := range users {
		hashedPassword, err := HashPassword(user.Password)
		if err != nil {
			log.Fatalf("Error hashing password for user %s: %v", user.Username, err)
		}

		user.Password = hashedPassword

		// Insert user into the database
		if err := database.Create(&user).Error; err != nil {
			log.Printf("Error inserting user %s: %v", user.Username, err)
		} else {
			fmt.Printf("User %s seeded successfully!\n", user.Username)
		}
	}
}

// HashPassword hashes a given password
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}