package report

import (
	"net/http"
	"strconv"
	"time"

	"github.com/HR-Shekhar/taggy-backend/internal/auth"
	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/postgres/sqlc"
	apperrors "github.com/HR-Shekhar/taggy-backend/internal/shared/errors"
	"github.com/HR-Shekhar/taggy-backend/internal/shared/logging"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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

	var req createReportRequest
	if err := c.Bind(&req); err != nil {
		return apperrors.ErrBadRequest
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	created, err := h.service.Create(c.Request().Context(), userPublicID, CreateInput{
		TargetType: req.TargetType,
		TargetID:   req.TargetID,
		Reason:     req.Reason,
	})
	if err != nil {
		return err
	}

	log.Info().Int64("report_id", created.ID).Msg("create report handled")
	return c.JSON(http.StatusCreated, toResponse(created))
}

func (h *Handler) ListMine(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := scopedUserPublicID(c, h.service)
	if err != nil {
		return err
	}

	var limit int32
	if raw := c.QueryParam("limit"); raw != "" {
		n, err := strconv.ParseInt(raw, 10, 32)
		if err != nil || n < 0 {
			return apperrors.ErrBadRequest
		}
		limit = int32(n)
	}

	rows, err := h.service.ListMine(c.Request().Context(), userPublicID, limit)
	if err != nil {
		return err
	}

	resp := make([]reportResponse, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, toResponse(row))
	}

	log.Info().Int("count", len(resp)).Msg("list my reports handled")
	return c.JSON(http.StatusOK, resp)
}

func scopedUserPublicID(c echo.Context, service *Service) (uuid.UUID, error) {
	return auth.ResolveScopedUserPublicID(
		c,
		c.Param("username"),
		service.GetUserPublicIDByUsername,
	)
}

func toResponse(row sqlc.Report) reportResponse {
	resp := reportResponse{
		ID:         row.ID,
		TargetType: string(row.TargetType),
		TargetID:   row.TargetID,
		Reason:     row.Reason,
		Status:     string(row.Status),
		CreatedAt:  formatTime(row.CreatedAt),
	}
	if row.ResolvedAt.Valid {
		s := row.ResolvedAt.Time.UTC().Format(time.RFC3339)
		resp.ResolvedAt = &s
	}
	return resp
}

func formatTime(t pgtype.Timestamptz) string {
	if !t.Valid {
		return ""
	}
	return t.Time.UTC().Format(time.RFC3339)
}
