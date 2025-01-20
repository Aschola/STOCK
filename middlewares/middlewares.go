// package middlewares

// import (
// 	"github.com/labstack/echo/v4"
// 	"log"
// 	"net/http"
// 	"stock/utils"
// 	"strings"
// )

// // AdminMiddleware checks if the user has the 'admin' role.
// func AdminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
// 	return func(c echo.Context) error {
// 		role := c.Get("roleName").(string) 
// 		if role != "Admin" { 
// 			return c.JSON(http.StatusForbidden, map[string]string{"message": "Access denied"})
// 		}
// 		return next(c)
// 	}
// }
// func OrganizationIDMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
//     return func(c echo.Context) error {
//         if c.Path() == "/login" || c.Path() == "/logout" {
//             return next(c)
//         }

//         authHeader := c.Request().Header.Get("Authorization")
//         if authHeader == "" {
//             return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Authorization header is required"})
//         }

//         tokenParts := strings.Split(authHeader, " ")
//         if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
//             return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid token format"})
//         }

//         tokenString := tokenParts[1]
//         token, err := utils.ParseToken(tokenString)
//         if err != nil {
//             log.Printf("Error parsing JWT: %v", err)
//             return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid token"})
//         }

//         // Extract the claims from the token
//         claims, ok := token.Claims.(*utils.Claims)
//         if !ok || !token.Valid {
//             log.Printf("Invalid token claims: %v", token)
//             return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid token claims"})
//         }

//         // Extract organizationID from claims
//         organizationID := claims.OrganizationID
//         log.Printf("Token parsed successfully. OrganizationID: %d", organizationID)

//         // Set organizationID in the context
//         c.Set("organizationID", organizationID)
// 		//c.Set("roleName", roleName)

//         return next(c)
//     }
// }

// // AuthMiddleware validates the JWT token and checks if the user's role is allowed.
// func AuthMiddleware(allowedRoles ...string) echo.MiddlewareFunc {
// 	return func(next echo.HandlerFunc) echo.HandlerFunc {
// 		return func(c echo.Context) error {
// 			if c.Path() == "/login" || c.Path() == "/logout" {
// 				return next(c)
// 			}

// 			authHeader := c.Request().Header.Get("Authorization")
// 			if authHeader == "" {
// 				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Authorization header is required"})
// 			}

// 			tokenParts := strings.Split(authHeader, " ")
// 			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
// 				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid token format"})
// 			}

// 			tokenString := tokenParts[1]
// 			token, err := utils.ParseToken(tokenString)
// 			if err != nil {
// 				log.Printf("Token parsing error: %v", err)
// 				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid token"})
// 			}

// 			claims, ok := token.Claims.(*utils.Claims)
// 			if !ok || !token.Valid {
// 				log.Printf("Invalid token claims: %v", token)
// 				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid token claims"})
// 			}

// 			userID := claims.UserID
// 			roleName := claims.RoleName 
// 			organizationID := claims.OrganizationID
// 			log.Printf("Token parsed successfully. UserID: %d, organization: %d, RoleName: %s", userID, roleName, organizationID)

// 			// Set context values
// 			c.Set("userID", userID)
// 			c.Set("roleName", roleName)
// 			c.Set("organizationID", organizationID)

// 			if !contains(allowedRoles, roleName) {
// 				log.Printf("Access denied for RoleName %s. Allowed roles: %v", roleName, allowedRoles)
// 				return c.JSON(http.StatusForbidden, echo.Map{"error": "Access forbidden"})
// 			}

// 			return next(c)
// 		}
// 	}
// }

// // Utility function to check if a slice contains a value.
// func contains(slice []string, value string) bool {
// 	for _, v := range slice {
// 		if v == value {
// 			return true
// 		}
// 	}
// 	return false
// }

// // Role-specific middlewares using roleName

// func SuperAdminOnly(next echo.HandlerFunc) echo.HandlerFunc {
// 	return func(c echo.Context) error {
// 		roleName := c.Get("roleName").(string)
// 		if roleName != "Superadmin" {
// 			log.Printf("Access denied for RoleName %s", roleName)
// 			return c.JSON(http.StatusForbidden, map[string]string{"error": "Access forbidden"})
// 		}
// 		return next(c)
// 	}
// }

