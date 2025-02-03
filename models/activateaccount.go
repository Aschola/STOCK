package models

import "time"

type ActivationToken struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    UserID    uint      `json:"user_id"`
    Token     string    `json:"token" gorm:"unique"`
    ExpiresAt time.Time `json:"expires_at"`
    Used      bool      `json:"used"`
    CreatedAt time.Time `json:"created_at"`
}

// ResetToken struct represents the password reset token in the database
type ResetToken struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
	User      User      `json:"user" gorm:"foreignKey:UserID"`
}
