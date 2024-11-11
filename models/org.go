package models

import (
	"time"
)

type Organization struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Name        string     `gorm:"not null" json:"name"`
	Email          string         `json:"email"`
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time  `gorm:"autoCreateTime" json:"createdAt"`
	UpdatedAt   time.Time  `gorm:"autoUpdateTime" json:"updatedAt"`
	DeletedAt   *time.Time `gorm:"index" json:"deletedAt,omitempty"`
}

