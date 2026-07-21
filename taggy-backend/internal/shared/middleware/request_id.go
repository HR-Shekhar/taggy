package middleware

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// RequestID creates middleware that assigns every request
// a unique ID and makes it available throughout the request lifecycle.
func RequestID() echo.MiddlewareFunc {

	// Runs once during application startup.
	return func(next echo.HandlerFunc) echo.HandlerFunc {

		// Runs once for each registered route.
		return func(c echo.Context) error {

			// Generate a new UUID for this request if it is not set already by something like reverse proxy(ngnix).
			requestID := c.Request().Header.Get(RequestIDHeader)

			if requestID == "" {
				requestID = uuid.NewString()
			}
			// Store it inside Echo's request context.
			// Any middleware or handler can retrieve it later.
			c.Set(RequestIDKey, requestID)

			// Also expose it to the client via a response header.
			c.Response().Header().Set(RequestIDHeader, requestID)

			// Continue executing the remaining middleware/handler chain.
			return next(c)
		}
	}
}