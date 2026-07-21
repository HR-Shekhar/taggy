package middleware

import (
	"time"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

// Logger logs every HTTP request after it has been processed.
func Logger(log *zerolog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {

		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			requestID, _ := c.Get(RequestIDKey).(string)

			req := c.Request()
			res := c.Response()
			path := c.Path()
			// for unmatched paths(404s), path not exist
			if path == "" {
				path = req.URL.Path
			}

			event := log.Info()

			switch {
			case res.Status >= 500:
				event = log.Error()
			case res.Status >= 400:
				event = log.Warn()
			}

			event.
				Str("request_id", requestID).
				Str("method", req.Method).
				Str("path", c.Path()).
				Int("status", res.Status).
				Dur("latency", time.Since(start)).
				Str("ip", c.RealIP()).
				Str("user_agent", req.UserAgent()).
				Int64("bytes_out", res.Size).
				Msg("HTTP request")

			return err
		}
	}
}