// func AdminOnly(next echo.HandlerFunc) echo.HandlerFunc {
// 	return func(c echo.Context) error {
// 		roleName := c.Get("roleName").(string)
// 		if roleName != "Admin" {
// 			log.Printf("Access denied for RoleName %s", roleName)
// 			return c.JSON(http.StatusForbidden, map[string]string{"error": "Access forbidden"})
// 		}
// 		return next(c)
// 	}
// }

// func ShopAttendantOnly(next echo.HandlerFunc) echo.HandlerFunc {
// 	return func(c echo.Context) error {
// 		roleName := c.Get("roleName").(string)
// 		if roleName != "Shopkeeper" {
// 			log.Printf("Access denied for RoleName %s", roleName)
// 			return c.JSON(http.StatusForbidden, map[string]string{"error": "Access forbidden"})
// 		}
// 		return next(c)
// 	}
// }

// func OrganizationAdminOnly(next echo.HandlerFunc) echo.HandlerFunc {
// 	return func(c echo.Context) error {
// 		roleName := c.Get("roleName").(string)
// 		if roleName != "organization_admin" {
// 			log.Printf("Access denied for RoleName %s", roleName)
// 			return c.JSON(http.StatusForbidden, map[string]string{"error": "Access forbidden"})
// 		}
// 		return next(c)
// 	}
// }

// func AuditorOnly(next echo.HandlerFunc) echo.HandlerFunc {
// 	return func(c echo.Context) error {
// 		roleName := c.Get("roleName").(string)
// 		if roleName != "Auditor" {
// 			log.Printf("Access denied for RoleName %s", roleName)
// 			return c.JSON(http.StatusForbidden, map[string]string{"error": "Access forbidden"})
// 		}
// 		return next(c)
// 	}
// }

package middlewares

import (
	"log"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"stock/utils"
)

// AdminMiddleware checks if the user has the 'Admin' role or 'Superadmin' role.
func AdminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		role := c.Get("roleName").(string)
		if role != "Admin" && role != "Superadmin" { // Allow Superadmin
			log.Println("[AdminMiddleware] Access denied for non-admin or non-superadmin user")
			return c.JSON(http.StatusForbidden, map[string]string{"message": "Access denied"})
		}
		log.Println("[AdminMiddleware] Admin or Superadmin access granted")
		return next(c)
	}
}

// AuthMiddleware validates the JWT token and checks user roles for access.
func AuthMiddleware(allowedRoles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			log.Printf("[AuthMiddleware] Processing route: %s", c.Path())

			// Skip auth checks for any login or logout route
			if isLoginOrLogoutOrSignupRoute(c.Path()) {
				log.Println("[AuthMiddleware] Skipping authorization for login/logout")
				return next(c)
			}

			// Check Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				log.Println("[AuthMiddleware] Missing Authorization header")
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Authorization header is required"})
			}

			// Parse the token
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				log.Println("[AuthMiddleware] Invalid token format")
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid token format"})
			}

			tokenString := tokenParts[1]
			token, err := utils.ParseToken(tokenString)
			if err != nil {
				log.Printf("[AuthMiddleware] Token parsing error: %v", err)
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid token"})
			}

			claims, ok := token.Claims.(*utils.Claims)
			if !ok || !token.Valid {
				log.Printf("[AuthMiddleware] Invalid token claims: %v", token)
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid token claims"})
			}

			// Set user and organization data in context
			userID := claims.UserID
			roleName := claims.RoleName
			organizationID := claims.OrganizationID
			log.Printf("[AuthMiddleware] Token validated. UserID: %d, OrganizationID: %d, RoleName: %s", userID, organizationID, roleName)

			c.Set("userID", userID)
			c.Set("roleName", roleName)
			c.Set("organizationID", organizationID)

			// Check if the user's role is allowed or is Superadmin
			if !contains(allowedRoles, roleName) && roleName != "Superadmin" { // Allow Superadmin
				log.Printf("[AuthMiddleware] Access denied for RoleName: %s. Allowed roles: %v", roleName, allowedRoles)
				return c.JSON(http.StatusForbidden, echo.Map{"error": "Access forbidden"})
			}

			return next(c)
		}
	}
}

