package models

//import "time"

// Category struct with product_description added
type Stock struct {
	CategoryID         int    `json:"category_id"`
	CategoryName       string `json:"category_name"`
	ProductName        string `json:"product_name"`
	ProductDescription string `json:"product_description"`
}