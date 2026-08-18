package notification

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

func (h *Handler) List(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := scopedUserPublicID(c, h.service)
	if err != nil {
		return err
	}

	unreadOnly := c.QueryParam("unread_only") == "true" || c.QueryParam("unread_only") == "1"
	var limit int32
	if raw := c.QueryParam("limit"); raw != "" {
		n, err := strconv.ParseInt(raw, 10, 32)
		if err != nil || n < 0 {
			return apperrors.ErrBadRequest
		}
		limit = int32(n)
	}

	rows, unread, err := h.service.List(c.Request().Context(), userPublicID, unreadOnly, limit)
	if err != nil {
		return err
	}

	resp := make([]notificationResponse, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, toResponse(row))
	}

	log.Info().Int("count", len(resp)).Int64("unread", unread).Msg("list notifications handled")
	return c.JSON(http.StatusOK, listNotificationsResponse{
		UnreadCount:   unread,
		Notifications: resp,
	})
}

func (h *Handler) MarkRead(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := scopedUserPublicID(c, h.service)
	if err != nil {
		return err
	}

	id, err := parseID(c.Param("id"))
	if err != nil {
		return err
	}

	row, err := h.service.MarkRead(c.Request().Context(), userPublicID, id)
	if err != nil {
		return err
	}

	log.Info().Int64("notification_id", id).Msg("mark notification read handled")
	return c.JSON(http.StatusOK, toResponse(row))
}

func (h *Handler) MarkAllRead(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := scopedUserPublicID(c, h.service)
	if err != nil {
		return err
	}

	n, err := h.service.MarkAllRead(c.Request().Context(), userPublicID)
	if err != nil {
		return err
	}

	log.Info().Int64("updated", n).Msg("mark all notifications read handled")
	return c.JSON(http.StatusOK, map[string]int64{"updated": n})
}

func (h *Handler) ClearRead(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := scopedUserPublicID(c, h.service)
	if err != nil {
		return err
	}

	n, err := h.service.ClearRead(c.Request().Context(), userPublicID)
	if err != nil {
		return err
	}

	log.Info().Int64("deleted", n).Msg("clear read notifications handled")
	return c.JSON(http.StatusOK, map[string]int64{"deleted": n})
}

func scopedUserPublicID(c echo.Context, service *Service) (uuid.UUID, error) {
	return auth.ResolveScopedUserPublicID(
		c,
		c.Param("username"),
		service.GetUserPublicIDByUsername,
	)
}

func parseID(raw string) (int64, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, apperrors.ErrBadRequest
	}
	return id, nil
}

func toResponse(row sqlc.Notification) notificationResponse {
	resp := notificationResponse{
		ID:        row.ID,
		Type:      string(row.Type),
		Title:     row.Title,
		Body:      row.Body,
		IsRead:    row.IsRead,
		CreatedAt: formatTime(row.CreatedAt),
	}
	if row.EntityType.Valid {
		resp.EntityType = &row.EntityType.String
	}
	if row.EntityID.Valid {
		v := row.EntityID.Int64
		resp.EntityID = &v
	}
	if row.ReadAt.Valid {
		s := row.ReadAt.Time.UTC().Format(time.RFC3339)
		resp.ReadAt = &s
	}
	return resp
}

func formatTime(t pgtype.Timestamptz) string {
	if !t.Valid {
		return ""
	}
	return t.Time.UTC().Format(time.RFC3339)
}
