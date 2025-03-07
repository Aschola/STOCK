package utils

import (
	"errors"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"log"
	"time"
	"fmt"
	"net/smtp"
	"crypto/rand"
    "encoding/hex"
)

var (
	JwtSecret = []byte("v2vrG4pXcSOSK7u2wjko/RFgrhlO7/rx3N2qLj0KLU0=") 
	DB        *gorm.DB
)

var ErrInvalidToken = errors.New("invalid token")

type Claims struct {
	UserID   uint   `json:"user_id"`
	RoleName string `json:"role_name"` 
	OrganizationID uint `json:"organization_id"`
	jwt.StandardClaims
}

// ParseToken parses and validates the JWT token
func ParseToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Printf("Unexpected signing method: %v", token.Method)
			return nil, ErrInvalidToken
		}
		return JwtSecret, nil
	})
	if err != nil {
		log.Printf("Error parsing token: %v", err)
		return nil, err
	}
	if !token.Valid {
		log.Printf("Token is not valid")
		return nil, ErrInvalidToken
	}
	return token, nil
}

// GenerateJWT generates a JWT token with userID and roleName
// func GenerateJWT(userID uint, roleName string) (string, error) { 
// 	log.Printf("Generating JWT for userID: %d, roleName: %s", userID, roleName, ) 
// 	expirationTime := time.Now().Add(24 * time.Hour)
// 	claims := &Claims{
// 		UserID:   userID,
// 		RoleName: roleName, 
// 		//OrganizationID: organizationID,
// 		StandardClaims: jwt.StandardClaims{
// 			ExpiresAt: expirationTime.Unix(),
// 		},
// 	}

// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
// 	signedToken, err := token.SignedString(JwtSecret)
// 	if err != nil {
// 		log.Printf("Error signing token: %v", err)
// 		return "", err
// 	}
// 	return signedToken, nil
// }
func GenerateJWT(userID uint, roleName string, organizationID uint) (string, error) {
	log.Printf("Generating JWT for userID: %d, roleName: %s, organizationID: %d", userID, roleName, organizationID)

	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &Claims{
		UserID:         userID,
		RoleName:       roleName,
		OrganizationID: organizationID, 
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(JwtSecret)
	if err != nil {
		log.Printf("Error signing token: %v", err)
		return "", err
	}
	return signedToken, nil
}


// VerifyJWT verifies the JWT token and returns the claims
func VerifyJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			log.Printf("Unexpected signing method: %v", token.Method)
			return nil, ErrInvalidToken
		}
		return JwtSecret, nil
	})

	if err != nil {
		log.Printf("Error verifying token: %v", err)
		return nil, err
	}

	if !token.Valid {
		log.Printf("Token is not valid")
		return nil, ErrInvalidToken
	}

	//log.Printf("Verified JWT with userID: %d, roleName: %s", claims.UserID, claims.RoleName) 
	log.Printf("Generating JWT for userID: %d, organizationID: %d, roleName: %s", claims.UserID, claims.RoleName, claims.OrganizationID ) 
	return claims, nil
}

// HashPassword hashes a password
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return "", err
	}
	return string(hashedPassword), nil
}

// CheckPasswordHash compares a password hash
func CheckPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		log.Printf("Password hash mismatch: %v", err)
	}
	return err
}

func GenerateResetToken() string {
    bytes := make([]byte, 16)
    rand.Read(bytes)
    return hex.EncodeToString(bytes)
}
func SendPasswordResetEmail(to string, resetLink string) {
    from := "smsafrica@infinitytechafrica.com"
    password := "Ngatia@01"

    smtpHost := "smtp.gmail.com"
    smtpPort := "587"

    message := []byte(fmt.Sprintf("To: %s\r\nSubject: Password Reset\r\n\r\nClick the following link to reset your password: %s", to, resetLink))

    auth := smtp.PlainAuth("", from, password, smtpHost)
    smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, message)
}

