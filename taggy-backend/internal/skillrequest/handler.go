package skillrequest

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

func (h *Handler) ListSimilar(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	if _, err := auth.UserPublicIDFromContext(c); err != nil {
		return err
	}
	q := c.QueryParam("q")
	if q == "" {
		q = c.QueryParam("query")
	}
	similar, err := h.service.ListSimilar(c.Request().Context(), q)
	if err != nil {
		return err
	}
	log.Info().Str("query", q).Int("count", len(similar)).Msg("similar skills listed")
	return c.JSON(http.StatusOK, similarResponse{Query: q, Similar: similar})
}

func (h *Handler) Create(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		return err
	}
	var req createRequestBody
	if err := c.Bind(&req); err != nil {
		return apperrors.ErrBadRequest
	}
	if err := c.Validate(&req); err != nil {
		return err
	}
	result, err := h.service.Create(c.Request().Context(), userPublicID, CreateInput{
		Name:        req.Name,
		Description: req.Description,
		Force:       req.Force,
	})
	if err != nil {
		return err
	}
	if result.RequiresConfirm {
		log.Info().Str("name", req.Name).Int("similar", len(result.Similar)).Msg("skill request requires confirm")
		return c.JSON(http.StatusOK, createResponse{
			RequiresConfirm: true,
			Similar:         result.Similar,
			Message:         "Similar skills found. Review them or resubmit with force=true.",
		})
	}
	log.Info().Str("request_id", result.Request.ID).Msg("skill request accepted for generation")
	return c.JSON(http.StatusAccepted, createResponse{Request: &result.Request})
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
	log.Info().Int("count", len(rows)).Msg("skill requests listed")
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
	log.Info().Str("request_id", id.String()).Msg("skill request cancelled")
	return c.JSON(http.StatusOK, row)
}

func (h *Handler) AdminList(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	rows, err := h.service.ListPendingAdmin(c.Request().Context())
	if err != nil {
		return err
	}
	log.Info().Int("count", len(rows)).Msg("admin skill requests listed")
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
	log.Info().Str("request_id", id.String()).Msg("admin approved skill request")
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
	log.Info().Str("request_id", id.String()).Msg("admin rejected skill request")
	return c.JSON(http.StatusOK, row)
}
