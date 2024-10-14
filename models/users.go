package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	Username       string         `json:"username" gorm:"unique"`
	Email          string         `json:"email"`
	Password       string         `json:"password"`
	FirstName      string         `json:"first_name"`
	LastName       string         `json:"last_name"`
	RoleName       string         `json:"role_name"`
	OrganizationID uint           `json:"organization_id"`
	IsActive       bool           `json:"is_active"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	CreatedBy      uint           `json:"created_by"`
	Phonenumber    int64          `json:"phonenumber"`
}
