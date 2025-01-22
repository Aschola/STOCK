package utils

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

// Utility function to get organizationID from context
func GetOrganizationID(c echo.Context) (uint, error) {
	organizationID, ok := c.Get("organizationID").(uint)
	if !ok {
		log.Println("Failed to get organizationID from context")
		return 0, echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}
	return organizationID, nil
}
