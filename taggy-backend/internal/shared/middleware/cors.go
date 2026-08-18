package middleware

import (
	echo "github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

func CORS() echo.MiddlewareFunc {
	return echoMiddleware.CORSWithConfig(
		echoMiddleware.CORSConfig{
			AllowOrigins: []string{
				"http://localhost:3000",
				"http://localhost:5173",
			},
			AllowMethods: []string{
				echo.GET,
				echo.POST,
				echo.PUT,
				echo.PATCH,
				echo.DELETE,
				echo.OPTIONS,
			},
			AllowHeaders: []string{
				echo.HeaderOrigin,
				echo.HeaderContentType,
				echo.HeaderAccept,
				echo.HeaderAuthorization,
				RequestIDHeader,
			},
			// allows browser to read custom headers
			ExposeHeaders: []string{
				RequestIDHeader,
			},
			// This allows cookies or authenticated browser requests.
			AllowCredentials: true,
		},
	)
}
