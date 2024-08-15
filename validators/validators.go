package validators

import (
    "regexp"

    "github.com/go-playground/validator/v10"
)

// LoginInput defines the structure for login input validation
type LoginInput struct {
    Username string `json:"username" validate:"required"`
    Password string `json:"password" validate:"required,password"`
}

// SignupInput defines the structure for signup input validation
type SignupInput struct {
    Username string `json:"username" validate:"required,usernameNoSpace"`
    Password string `json:"password" validate:"required,password"`
}

// Custom validation function to check if the username contains no spaces
func usernameNoSpaceValidation(fl validator.FieldLevel) bool {
    username := fl.Field().String()
    if len(username) == 0 {
        return false
    }
    return !regexp.MustCompile(`\s`).MatchString(username)
}

// Custom validation function to check if the password contains both numbers and letters
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

// ValidateSignupInput validates the signup input based on custom rules
func ValidateSignupInput(input SignupInput) error {
    validate := validator.New()

    // Register custom validation functions
    validate.RegisterValidation("usernameNoSpace", usernameNoSpaceValidation)
    validate.RegisterValidation("password", passwordValidation)

    // Validate the input
    return validate.Struct(input)
}
