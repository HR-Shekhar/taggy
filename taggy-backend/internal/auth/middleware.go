package auth

import (
	"context"
	"strings"

	"github.com/HR-Shekhar/taggy-backend/internal/security/jwt"
	apperrors "github.com/HR-Shekhar/taggy-backend/internal/shared/errors"
	"github.com/HR-Shekhar/taggy-backend/internal/shared/logging"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

const claimsContextKey = "auth_claims"

// JWT returns middleware that verifies Bearer access tokens and stores
// parsed claims in the request context for downstream handlers.
func JWT(jwtService *jwt.Service, log zerolog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			reqLog := logging.FromEcho(c, log)

			authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
			if authHeader == "" {
				reqLog.Warn().Msg("missing authorization header")
				return apperrors.ErrUnauthorized
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				reqLog.Warn().Msg("invalid authorization header format")
				return apperrors.ErrUnauthorized
			}

			claims, err := jwtService.Verify(parts[1])
			if err != nil {
				reqLog.Warn().Err(err).Msg("jwt verification failed")
				return apperrors.ErrUnauthorized
			}

			c.Set(claimsContextKey, claims)
			return next(c)
		}
	}
}

// ClaimsFromContext reads JWT claims stored by the JWT middleware.
func ClaimsFromContext(c echo.Context) (*jwt.Claims, bool) {
	claims, ok := c.Get(claimsContextKey).(*jwt.Claims)
	return claims, ok
}

// OptionalJWT attaches claims when a valid Bearer token is present; otherwise continues.
// Used for GitHub-style profile reads where auth unlocks private fields for the owner.
func OptionalJWT(jwtService *jwt.Service, log zerolog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get(echo.HeaderAuthorization)
			if authHeader == "" {
				return next(c)
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				return next(c)
			}

			claims, err := jwtService.Verify(parts[1])
			if err != nil {
				return next(c)
			}

			c.Set(claimsContextKey, claims)
			return next(c)
		}
	}
}

// RequireAdmin ensures the authenticated user has is_admin=true (loaded from DB).
func RequireAdmin(
	lookup func(ctx context.Context, publicID uuid.UUID) (bool, error),
	log zerolog.Logger,
) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			reqLog := logging.FromEcho(c, log)
			publicID, err := UserPublicIDFromContext(c)
			if err != nil {
				return err
			}
			ok, err := lookup(c.Request().Context(), publicID)
			if err != nil {
				reqLog.Warn().Err(err).Msg("admin lookup failed")
				return apperrors.ErrUnauthorized
			}
			if !ok {
				reqLog.Warn().Str("user_id", publicID.String()).Msg("admin access denied")
				return apperrors.ErrForbidden
			}
			return next(c)
		}
	}
}
