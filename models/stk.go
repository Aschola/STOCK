package models

import (
    "time"
   // "github.com/jinzhu/gorm"
)

type MPesaSettings struct {
    ID                int64     `gorm:"primary_key"`
    OrganizationID    int64     `gorm:"index"`
    ConsumerKey       string
    ConsumerSecret    string
    BusinessShortCode string
    PassKey           string
    CallbackURL       string
    Environment       string 
    CreatedAt         time.Time
    UpdatedAt         time.Time
}

func (MPesaSettings) TableName() string {
	return "mpesasettings" // Explicitly set the table name to "stock"
}