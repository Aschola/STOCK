package models

import "time"

// Category struct with product_description added
type Category struct {
	CategoryID         int    `json:"category_id"`
	CategoryName       string `json:"category_name"`
	ProductName        string `json:"product_name"`
	ProductDescription string `json:"product_description"`
}

type Product struct {
	ProductID          int        `json:"product_id"`
	CategoryName       string     `json:"category_name"`
	ProductName        string     `json:"product_name"`
	ProductCode        string     `json:"product_code"`
	ProductDescription string     `json:"product_description"`
	DateCreated        time.Time  `json:"date_created"`
	Quantity           int        `json:"quantity"`
	ReorderLevel       int        `json:"reorder_level"`
	BuyingPrice        float64    `json:"buying_price"`
	SellingPrice       float64    `json:"selling_price"`
	DateDeleted        *time.Time `json:"date_deleted"`  // Change to *time.Time
	DateRestored       *time.Time `json:"date_restored"` // Change to *time.Time
}

type ProductAtPendingDeletion struct {
	ProductID          int       `json:"product_id"`
	CategoryName       string    `json:"category_name"`
	ProductName        string    `json:"product_name"`
	ProductCode        string    `json:"product_code"`
	ProductDescription string    `json:"product_description"`
	DateCreated        time.Time `json:"date_created"`
	Quantity           int       `json:"quantity"`
	ReorderLevel       int       `json:"reorder_level"`
	BuyingPrice        float64   `json:"buying_price"`
	SellingPrice       float64   `json:"selling_price"`
	DateDeleted        time.Time `json:"date_deleted"`
}

type RestoredProduct struct {
	ProductID          int        `json:"product_id"`
	CategoryName       string     `json:"category_name"`
	ProductName        string     `json:"product_name"`
	ProductCode        string     `json:"product_code"`
	ProductDescription string     `json:"product_description"`
	DateCreated        time.Time  `json:"date_created"`
	Quantity           int        `json:"quantity"`
	ReorderLevel       int        `json:"reorder_level"`
	BuyingPrice        float64    `json:"buying_price"`
	SellingPrice       float64    `json:"selling_price"`
	DateDeleted        *time.Time `json:"date_deleted"`  // Retain from pending_deletion_products
	DateRestored       *time.Time `json:"date_restored"` // Set current time
}
type Sale struct {
	SaleID            int       `json:"sale_id" gorm:"primaryKey"`
	Name              string    `json:"name"`
	Quantity          int       `json:"quantity"`
	UnitBuyingPrice   float64   `json:"unit_buying_price"`
	TotalBuyingPrice  float64   `json:"total_buying_price"`
	UnitSellingPrice  float64   `json:"unit_selling_price"`
	UserID            string    `json:"user_id"`
	Date              time.Time `json:"date"`
	CategoryName      string    `json:"category_name"`
	TotalSellingPrice float64   `json:"total_selling_price"`
	Profit            float64   `json:"profit"`
}

type SalebyCash struct {
	SaleID            int       `json:"sale_id" gorm:"primaryKey"`
	Name              string    `json:"name"`
	Quantity          int       `json:"quantity"`
	UnitBuyingPrice   float64   `json:"unit_buying_price"`
	TotalBuyingPrice  float64   `json:"total_buying_price"`
	UnitSellingPrice  float64   `json:"unit_selling_price"`
	UserID            string    `json:"user_id"`
	Date              time.Time `json:"date"`
	CategoryName      string    `json:"category_name"`
	TotalSellingPrice float64   `json:"total_selling_price"`
	Profit            float64   `json:"profit"`
	CashReceive       float64   `json:"cash_receive"`
	Balance           float64   `json:"balance"`
}

type PendingDeletionProduct struct {
	ID                 int       `json:"id" gorm:"primaryKey;autoIncrement;not null"`
	ProductID          int       `json:"product_id" gorm:"index"` // Foreign key to products
	CategoryName       string    `json:"category_name"`
	ProductName        string    `json:"product_name"`
	ProductCode        string    `json:"product_code"`
	ProductDescription string    `json:"product_description"`
	Date               time.Time `json:"date"`
	Quantity           int       `json:"quantity"`
	ReorderLevel       int       `json:"reorder_level"`
	Price              float64   `json:"price"`
	DateDeleted        time.Time `json:"date_deleted"`
}