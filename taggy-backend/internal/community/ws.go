package community

import (
	"net/http"
	"strings"

	"github.com/HR-Shekhar/taggy-backend/internal/security/jwt"
	apperrors "github.com/HR-Shekhar/taggy-backend/internal/shared/errors"
	"github.com/HR-Shekhar/taggy-backend/internal/shared/logging"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// ServeChatWS upgrades to WebSocket for live message fan-out.
//
// Query:
//   - token (required) — access JWT
//   - room (required) — `pod:{podSlug}` or `channel:{skillSlug}:{channelSlug}`
func (h *Handler) ServeChatWS(c echo.Context) error {
	log := logging.FromEcho(c, h.log)

	token := strings.TrimSpace(c.QueryParam("token"))
	if token == "" {
		return apperrors.ErrUnauthorized
	}
	if h.jwt == nil {
		return apperrors.ErrUnauthorized
	}

	claims, err := h.jwt.Verify(token)
	if err != nil {
		log.Warn().Err(err).Msg("chat ws jwt verification failed")
		return apperrors.ErrUnauthorized
	}

	userPublicID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return apperrors.ErrUnauthorized
	}

	room := strings.TrimSpace(c.QueryParam("room"))
	if room == "" {
		return apperrors.ErrBadRequest
	}

	if err := h.service.AuthorizeRealtimeRoom(c.Request().Context(), userPublicID, room); err != nil {
		return err
	}

	if h.hub == nil {
		return echo.NewHTTPError(http.StatusServiceUnavailable, "chat realtime unavailable")
	}

	conn, err := h.hub.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		log.Warn().Err(err).Msg("chat ws upgrade failed")
		return nil
	}

	client := &wsClient{
		hub:  h.hub,
		room: room,
		conn: conn,
		send: make(chan []byte, 32),
	}
	h.hub.subscribe(room, client)

	log.Info().
		Str("room", room).
		Str("user_id", userPublicID.String()).
		Msg("chat ws connected")

	go client.writePump()
	client.readPump()
	return nil
}

// jwt verifier used by the websocket handler.
type tokenVerifier interface {
	Verify(tokenString string) (*jwt.Claims, error)
}
