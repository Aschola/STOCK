package middlewares

import (
	"github.com/labstack/echo/v4"
	"log"
	"net/http"
	"stock/utils"
	"strings"
)

// AdminMiddleware checks if the user has the 'admin' role.
func AdminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		role := c.Get("roleName").(string) 
		if role != "Admin" { 
			return c.JSON(http.StatusForbidden, map[string]string{"message": "Access denied"})
		}
		return next(c)
	}
}

// AuthMiddleware validates the JWT token and checks if the user's role is allowed.
func AuthMiddleware(allowedRoles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Path() == "/login" || c.Path() == "/logout" {
				return next(c)
			}

			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Authorization header is required"})
			}

			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid token format"})
			}

			tokenString := tokenParts[1]
			token, err := utils.ParseToken(tokenString)
			if err != nil {
				log.Printf("Token parsing error: %v", err)
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid token"})
			}

			claims, ok := token.Claims.(*utils.Claims)
			if !ok || !token.Valid {
				log.Printf("Invalid token claims: %v", token)
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid token claims"})
			}

			userID := claims.UserID
			roleName := claims.RoleName 
			organizationID := claims.OrganizationID
			log.Printf("Token parsed successfully. UserID: %d, organization: %d, RoleName: %s", userID, roleName, organizationID)

			// Set context values
			c.Set("userID", userID)
			c.Set("roleName", roleName)
			c.Set("organizationID", organizationID)

			if !contains(allowedRoles, roleName) {
				log.Printf("Access denied for RoleName %s. Allowed roles: %v", roleName, allowedRoles)
				return c.JSON(http.StatusForbidden, echo.Map{"error": "Access forbidden"})
			}

			return next(c)
		}
	}
}

// Utility function to check if a slice contains a value.
func contains(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// Role-specific middlewares using roleName

func SuperAdminOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		roleName := c.Get("roleName").(string)
		if roleName != "Superadmin" {
			log.Printf("Access denied for RoleName %s", roleName)
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Access forbidden"})
		}
		return next(c)
	}
}

func AdminOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		roleName := c.Get("roleName").(string)
		if roleName != "Admin" {
			log.Printf("Access denied for RoleName %s", roleName)
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Access forbidden"})
		}
		return next(c)
	}
}

func ShopAttendantOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		roleName := c.Get("roleName").(string)
		if roleName != "Shopkeeper" {
			log.Printf("Access denied for RoleName %s", roleName)
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Access forbidden"})
		}
		return next(c)
	}
}

func OrganizationAdminOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		roleName := c.Get("roleName").(string)
		if roleName != "organization_admin" {
			log.Printf("Access denied for RoleName %s", roleName)
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Access forbidden"})
		}
		return next(c)
	}
}

func AuditorOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		roleName := c.Get("roleName").(string)
		if roleName != "Auditor" {
			log.Printf("Access denied for RoleName %s", roleName)
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Access forbidden"})
		}
		return next(c)
	}
}
