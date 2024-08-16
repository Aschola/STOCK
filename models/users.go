package models

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	Username       string         `json:"username" gorm:"type:varchar(255);unique_index"`
	Email          string         `json:"email" gorm:"type:varchar(255);unique_index"`
	Password       string         `json:"password" gorm:"type:varchar(255)"`
	FirstName      string         `json:"first_name" gorm:"type:varchar(255)"`
	LastName       string         `json:"last_name" gorm:"type:varchar(255)"`
	RoleID         uint           `json:"role_id"`
	OrganizationID uint           `json:"organization_id,omitempty"` // Nullable
	IsActive       bool           `json:"is_active" gorm:"default:true"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at"`
	CreatedBy      uint           `json:"created_by"`
	UpdatedBy      uint           `json: updated_by`
}
