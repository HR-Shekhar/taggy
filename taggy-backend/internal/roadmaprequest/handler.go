package roadmaprequest

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

func (h *Handler) Create(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		return err
	}
	skillSlug := c.Param("skillSlug")
	if skillSlug == "" {
		return apperrors.ErrBadRequest
	}
	var req createRequestBody
	_ = c.Bind(&req)
	_ = c.Validate(&req)
	row, err := h.service.Create(c.Request().Context(), userPublicID, skillSlug, CreateInput{Rationale: req.Rationale})
	if err != nil {
		return err
	}
	log.Info().Str("skill", skillSlug).Str("request_id", row.ID).Msg("roadmap edit request created")
	return c.JSON(http.StatusCreated, row)
}

func (h *Handler) ListMine(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := auth.ResolveScopedUserPublicID(c, c.Param("username"), h.service.GetUserPublicIDByUsername)
	if err != nil {
		return err
	}
	rows, err := h.service.ListMine(c.Request().Context(), userPublicID)
	if err != nil {
		return err
	}
	log.Info().Int("count", len(rows)).Msg("roadmap edit requests listed")
	return c.JSON(http.StatusOK, rows)
}

func (h *Handler) Cancel(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := auth.ResolveScopedUserPublicID(c, c.Param("username"), h.service.GetUserPublicIDByUsername)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperrors.ErrBadRequest
	}
	row, err := h.service.Cancel(c.Request().Context(), userPublicID, id)
	if err != nil {
		return err
	}
	log.Info().Str("request_id", id.String()).Msg("roadmap edit request cancelled")
	return c.JSON(http.StatusOK, row)
}

func (h *Handler) AdminList(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	rows, err := h.service.ListPendingAdmin(c.Request().Context())
	if err != nil {
		return err
	}
	log.Info().Int("count", len(rows)).Msg("admin roadmap edit requests listed")
	return c.JSON(http.StatusOK, rows)
}

func (h *Handler) AdminApprove(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	adminID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperrors.ErrBadRequest
	}
	row, err := h.service.Approve(c.Request().Context(), adminID, id)
	if err != nil {
		return err
	}
	log.Info().Str("request_id", id.String()).Msg("admin approved roadmap edit request")
	return c.JSON(http.StatusOK, row)
}

func (h *Handler) AdminReject(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	adminID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return apperrors.ErrBadRequest
	}
	var req rejectRequestBody
	_ = c.Bind(&req)
	row, err := h.service.Reject(c.Request().Context(), adminID, id, req.AdminNote)
	if err != nil {
		return err
	}
	log.Info().Str("request_id", id.String()).Msg("admin rejected roadmap edit request")
	return c.JSON(http.StatusOK, row)
}
