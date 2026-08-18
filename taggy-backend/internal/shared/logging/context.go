package logging

import (
	"github.com/HR-Shekhar/taggy-backend/internal/shared/middleware"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

// FromEcho enriches logs with request-scoped fields available in handlers/middleware.
func FromEcho(c echo.Context, log zerolog.Logger) zerolog.Logger {
	requestID, _ := c.Get(middleware.RequestIDKey).(string)

	return log.With().
		Str("request_id", requestID).
		Str("method", c.Request().Method).
		Str("path", c.Path()).
		Logger()
}
