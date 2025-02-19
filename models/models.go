package models

import (
	//"strings"
	"time"
)

type CategoriesOnly struct {
	CategoryID      int    `json:"category_id"`
	CategoryName    string `json:"category_name"`
	OrganizationsID uint   `json:"organizations_id"`
}

func (CategoriesOnly) TableName() string {
	return "categories"
}

type Product struct {
	ProductID          int        `gorm:"primaryKey;autoIncrement" json:"product_id"`
	CategoryName       string     `json:"category_name"`
	ProductName        string     `json:"product_name"`
	ProductDescription string     `json:"product_description"`
	ReorderLevel       int        `json:"reorder_level"`
	CreatedAt          time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt          *time.Time `gorm:"index" json:"deleted_at,omitempty"`
	OrganizationsID    int64      `json:"organizations_id"`
}

type SaleByCategory struct {
	SaleID       int     `json:"sale_id"`
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	Quantity     int     `json:"quantity"`
	UserID       string  `json:"user_id"`
	Date         string  `json:"date"`
	CategoryName string  `json:"category_name"`
}

// // Sale represents the structure of the sales_transactions table
// type Sale struct {
// 	SaleID            int64     `gorm:"primaryKey;autoIncrement" json:"sale_id"`
// 	Name              string    `gorm:"type:varchar(255)" json:"product_name"`
// 	ProductID         int       `gorm:"column:product_id" json:"product_id"`
// 	UnitBuyingPrice   float64   `gorm:"type:decimal(10,2)" json:"unit_buying_price"`
// 	TotalBuyingPrice  float64   `gorm:"type:decimal(10,2)" json:"total_buying_price"`
// 	UnitSellingPrice  float64   `gorm:"type:decimal(10,2)" json:"unit_selling_price"`
// 	TotalSellingPrice int64   `gorm:"type:int" json:"total_selling_price"`
// 	Profit            float64   `gorm:"type:float" json:"profit"`
// 	Quantity          int       `gorm:"type:int" json:"quantity"`
// 	CashReceived      float64   `gorm:"type:decimal(10,2)" json:"cash_received"`
// 	Balance           float64   `gorm:"type:decimal(10,2)" json:"balance"`
// 	PaymentMode       string    `gorm:"type:varchar(50)" json:"payment_mode"`
// 	UserID            int64     `json:"user_id"`
// 	Date              time.Time `gorm:"type:timestamp" json:"date"`
// 	CategoryName      string    `gorm:"type:varchar(255)" json:"category_name"`
// 	OrganizationsID   uint      `json:"organization_id"`
// 	TransactionID     string  `json:"transaction_id"`
// 	TransactionStatus string `json:"status"`
// 	Username string `json:"username"`

// }

type Sale struct {
	SaleID            int64     `gorm:"primaryKey;autoIncrement" json:"sale_id"`
	Name              string    `gorm:"type:varchar(255)" json:"product_name"`
	ProductID         int       `gorm:"column:product_id" json:"product_id"`
	UnitBuyingPrice   float64   `gorm:"type:decimal(10,2)" json:"unit_buying_price"`
	TotalBuyingPrice  float64   `gorm:"type:decimal(10,2)" json:"total_buying_price"`
	UnitSellingPrice  float64   `gorm:"type:decimal(10,2)" json:"unit_selling_price"`
	TotalSellingPrice float64   `gorm:"type:decimal(10,2)" json:"total_selling_price"` // Change to float64
	Profit            float64   `gorm:"type:float" json:"profit"`
	Quantity          int       `gorm:"type:int" json:"quantity"`
	CashReceived      float64   `gorm:"type:decimal(10,2)" json:"cash_received"`
	Balance           float64   `gorm:"type:decimal(10,2)" json:"balance"`
	PaymentMode       string    `gorm:"type:varchar(50)" json:"payment_mode"`
	UserID            int64     `json:"user_id"`
	Date              time.Time `gorm:"type:timestamp" json:"date"`
	CategoryName      string    `gorm:"type:varchar(255)" json:"category_name"`
	OrganizationsID   uint      `json:"organization_id"`
	TransactionID     string    `json:"transaction_id"`
	TransactionStatus string    `json:"status"`
	Username          string    `json:"username"`
}

// TableName overrides the default table name (sales -> sales_transactions)
func (Sale) TableName() string {
	return "sales_transactions"
}

// SalePayload represents the data structure for sale request
type SalePayload struct {
	UserID       int        `json:"user_id"`
	CashReceived float64    `json:"cash_received"`
	Items        []SaleItem `json:"items"`
	PaymentMode  string     `json:"payment_mode"`
	PhoneNumber  int64      `json:"phone_number"`
}

// SaleItem represents an individual item in the sale
type SaleItem struct {
	ProductID    int `json:"product_id"`
	QuantitySold int `json:"quantity_sold"`
}

// Define the CompanySetting struct to match the 'company_settings' table
type CompanySetting struct {
	ID             uint    `gorm:"primaryKey;autoIncrement"`
	Name           string  `gorm:"not null"`
	Address        string  `gorm:"not null"`
	Telephone      string  `gorm:"not null"`
	OrganizationID uint    `gorm:"index"`
	KraPin         *string `json:"kra_pin"`
	Email   	string  `json:"email"`
}
