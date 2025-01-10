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
	FullName      string         `json:"full_name"`
	Organization      string         `json:"organization"`
	RoleName       string         `json:"role_name"`
	OrganizationID uint           `json:"organization_id"`
	IsActive       bool           `json:"is_active"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	CreatedBy      uint           `json:"created_by"`
	Phonenumber    int64          `json:"phone_number" gorm:"column:phone_number"`
	//Phonenumber    int64          `json:"phone_number"`
	//ResetToken       string    
    //ResetTokenExpiry time.Time 
}

type Users struct {
    ID                  uint       `gorm:"primaryKey" json:"id"`
    UserID              uint       `gorm:"not null" json:"user_id"`
	Email          string         `json:"email"`
	Password       string         `json:"password"`
    ResetToken          string     
    ResetTokenExpiry    time.Time 
}

type Suppliers struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	Name        string `json:"name"`
	Phonenumber int64  `json:"phonenumber"`
	OrganizationID uint `json:"organization_id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	CreatedBy      uint           `json:"created_by"`
	Stock          []Stock        `json:"stock" gorm:"foreignKey:SupplierID"`
}
func (Suppliers) TableName() string {
    return "suppliers"
}












