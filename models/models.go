package models

import "time"

// Category struct with product_description added
type Category struct {
	CategoryID         int    `json:"category_id"`
	CategoryName       string `json:"category_name"`
	ProductName        string `json:"product_name"`
	ProductDescription string `json:"product_description"`
}

type Categories_Only struct {
	CategoryID   int    `json:"category_id"`
	CategoryName string `json:"category_name"`
}

func (Categories_Only) TableName() string {
	return "categories_onlies" // Ensure the table name matches
}

type Product struct {
	ProductID          int       `gorm:"primaryKey;autoIncrement" json:"product_id"`
	CategoryName       string    `json:"category_name"` // Updated field name and JSON tag
	ProductName        string    `json:"product_name"`  // Updated field name and JSON tag
	ProductDescription string    `json:"product_description"`
	ReorderLevel       int       `json:"reorder_level"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
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

// Sale represents the structure of the sales_by_cash table
type Sale struct {
	SaleID            int       `gorm:"primaryKey;autoIncrement" json:"sale_id"`
	Name              string    `gorm:"type:varchar(255)" json:"name"`
	UnitBuyingPrice   float64   `gorm:"type:decimal(10,2)" json:"unit_buying_price"`
	TotalBuyingPrice  float64   `gorm:"type:decimal(10,2)" json:"total_buying_price"`
	UnitSellingPrice  float64   `gorm:"type:decimal(10,2)" json:"unit_selling_price"`
	TotalSellingPrice float64   `gorm:"type:float" json:"total_selling_price"`
	Profit            float64   `gorm:"type:float" json:"profit"`
	Quantity          int       `gorm:"type:int" json:"quantity"`
	CashReceive       float64   `gorm:"type:decimal(10,2)" json:"cash_receive"`
	Balance           float64   `gorm:"type:decimal(10,2)" json:"balance"`
	UserID            string    `gorm:"type:varchar(255)" json:"user_id"`
	Date              time.Time `gorm:"type:timestamp" json:"date"`
	CategoryName      string    `gorm:"type:varchar(255)" json:"category_name"`
}

// TableName overrides the default table name (sales -> sales_by_cash)
func (Sale) TableName() string {
	return "sales_by_cash"
}

type SalePayload struct {
	ProductID    int     `json:"product_id"`
	QuantitySold int     `json:"quantity_sold"`
	UserID       string  `json:"user_id"`
	CashReceived float64 `json:"cash_received"`
}
