package api

import "github.com/labstack/echo/v4"

func RegisterRoutes(e *echo.Echo) {
	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status": "ok",
		})
	})
}