package http

import (
	"github.com/HR-Shekhar/taggy-backend/internal/shared/config"
	"github.com/HR-Shekhar/taggy-backend/internal/shared/middleware"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

// RegisterMiddleware attaches global middleware to the Echo instance.
func RegisterMiddleware(e *echo.Echo, log zerolog.Logger, cfg *config.Config) {
	e.Use(middleware.Recover(log))
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger(&log))

	if cfg.App.Environment == "production" {
		e.Use(middleware.Security())
	}

	e.Use(middleware.CORS())
}
