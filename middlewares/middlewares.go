package middleware

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

// AdminMiddleware checks if the user has the admin role
func AdminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Extract user role from context or request header
		role := c.Request().Header.Get("Role") // Assuming Role is set in request header for simplicity
		if role != "2" {                       // '2' is assumed to be the admin role ID
			return c.JSON(http.StatusForbidden, map[string]string{"message": "Access denied"})
		}
		return next(c)
	}
}
