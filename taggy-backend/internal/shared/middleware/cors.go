package middleware

import (
	"strings"

	echo "github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

func CORS(frontendURL string) echo.MiddlewareFunc {
	origins := []string{
		"http://localhost:3000",
		"http://127.0.0.1:3000",
		"http://localhost:5173",
		"http://127.0.0.1:5173",
	}
	if extra := strings.TrimRight(strings.TrimSpace(frontendURL), "/"); extra != "" {
		seen := false
		for _, o := range origins {
			if o == extra {
				seen = true
				break
			}
		}
		if !seen {
			origins = append(origins, extra)
		}
	}

	return echoMiddleware.CORSWithConfig(
		echoMiddleware.CORSConfig{
			AllowOrigins: origins,
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
			ExposeHeaders: []string{
				RequestIDHeader,
			},
			AllowCredentials: true,
		},
	)
}
