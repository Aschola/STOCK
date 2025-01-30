package models

import (
	"time"
)
type ChangePasswordRequest struct {
	
	ID        uint      `json:"id" gorm:"primaryKey"`
    UserID    uint      `json:"user_id" gorm:"index"`
    //Password  string    `json:"password"` // Hashed password
    CreatedAt time.Time `json:"created_at"`
	OldPassword     string `json:"old_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`



}
func (ChangePasswordRequest) TableName() string{
	return "password"
}