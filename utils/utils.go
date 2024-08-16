package utils

import (
	"errors"
	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"log"
	"time"
)

var (
	JwtSecret = []byte("your-secret-key") // Replace with your actual secret key
	DB        *gorm.DB
)

var ErrInvalidToken = errors.New("invalid token")

// Claims represents the JWT claims
type Claims struct {
	UserID uint `json:"user_id"`
	RoleID uint `json:"role_id"`
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

// GenerateJWT generates a JWT token with userID and roleID
func GenerateJWT(userID, roleID uint) (string, error) {
	log.Printf("Generating JWT for userID: %d, roleID: %d", userID, roleID)
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: userID,
		RoleID: roleID,
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

	log.Printf("Verified JWT with userID: %d, roleID: %d", claims.UserID, claims.RoleID)
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
