package audio

import (
	"net/http"
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

func (h *Handler) CreatePodRoom(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		return err
	}

	podSlug := c.Param("podSlug")
	if podSlug == "" {
		return apperrors.ErrBadRequest
	}

	input, err := bindCreateRoom(c)
	if err != nil {
		return err
	}

	room, err := h.service.CreatePodRoom(c.Request().Context(), userPublicID, podSlug, input)
	if err != nil {
		return err
	}

	log.Info().Str("pod_slug", podSlug).Str("room_id", room.PublicID.String()).Msg("create pod audio room handled")
	return c.JSON(http.StatusCreated, toRoomResponse(room))
}

func (h *Handler) ListPodRooms(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		return err
	}

	podSlug := c.Param("podSlug")
	if podSlug == "" {
		return apperrors.ErrBadRequest
	}

	rows, err := h.service.ListPodRooms(c.Request().Context(), userPublicID, podSlug, c.QueryParam("status"))
	if err != nil {
		return err
	}

	resp := make([]roomResponse, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, toListPodRoomResponse(row, podSlug))
	}

	log.Info().Str("pod_slug", podSlug).Int("count", len(resp)).Msg("list pod audio rooms handled")
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) CreateChannelRoom(c echo.Context) error {
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

	input, err := bindCreateRoom(c)
	if err != nil {
		return err
	}

	room, err := h.service.CreateChannelRoom(c.Request().Context(), userPublicID, skillSlug, channelSlug, input)
	if err != nil {
		return err
	}

	log.Info().
		Str("skill_slug", skillSlug).
		Str("channel_slug", channelSlug).
		Str("room_id", room.PublicID.String()).
		Msg("create channel audio room handled")
	return c.JSON(http.StatusCreated, toRoomResponse(room))
}

func (h *Handler) ListChannelRooms(c echo.Context) error {
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

	rows, err := h.service.ListChannelRooms(
		c.Request().Context(),
		userPublicID,
		skillSlug,
		channelSlug,
		c.QueryParam("status"),
	)
	if err != nil {
		return err
	}

	resp := make([]roomResponse, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, toListChannelRoomResponse(row, skillSlug, channelSlug))
	}

	log.Info().
		Str("skill_slug", skillSlug).
		Str("channel_slug", channelSlug).
		Int("count", len(resp)).
		Msg("list channel audio rooms handled")
	return c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetRoom(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		return err
	}

	roomID, err := parseRoomID(c.Param("roomId"))
	if err != nil {
		return err
	}

	room, participants, err := h.service.GetRoom(c.Request().Context(), userPublicID, roomID)
	if err != nil {
		return err
	}

	partResp := make([]participantResponse, 0, len(participants))
	for _, p := range participants {
		partResp = append(partResp, participantResponse{
			Username: p.Username,
			Name:     p.UserName,
			Role:     string(p.Role),
			JoinedAt: formatRequiredTime(p.JoinedAt),
		})
	}

	log.Info().Str("room_id", roomID.String()).Msg("get audio room handled")
	return c.JSON(http.StatusOK, roomDetailResponse{
		Room:         toRoomResponse(room),
		Participants: partResp,
	})
}

func (h *Handler) JoinRoom(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		return err
	}

	roomID, err := parseRoomID(c.Param("roomId"))
	if err != nil {
		return err
	}

	result, err := h.service.JoinRoom(c.Request().Context(), userPublicID, roomID)
	if err != nil {
		return err
	}

	log.Info().Str("room_id", roomID.String()).Msg("join audio room handled")
	return c.JSON(http.StatusOK, joinResponse{
		RoomID:          result.RoomID,
		LiveKitURL:      result.LiveKitURL,
		LiveKitRoomName: result.LiveKitRoomName,
		Token:           result.Token,
		Role:            result.Role,
	})
}

func (h *Handler) LeaveRoom(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		return err
	}

	roomID, err := parseRoomID(c.Param("roomId"))
	if err != nil {
		return err
	}

	if err := h.service.LeaveRoom(c.Request().Context(), userPublicID, roomID); err != nil {
		return err
	}

	log.Info().Str("room_id", roomID.String()).Msg("leave audio room handled")
	return c.NoContent(http.StatusNoContent)
}

