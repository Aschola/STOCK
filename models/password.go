package models

import (

)
type ChangePasswordRequest struct {
	OldPassword     string `json:"old_password" validate:"required"`
	NewPassword     string `json:"new_password" validate:"required,min=8"`
	ConfirmPassword string `json:"confirm_password" validate:"required"`
}
func (ChangePasswordRequest) TableName() string{
	return "password"
}