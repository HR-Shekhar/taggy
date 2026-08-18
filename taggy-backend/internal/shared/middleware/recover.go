package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

// Recover creates middleware that catches panics, logs them using Zerolog,
// and prevents the server from crashing.
func Recover(log zerolog.Logger) echo.MiddlewareFunc {

	// This function is called once during application startup.
	return func(next echo.HandlerFunc) echo.HandlerFunc {

		// This function is created once for each registered route.
		return func(c echo.Context) (err error) {

			// defer runs when this function returns.
			// If a panic occurs anywhere below, recover() will catch it.
			defer func() {

				// recover() returns the panic value.
				// If there was no panic, it returns nil.
				if r := recover(); r != nil {

					log.Error().
						Interface("panic", r).
						Str("method", c.Request().Method).
						Str("path", c.Path()).
						Str("stack", string(debug.Stack())).
						Msg("Unhandled panic")

					err = echo.NewHTTPError(
						http.StatusInternalServerError,
						"Internal Server Error",
					)
				}
			}()

			// Execute the next middleware or the final handler.
			return next(c)
		}
	}
}

/*
Why is the return value named?

Notice something unusual:

func(c echo.Context) (err error)

instead of

func(c echo.Context) error

Why?

Because defer runs after the function has started returning.

Suppose:

return next(c)

panics.

We never reach a normal return.

Instead,

the deferred function runs.

Inside the deferred function we do

err = echo.NewHTTPError(...)

Since err is the named return variable,
when the deferred function finishes,
Go automatically returns it.
*/
