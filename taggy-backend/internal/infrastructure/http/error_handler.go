package http

import (
	"errors"
	"net/http"

	apperrors "github.com/HR-Shekhar/taggy-backend/internal/shared/errors"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

// ErrorResponse is the JSON structure returned to API clients whenever
// an error occurs.
//
// Keep this intentionally small for now.
// We can extend it later (e.g. validation fields, error codes)
// without changing the overall architecture.
type ErrorResponse struct {
	Message string `json:"message"`
}

// ErrorHandler returns Echo's global HTTP error handler.
//
// Any handler that returns an error will eventually end up here.
//
// Handler
//     ↓
// return err
//     ↓
// ErrorHandler
//     ↓
// JSON Response
func ErrorHandler(log zerolog.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {

		// If Echo has already started writing the response,
		// we cannot change the status code or response body.
		//
		// This usually happens when a handler has already called
		// c.JSON(), c.String(), c.Blob(), etc.
		if c.Response().Committed {
			return
		}

		var (
			status  int
			message string
		)

		// ----------------------------------------------------
		// Handle Echo-generated HTTP errors first.
		//
		// Examples:
		//   - echo.NewHTTPError(...)
		//   - automatic 404
		//   - automatic 405
		//
		// We preserve the status code that Echo has already decided.
		// ----------------------------------------------------
		var httpErr *echo.HTTPError
		if errors.As(err, &httpErr) {
			status = httpErr.Code

			if msg, ok := httpErr.Message.(string); ok {
				message = msg
			} else {
				message = http.StatusText(status)
			}

		} else {

			// ------------------------------------------------
			// Handle our application errors.
			//
			// We use errors.Is() so wrapped errors still match.
			// ------------------------------------------------
			switch {

			case errors.Is(err, apperrors.ErrBadRequest):
				status = http.StatusBadRequest
				message = err.Error()

			case errors.Is(err, apperrors.ErrUnauthorized):
				status = http.StatusUnauthorized
				message = err.Error()

			case errors.Is(err, apperrors.ErrForbidden):
				status = http.StatusForbidden
				message = err.Error()

			case errors.Is(err, apperrors.ErrNotFound):
				status = http.StatusNotFound
				message = err.Error()

			case errors.Is(err, apperrors.ErrConflict):
				status = http.StatusConflict
				message = err.Error()

			default:
				status = http.StatusInternalServerError
				message = "internal server error"
			}
		}

		// ----------------------------------------------------
		// Log only unexpected server errors.
		//
		// 4xx errors are usually caused by invalid client input
		// and are already visible in the request logger.
		// ----------------------------------------------------
		if status >= 500 {
			log.Error().
				Err(err).
				Str("method", c.Request().Method).
				Str("path", c.Path()).
				Msg("request failed")
		}

		// Send a consistent JSON response to the client.
		_ = c.JSON(status, ErrorResponse{
			Message: message,
		})
	}
}