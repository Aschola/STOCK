package models

import "time"

type CategoriesOnly struct {
	CategoryID      int    `json:"category_id"`
	CategoryName    string `json:"category_name"`
	OrganizationsID uint   `json:"organizations_id"`
}

func (CategoriesOnly) TableName() string {
	return "categories" // Updated table name
}

type Product struct {
	ProductID          int        `gorm:"primaryKey;autoIncrement" json:"product_id"`
	CategoryName       string     `json:"category_name"`                     // Name of the category
	ProductName        string     `json:"product_name"`                      // Name of the product
	ProductDescription string     `json:"product_description"`               // Description of the product
	ReorderLevel       int        `json:"reorder_level"`                     // Reorder level for inventory
	CreatedAt          time.Time  `gorm:"autoCreateTime" json:"created_at"`  // Timestamp when created
	UpdatedAt          time.Time  `gorm:"autoUpdateTime" json:"updated_at"`  // Timestamp when last updated
	DeletedAt          *time.Time `gorm:"index" json:"deleted_at,omitempty"` // Soft delete timestamp, nullable
	OrganizationsID    int64      `json:"organizations_id"`
	//Organization   string         `json:"organization"`
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
	SaleID            int64   `gorm:"primaryKey;autoIncrement" json:"sale_id"`
	Name              string  `gorm:"type:varchar(255)" json:"product_name"` // Removed the extra space
	UnitBuyingPrice   float64 `gorm:"type:decimal(10,2)" json:"unit_buying_price"`
	TotalBuyingPrice  float64 `gorm:"type:decimal(10,2)" json:"total_buying_price"`
	UnitSellingPrice  float64 `gorm:"type:decimal(10,2)" json:"unit_selling_price"`
	TotalSellingPrice float64 `gorm:"type:float" json:"total_selling_price"`
	Profit            float64 `gorm:"type:float" json:"profit"`
	Quantity          int     `gorm:"type:int" json:"quantity"`
	CashReceived      float64 `gorm:"type:decimal(10,2)" json:"cash_receive"`
	Balance           float64 `gorm:"type:decimal(10,2)" json:"balance"`
	UserID            int64   `json:"user_id"`

	Date            time.Time `gorm:"type:timestamp" json:"date"`
	CategoryName    string    `gorm:"type:varchar(255)" json:"category_name"`
	OrganizationsID uint      `json:"organization_id"` // Add this line
}

// TableName overrides the default table name (sales -> sales_by_cash)
func (Sale) TableName() string {
	return "sales_by_cash"
}

// SalePayload represents the data structure for sale request
type SalePayload struct {
	UserID       int        `json:"user_id"`
	CashReceived float64    `json:"cash_received"`
	Items        []SaleItem `json:"items"`
}

// SaleItem represents an individual item in the sale
type SaleItem struct {
	ProductID    int `json:"product_id"`
	QuantitySold int `json:"quantity_sold"`
}

// Define the CompanySetting struct to match the 'company_settings' table
type CompanySetting struct {
	Name           string `gorm:"not null"`
	Address        string `gorm:"not null"`
	Telephone      string `gorm:"not null"`
	OrganizationID uint   `gorm:"index"` // Foreign key to organizations
}
