package models

import (
	"time"
)

// Product represents the product model.
type Stock struct {
	ID                uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductID         uint64    `gorm:"not null" json:"product_id"`
	Quantity          int       `gorm:"not null" json:"quantity"`
	BuyingPrice       float64   `gorm:"not null" json:"buying_price"`
	SellingPrice      float64   `gorm:"not null" json:"selling_price"`
	ExpiryDate        *time.Time `gorm:"type:date" json:"expiry_date,omitempty"`
	ProductDescription string   `gorm:"type:text" json:"product_description,omitempty"`
	CreatedBy         uint64    `gorm:"not null" json:"created_by"`
	CreatedAt         time.Time `gorm:"autoCreateTime" json:"created_at"`
	DeletedAt    *time.Time `gorm:"type:timestamp;column:deleted_at" json:"deleted_at,omitempty"`

}

func (Stock) TableName() string {
	return "stock" // Explicitly set the table name to "stock"
}