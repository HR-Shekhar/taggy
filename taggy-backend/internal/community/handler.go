package community

import (
	"net/http"
	"strconv"
	"time"

	"github.com/HR-Shekhar/taggy-backend/internal/auth"
	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/postgres/sqlc"
	apperrors "github.com/HR-Shekhar/taggy-backend/internal/shared/errors"
	"github.com/HR-Shekhar/taggy-backend/internal/shared/logging"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

type Handler struct {
	service *Service
	hub     *Hub
	jwt     tokenVerifier
	log     zerolog.Logger
}

func NewHandler(service *Service, hub *Hub, jwt tokenVerifier, log zerolog.Logger) *Handler {
	return &Handler{service: service, hub: hub, jwt: jwt, log: log}
}

func (h *Handler) GetCommunity(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		return err
	}

	skillSlug := c.Param("skillSlug")
	if skillSlug == "" {
		return apperrors.ErrBadRequest
	}

	community, err := h.service.GetCommunity(c.Request().Context(), userPublicID, skillSlug)
	if err != nil {
		return err
	}

	log.Info().Str("skill_slug", skillSlug).Msg("get community handled")
	return c.JSON(http.StatusOK, toCommunityResponse(community))
}

func (h *Handler) ListChannels(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		return err
	}

	skillSlug := c.Param("skillSlug")
	if skillSlug == "" {
		return apperrors.ErrBadRequest
	}

	channels, err := h.service.ListChannels(c.Request().Context(), userPublicID, skillSlug)
	if err != nil {
		return err
	}

	resp := make([]channelResponse, 0, len(channels))
	for _, ch := range channels {
		resp = append(resp, toChannelResponse(ch))
	}

	log.Info().Str("skill_slug", skillSlug).Int("count", len(resp)).Msg("list channels handled")
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) ListChannelMessages(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		return err
	}

	skillSlug := c.Param("skillSlug")
	channelSlug := c.Param("channelSlug")
	if skillSlug == "" || channelSlug == "" {
		return apperrors.ErrBadRequest
	}

	input, err := parseListMessagesInput(c)
	if err != nil {
		return err
	}

	rows, err := h.service.ListChannelMessages(c.Request().Context(), userPublicID, skillSlug, channelSlug, input)
	if err != nil {
		return err
	}

	resp := make([]messageResponse, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, toChannelMessageResponse(row))
	}

	log.Info().
		Str("skill_slug", skillSlug).
		Str("channel_slug", channelSlug).
		Int("count", len(resp)).
		Msg("list channel messages handled")
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) SendChannelMessage(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		return err
	}

	skillSlug := c.Param("skillSlug")
	channelSlug := c.Param("channelSlug")
	if skillSlug == "" || channelSlug == "" {
		return apperrors.ErrBadRequest
	}

	var req sendMessageRequest
	if err := c.Bind(&req); err != nil {
		return apperrors.ErrBadRequest
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	msg, err := h.service.SendChannelMessage(c.Request().Context(), userPublicID, skillSlug, channelSlug, SendMessageInput{
		Content:          req.Content,
		ReplyToMessageID: req.ReplyToMessageID,
	})
	if err != nil {
		return err
	}

	log.Info().
		Str("skill_slug", skillSlug).
		Str("channel_slug", channelSlug).
		Int64("message_id", msg.ID).
		Msg("send channel message handled")
	return c.JSON(http.StatusCreated, toGetMessageResponse(msg))
}

func (h *Handler) ListPodMessages(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		return err
	}

	podSlug := c.Param("podSlug")
	if podSlug == "" {
		return apperrors.ErrBadRequest
	}

	input, err := parseListMessagesInput(c)
	if err != nil {
		return err
	}

	rows, err := h.service.ListPodMessages(c.Request().Context(), userPublicID, podSlug, input)
	if err != nil {
		return err
	}

	resp := make([]messageResponse, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, toPodMessageResponse(row))
	}

	log.Info().Str("pod_slug", podSlug).Int("count", len(resp)).Msg("list pod messages handled")
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) SendPodMessage(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		return err
	}

	podSlug := c.Param("podSlug")
	if podSlug == "" {
		return apperrors.ErrBadRequest
	}

	var req sendMessageRequest
	if err := c.Bind(&req); err != nil {
		return apperrors.ErrBadRequest
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	msg, err := h.service.SendPodMessage(c.Request().Context(), userPublicID, podSlug, SendMessageInput{
		Content:          req.Content,
		ReplyToMessageID: req.ReplyToMessageID,
	})
	if err != nil {
		return err
	}

	log.Info().Str("pod_slug", podSlug).Int64("message_id", msg.ID).Msg("send pod message handled")
	return c.JSON(http.StatusCreated, toGetMessageResponse(msg))
}

