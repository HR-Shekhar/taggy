package auth

import (
	"context"

	apperrors "github.com/HR-Shekhar/taggy-backend/internal/shared/errors"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ResolveScopedUserPublicID maps a path username to the user's public ID.
// The authenticated user must own that username (private data only for self).
func ResolveScopedUserPublicID(
	c echo.Context,
	username string,
	getByUsername func(ctx context.Context, username string) (uuid.UUID, error),
) (uuid.UUID, error) {
	authPublicID, err := UserPublicIDFromContext(c)
	if err != nil {
		return uuid.UUID{}, err
	}

	targetPublicID, err := getByUsername(c.Request().Context(), username)
	if err != nil {
		return uuid.UUID{}, apperrors.ErrNotFound
	}

	if targetPublicID != authPublicID {
		return uuid.UUID{}, apperrors.ErrForbidden
	}

	return targetPublicID, nil
}

// UserPublicIDFromContext reads the authenticated user's public ID from JWT claims.
func UserPublicIDFromContext(c echo.Context) (uuid.UUID, error) {
	claims, ok := ClaimsFromContext(c)
	if !ok {
		return uuid.UUID{}, apperrors.ErrUnauthorized
	}

	publicID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.UUID{}, apperrors.ErrUnauthorized
	}

	return publicID, nil
}

// ViewerPublicIDFromContext returns the viewer's public ID when a valid JWT is present.
func ViewerPublicIDFromContext(c echo.Context) (uuid.UUID, bool) {
	claims, ok := ClaimsFromContext(c)
	if !ok {
		return uuid.UUID{}, false
	}

	publicID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.UUID{}, false
	}

	return publicID, true
}