// Role-specific middlewares using roleName

// SuperAdminOnly allows only Superadmin to proceed.
func SuperAdminOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		roleName := c.Get("roleName").(string)
		if roleName != "Superadmin" {
			log.Printf("[SuperAdminOnly] Access denied for RoleName: %s", roleName)
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Access forbidden"})
		}
		log.Println("[SuperAdminOnly] Superadmin access granted")
		return next(c)
	}
}

// AdminOnly allows both Admin and Superadmin.
func AdminOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		roleName := c.Get("roleName").(string)
		if roleName != "Admin" && roleName != "Superadmin" { // Allow Superadmin
			log.Printf("[AdminOnly] Access denied for RoleName: %s", roleName)
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Access forbidden"})
		}
		log.Println("[AdminOnly] Admin or Superadmin access granted")
		return next(c)
	}
}

// ShopAttendantOnly allows Shopkeeper and Superadmin.
func ShopAttendantOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		roleName := c.Get("roleName").(string)
		if roleName != "Shopkeeper" && roleName != "Superadmin" { // Allow Superadmin
			log.Printf("[ShopAttendantOnly] Access denied for RoleName: %s", roleName)
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Access forbidden"})
		}
		log.Println("[ShopAttendantOnly] Shopkeeper or Superadmin access granted")
		return next(c)
	}
}

// OrganizationAdminOnly allows organization_admin and Superadmin.
func OrganizationAdminOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		roleName := c.Get("roleName").(string)
		if roleName != "organization_admin" && roleName != "Superadmin" { // Allow Superadmin
			log.Printf("[OrganizationAdminOnly] Access denied for RoleName: %s", roleName)
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Access forbidden"})
		}
		log.Println("[OrganizationAdminOnly] Organization Admin or Superadmin access granted")
		return next(c)
	}
}

// AuditorOnly allows Auditor and Superadmin.
func AuditorOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		roleName := c.Get("roleName").(string)
		if roleName != "Auditor" && roleName != "Superadmin" { // Allow Superadmin
			log.Printf("[AuditorOnly] Access denied for RoleName: %s", roleName)
			return c.JSON(http.StatusForbidden, map[string]string{"error": "Access forbidden"})
		}
		log.Println("[AuditorOnly] Auditor or Superadmin access granted")
		return next(c)
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

// Utility function to check if the route is a login or logout route.
func isLoginOrLogoutOrSignupRoute(path string) bool {
	// Skip authentication for routes that contain "/login" or "/logout"
	return strings.Contains(path, "/login") || strings.Contains(path, "/logout") || strings.Contains(path, "/signup")
}

// OrganizationIDMiddleware is unchanged as per the request.
func OrganizationIDMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		log.Printf("[OrganizationIDMiddleware] Processing route: %s", c.Path())

		if isLoginOrLogoutOrSignupRoute(c.Path()) {
			log.Println("[OrganizationIDMiddleware] Skipping authorization for login/logout")
			return next(c)
		}

		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			log.Println("[OrganizationIDMiddleware] Missing Authorization header")
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Authorization header is required"})
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			log.Println("[OrganizationIDMiddleware] Invalid token format")
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid token format"})
		}

		tokenString := tokenParts[1]
		token, err := utils.ParseToken(tokenString)
		if err != nil {
			log.Printf("[OrganizationIDMiddleware] Error parsing JWT: %v", err)
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid token"})
		}

		claims, ok := token.Claims.(*utils.Claims)
		if !ok || !token.Valid {
			log.Printf("[OrganizationIDMiddleware] Invalid token claims: %v", token)
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "Invalid token claims"})
		}

		organizationID := claims.OrganizationID
		log.Printf("[OrganizationIDMiddleware] Token validated. OrganizationID: %d", organizationID)
		c.Set("organizationID", organizationID)

		return next(c)
	}
}
