package middleware

import (
	echo "github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

/* Disabling XFrameOptions
A user could be tricked into clicking your UI 
while thinking they're clicking something else (clickjacking).*/

/*H STSMaxAge: 31536000, -- One year.
Once a browser sees this over HTTPS, it remembers: "Never use HTTP for this site again."
Don't enable HSTS while developing over plain HTTP.
It's a production setting for HTTPS deployments.*/

func Security() echo.MiddlewareFunc {
	return echoMiddleware.SecureWithConfig(
		echoMiddleware.SecureConfig{
			XSSProtection:         "1; mode=block",
			ContentTypeNosniff:    "nosniff",
			XFrameOptions:         "DENY",
			HSTSMaxAge:            31536000,
			HSTSExcludeSubdomains: false,
			ContentSecurityPolicy: "default-src 'self'",
			ReferrerPolicy:        "strict-origin-when-cross-origin",
		},
	)
}