func (h *Handler) EndRoom(c echo.Context) error {
	log := logging.FromEcho(c, h.log)
	userPublicID, err := auth.UserPublicIDFromContext(c)
	if err != nil {
		return err
	}

	roomID, err := parseRoomID(c.Param("roomId"))
	if err != nil {
		return err
	}

	if err := h.service.EndRoom(c.Request().Context(), userPublicID, roomID); err != nil {
		return err
	}

	log.Info().Str("room_id", roomID.String()).Msg("end audio room handled")
	return c.NoContent(http.StatusNoContent)
}

func bindCreateRoom(c echo.Context) (CreateRoomInput, error) {
	var req createRoomRequest
	if err := c.Bind(&req); err != nil {
		return CreateRoomInput{}, apperrors.ErrBadRequest
	}
	if err := c.Validate(&req); err != nil {
		return CreateRoomInput{}, err
	}
	return CreateRoomInput{
		Title:           req.Title,
		Description:     req.Description,
		MaxParticipants: req.MaxParticipants,
	}, nil
}

func parseRoomID(raw string) (uuid.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.UUID{}, apperrors.ErrBadRequest
	}
	return id, nil
}

func toRoomResponse(row sqlc.GetAudioRoomByPublicIDRow) roomResponse {
	resp := roomResponse{
		ID:              row.PublicID.String(),
		EntityID:        row.ID,
		RoomType:        string(row.RoomType),
		Title:           row.Title,
		Status:          string(row.Status),
		HostUsername:    row.HostUsername,
		LiveKitRoomName: row.LivekitRoomName,
		ActualStartTime: formatOptionalTime(row.ActualStartTime),
		EndedAt:         formatOptionalTime(row.EndedAt),
		CreatedAt:       formatRequiredTime(row.CreatedAt),
	}
	if row.Description.Valid {
		resp.Description = &row.Description.String
	}
	if row.PodSlug.Valid {
		resp.PodSlug = &row.PodSlug.String
	}
	if row.SkillSlug.Valid {
		resp.SkillSlug = &row.SkillSlug.String
	}
	if row.ChannelSlug.Valid {
		resp.ChannelSlug = &row.ChannelSlug.String
	}
	if row.MaxParticipants.Valid {
		v := row.MaxParticipants.Int32
		resp.MaxParticipants = &v
	}
	return resp
}

func toListPodRoomResponse(row sqlc.ListAudioRoomsByPodIDRow, podSlug string) roomResponse {
	resp := roomResponse{
		ID:              row.PublicID.String(),
		EntityID:        row.ID,
		RoomType:        string(row.RoomType),
		Title:           row.Title,
		Status:          string(row.Status),
		HostUsername:    row.HostUsername,
		LiveKitRoomName: row.LivekitRoomName,
		PodSlug:         &podSlug,
		ActualStartTime: formatOptionalTime(row.ActualStartTime),
		EndedAt:         formatOptionalTime(row.EndedAt),
		CreatedAt:       formatRequiredTime(row.CreatedAt),
	}
	if row.Description.Valid {
		resp.Description = &row.Description.String
	}
	if row.MaxParticipants.Valid {
		v := row.MaxParticipants.Int32
		resp.MaxParticipants = &v
	}
	return resp
}

func toListChannelRoomResponse(row sqlc.ListAudioRoomsByChannelIDRow, skillSlug, channelSlug string) roomResponse {
	resp := roomResponse{
		ID:              row.PublicID.String(),
		EntityID:        row.ID,
		RoomType:        string(row.RoomType),
		Title:           row.Title,
		Status:          string(row.Status),
		HostUsername:    row.HostUsername,
		LiveKitRoomName: row.LivekitRoomName,
		SkillSlug:       &skillSlug,
		ChannelSlug:     &channelSlug,
		ActualStartTime: formatOptionalTime(row.ActualStartTime),
		EndedAt:         formatOptionalTime(row.EndedAt),
		CreatedAt:       formatRequiredTime(row.CreatedAt),
	}
	if row.Description.Valid {
		resp.Description = &row.Description.String
	}
	if row.MaxParticipants.Valid {
		v := row.MaxParticipants.Int32
		resp.MaxParticipants = &v
	}
	return resp
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
