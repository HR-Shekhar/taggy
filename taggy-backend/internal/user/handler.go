package user

import (
	"net/http"

	"github.com/HR-Shekhar/taggy-backend/internal/auth"
	apperrors "github.com/HR-Shekhar/taggy-backend/internal/shared/errors"
	"github.com/HR-Shekhar/taggy-backend/internal/shared/logging"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

type Handler struct {
	service *Service
	log     zerolog.Logger
}

func NewHandler(service *Service, log zerolog.Logger) *Handler {
	return &Handler{service: service, log: log}
}

// GetProfile is GitHub-style: public profile for anyone; private fields when the viewer is the owner.
func (h *Handler) GetProfile(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	username := c.Param("username")
	if username == "" {
		return apperrors.ErrBadRequest
	}

	var viewerPublicID *uuid.UUID
	if id, ok := auth.ViewerPublicIDFromContext(c); ok {
		viewerPublicID = &id
	}

	user, isOwner, err := h.service.GetProfileByUsername(
		c.Request().Context(),
		username,
		viewerPublicID,
	)
	if err != nil {
		return err
	}

	log.Info().
		Str("username", username).
		Bool("is_owner", isOwner).
		Msg("profile returned")

	return c.JSON(http.StatusOK, toProfileResponse(user, isOwner))
}

func (h *Handler) UpdateProfile(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	viewerPublicID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		log.Warn().Msg("unauthorized profile update")
		return err
	}

	username := c.Param("username")
	if username == "" {
		return apperrors.ErrBadRequest
	}

	var req updateProfileRequest
	if err := c.Bind(&req); err != nil {
		log.Warn().Err(err).Msg("invalid profile update payload")
		return apperrors.ErrBadRequest
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	user, err := h.service.UpdateProfileByUsername(c.Request().Context(), username, viewerPublicID, UpdateProfileInput{
		Name:              req.Name,
		Bio:               req.Bio,
		ProfilePictureURL: req.ProfilePictureURL,
		Username:          req.Username,
	})
	if err != nil {
		return err
	}

	log.Info().
		Str("username", user.Username).
		Msg("profile update handled")

	return c.JSON(http.StatusOK, toProfileResponse(user, true))
}

func (h *Handler) UploadAvatar(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	viewerPublicID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		return err
	}
	username := c.Param("username")
	if username == "" {
		return apperrors.ErrBadRequest
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		log.Warn().Err(err).Msg("avatar upload missing file")
		return ErrInvalidAvatar
	}

	src, err := fileHeader.Open()
	if err != nil {
		return logging.Unexpected(log, err, "avatar upload: open file failed")
	}
	defer src.Close()

	user, err := h.service.UploadAvatar(
		c.Request().Context(),
		username,
		viewerPublicID,
		src,
		fileHeader.Filename,
		fileHeader.Size,
		fileHeader.Header.Get("Content-Type"),
	)
	if err != nil {
		return err
	}

	log.Info().Str("username", user.Username).Msg("avatar upload handled")
	return c.JSON(http.StatusOK, toProfileResponse(user, true))
}
