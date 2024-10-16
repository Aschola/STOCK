package validators

import (
	"github.com/go-playground/validator/v10"
	"regexp"
)

type LoginInput struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,password"`
}

type SignupInput struct {
	Username string `json:"username" validate:"required,usernameNoSpace"`
	Password string `json:"password" validate:"required,password"`
}

type AdminEditInput struct {
	Username  string `json:"username" validate:"required,usernameNoSpace"`
	FirstName string `json:"firstname" validate:"required,alpha"`
	LastName  string `json:"lastname" validate:"required,alpha"`   
	Email     string `json:"email" validate:"required,email"`      
	Password  string `json:"password" validate:"omitempty,password"` 
}

// Custom validation function to check if the username contains no spaces
func usernameNoSpaceValidation(fl validator.FieldLevel) bool {
	username := fl.Field().String()
	if len(username) == 0 {
		return false
	}
	return !regexp.MustCompile(`\s`).MatchString(username)
}

func passwordValidation(fl validator.FieldLevel) bool {
	password := fl.Field().String()
	if len(password) < 6 {
		return false
	}
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	return hasLetter && hasNumber
}

// ValidateLoginInput validates the login input based on custom rules
func ValidateLoginInput(input LoginInput) error {
	validate := validator.New()

	// Register custom validation functions
	validate.RegisterValidation("password", passwordValidation)

	// Validate the input
	return validate.Struct(input)
}

func ValidateSignupInput(input SignupInput) error {
	validate := validator.New()

	// Register custom validation functions
	validate.RegisterValidation("usernameNoSpace", usernameNoSpaceValidation)
	validate.RegisterValidation("password", passwordValidation)

	return validate.Struct(input)
}

func ValidateAdminEditInput(input AdminEditInput) error {
	validate := validator.New()

	// Register custom validation functions
	validate.RegisterValidation("usernameNoSpace", usernameNoSpaceValidation)
	validate.RegisterValidation("password", passwordValidation)

	return validate.Struct(input)
}