func (h *Handler) EditMessage(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		return err
	}

	messageID, err := parseMessageID(c.Param("id"))
	if err != nil {
		return err
	}

	var req editMessageRequest
	if err := c.Bind(&req); err != nil {
		return apperrors.ErrBadRequest
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	msg, err := h.service.EditMessage(c.Request().Context(), userPublicID, messageID, SendMessageInput{
		Content: req.Content,
	})
	if err != nil {
		return err
	}

	log.Info().Int64("message_id", messageID).Msg("edit message handled")
	return c.JSON(http.StatusOK, toGetMessageResponse(msg))
}

func (h *Handler) DeleteMessage(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		return err
	}

	messageID, err := parseMessageID(c.Param("id"))
	if err != nil {
		return err
	}

	if err := h.service.DeleteMessage(c.Request().Context(), userPublicID, messageID); err != nil {
		return err
	}

	log.Info().Int64("message_id", messageID).Msg("delete message handled")
	return c.NoContent(http.StatusNoContent)
}

func parseListMessagesInput(c echo.Context) (ListMessagesInput, error) {
	input := ListMessagesInput{}

	if before := c.QueryParam("before"); before != "" {
		id, err := strconv.ParseInt(before, 10, 64)
		if err != nil || id < 0 {
			return ListMessagesInput{}, apperrors.ErrBadRequest
		}
		input.BeforeID = id
	}

	if limitRaw := c.QueryParam("limit"); limitRaw != "" {
		limit, err := strconv.ParseInt(limitRaw, 10, 32)
		if err != nil || limit < 0 {
			return ListMessagesInput{}, apperrors.ErrBadRequest
		}
		input.Limit = int32(limit)
	}

	return input, nil
}

func parseMessageID(raw string) (int64, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, apperrors.ErrBadRequest
	}
	return id, nil
}

func toCommunityResponse(row sqlc.GetCommunityBySkillSlugRow) communityResponse {
	resp := communityResponse{
		Slug: row.SkillSlug,
		Name: row.Name,
	}
	if row.Description.Valid {
		resp.Description = &row.Description.String
	}
	return resp
}

func toChannelResponse(row sqlc.ListChannelsByCommunityIDRow) channelResponse {
	resp := channelResponse{
		ID:   row.ID,
		Slug: row.Slug,
		Name: row.Name,
	}
	if row.Description.Valid {
		resp.Description = &row.Description.String
	}
	return resp
}

func toChannelMessageResponse(row sqlc.ListChannelMessagesRow) messageResponse {
	return messageResponse{
		ID:                    row.ID,
		AuthorUsername:        row.AuthorUsername,
		AuthorName:            row.AuthorName,
		Content:               row.Content,
		EditedAt:              formatOptionalTime(row.EditedAt),
		CreatedAt:             formatRequiredTime(row.CreatedAt),
		ReplyToMessageID:      optionalInt64Ptr(row.ReplyToMessageID),
		ReplyToContent:        optionalTextPtr(row.ReplyToContent),
		ReplyToAuthorUsername: optionalTextPtr(row.ReplyToAuthorUsername),
	}
}

func toPodMessageResponse(row sqlc.ListPodMessagesRow) messageResponse {
	return messageResponse{
		ID:                    row.ID,
		AuthorUsername:        row.AuthorUsername,
		AuthorName:            row.AuthorName,
		Content:               row.Content,
		EditedAt:              formatOptionalTime(row.EditedAt),
		CreatedAt:             formatRequiredTime(row.CreatedAt),
		ReplyToMessageID:      optionalInt64Ptr(row.ReplyToMessageID),
		ReplyToContent:        optionalTextPtr(row.ReplyToContent),
		ReplyToAuthorUsername: optionalTextPtr(row.ReplyToAuthorUsername),
	}
}

func toGetMessageResponse(row sqlc.GetMessageByIDRow) messageResponse {
	return messageResponse{
		ID:                    row.ID,
		AuthorUsername:        row.AuthorUsername,
		AuthorName:            row.AuthorName,
		Content:               row.Content,
		EditedAt:              formatOptionalTime(row.EditedAt),
		CreatedAt:             formatRequiredTime(row.CreatedAt),
		ReplyToMessageID:      optionalInt64Ptr(row.ReplyToMessageID),
		ReplyToContent:        optionalTextPtr(row.ReplyToContent),
		ReplyToAuthorUsername: optionalTextPtr(row.ReplyToAuthorUsername),
	}
}

func optionalInt64Ptr(v pgtype.Int8) *int64 {
	if !v.Valid {
		return nil
	}
	id := v.Int64
	return &id
}

func optionalTextPtr(v pgtype.Text) *string {
	if !v.Valid {
		return nil
	}
	s := v.String
	return &s
}

func formatOptionalTime(t pgtype.Timestamptz) *string {
	if !t.Valid {
		return nil
	}
	s := t.Time.UTC().Format(time.RFC3339)
	return &s
}

func formatRequiredTime(t pgtype.Timestamptz) string {
	if !t.Valid {
		return ""
	}
	return t.Time.UTC().Format(time.RFC3339)
}
