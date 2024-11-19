package models

import (
	"time"
)

// Product represents the product model.
type Stock struct {
	ID                uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductID         uint64    `gorm:"not null" json:"product_id"`
	ProductName       string    `gorm:"type:varchar(255);not null" json:"product_name"`
	Quantity          int       `gorm:"not null" json:"quantity"`
	BuyingPrice       float64   `gorm:"not null" json:"buying_price"`
	SellingPrice      float64   `gorm:"not null" json:"selling_price"`
	CategoryName      string    `gorm:"type:varchar(255);not null" json:"category_name"`
	ReferenceNumber   string    `gorm:"type:varchar(255)" json:"reference_number,omitempty"`
	ExpiryDate        *time.Time `gorm:"type:date" json:"expiry_date,omitempty"`
	ProductDescription string   `gorm:"type:text" json:"product_description,omitempty"`
	CreatedBy         uint64    `gorm:"not null" json:"created_by"`
	CreatedAt         time.Time `gorm:"autoCreateTime" json:"created_at"`
	ReorderLevel      int       `gorm:"not null" json:"reorder_level"`
}
