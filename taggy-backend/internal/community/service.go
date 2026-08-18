package community

import (
	"context"
	"errors"
	"strings"

	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/postgres/sqlc"
	apperrors "github.com/HR-Shekhar/taggy-backend/internal/shared/errors"
	"github.com/HR-Shekhar/taggy-backend/internal/shared/logging"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
)

const (
	defaultMessageLimit int32 = 50
	maxMessageLimit     int32 = 100
	maxMessageContent         = 4000
)

type Service struct {
	repo *Repository
	hub  *Hub
	log  zerolog.Logger
}

func NewService(repo *Repository, hub *Hub, log zerolog.Logger) *Service {
	return &Service{repo: repo, hub: hub, log: log}
}

func (s *Service) GetCommunity(ctx context.Context, userPublicID uuid.UUID, skillSlug string) (sqlc.GetCommunityBySkillSlugRow, error) {
	if _, err := s.requireEnrolled(ctx, userPublicID, skillSlug); err != nil {
		return sqlc.GetCommunityBySkillSlugRow{}, err
	}

	community, err := s.repo.GetCommunityBySkillSlug(ctx, skillSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.GetCommunityBySkillSlugRow{}, ErrCommunityNotFound
		}
		return sqlc.GetCommunityBySkillSlugRow{}, logging.Unexpected(s.log, err, "get community failed")
	}
	return community, nil
}

func (s *Service) ListChannels(ctx context.Context, userPublicID uuid.UUID, skillSlug string) ([]sqlc.ListChannelsByCommunityIDRow, error) {
	if _, err := s.requireEnrolled(ctx, userPublicID, skillSlug); err != nil {
		return nil, err
	}

	community, err := s.repo.GetCommunityBySkillSlug(ctx, skillSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrCommunityNotFound
		}
		return nil, logging.Unexpected(s.log, err, "list channels: get community failed")
	}

	rows, err := s.repo.ListChannelsByCommunityID(ctx, community.ID)
	if err != nil {
		return nil, logging.Unexpected(s.log, err, "list channels failed")
	}
	return rows, nil
}

func (s *Service) ListChannelMessages(
	ctx context.Context,
	userPublicID uuid.UUID,
	skillSlug, channelSlug string,
	input ListMessagesInput,
) ([]sqlc.ListChannelMessagesRow, error) {
	if _, err := s.requireEnrolled(ctx, userPublicID, skillSlug); err != nil {
		return nil, err
	}

	channel, err := s.repo.GetChannelBySkillSlugAndChannelSlug(ctx, skillSlug, channelSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrChannelNotFound
		}
		return nil, logging.Unexpected(s.log, err, "list channel messages: get channel failed")
	}

	rows, err := s.repo.ListChannelMessages(ctx, channel.ID, input.BeforeID, normalizeLimit(input.Limit))
	if err != nil {
		return nil, logging.Unexpected(s.log, err, "list channel messages failed")
	}
	return reverseChannelMessages(rows), nil
}

func (s *Service) SendChannelMessage(
	ctx context.Context,
	userPublicID uuid.UUID,
	skillSlug, channelSlug string,
	input SendMessageInput,
) (sqlc.GetMessageByIDRow, error) {
	content, err := normalizeContent(input.Content)
	if err != nil {
		return sqlc.GetMessageByIDRow{}, logging.Reject(s.log, err, "send channel message rejected: invalid content")
	}

	user, err := s.requireEnrolled(ctx, userPublicID, skillSlug)
	if err != nil {
		return sqlc.GetMessageByIDRow{}, s.rejectIfDomain(err, "send channel message rejected")
	}

	channel, err := s.repo.GetChannelBySkillSlugAndChannelSlug(ctx, skillSlug, channelSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.GetMessageByIDRow{}, logging.Reject(s.log, ErrChannelNotFound, "send channel message rejected: channel not found")
		}
		return sqlc.GetMessageByIDRow{}, logging.Unexpected(s.log, err, "send channel message: get channel failed")
	}

	if err := s.validateReplyTarget(ctx, input.ReplyToMessageID, channel.ID, 0); err != nil {
		return sqlc.GetMessageByIDRow{}, err
	}

	created, err := s.repo.CreateChannelMessage(ctx, user.ID, channel.ID, content, input.ReplyToMessageID)
	if err != nil {
		return sqlc.GetMessageByIDRow{}, logging.Unexpected(s.log, err, "send channel message failed")
	}

	msg, err := s.repo.GetMessageByID(ctx, created.ID)
	if err != nil {
		return sqlc.GetMessageByIDRow{}, logging.Unexpected(s.log, err, "send channel message: reload failed")
	}

	s.log.Info().
		Str("skill_slug", skillSlug).
		Str("channel_slug", channelSlug).
		Str("author", user.Username).
		Int64("message_id", msg.ID).
		Msg("channel message created")

	s.publish(RoomKeyChannel(skillSlug, channelSlug), RealtimeEvent{
		Type:    EventMessageCreated,
		Message: ptrMessage(toGetMessageResponse(msg)),
	})

	return msg, nil
}

func (s *Service) ListPodMessages(
	ctx context.Context,
	userPublicID uuid.UUID,
	podSlug string,
	input ListMessagesInput,
) ([]sqlc.ListPodMessagesRow, error) {
	if _, _, err := s.requireAcceptedPodMember(ctx, userPublicID, podSlug); err != nil {
		return nil, err
	}

	podRow, err := s.repo.GetPodBySlug(ctx, podSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPodNotFound
		}
		return nil, logging.Unexpected(s.log, err, "list pod messages: get pod failed")
	}

	rows, err := s.repo.ListPodMessages(ctx, podRow.ID, input.BeforeID, normalizeLimit(input.Limit))
	if err != nil {
		return nil, logging.Unexpected(s.log, err, "list pod messages failed")
	}
	return reversePodMessages(rows), nil
}

func (s *Service) SendPodMessage(
	ctx context.Context,
	userPublicID uuid.UUID,
	podSlug string,
	input SendMessageInput,
) (sqlc.GetMessageByIDRow, error) {
	content, err := normalizeContent(input.Content)
	if err != nil {
		return sqlc.GetMessageByIDRow{}, logging.Reject(s.log, err, "send pod message rejected: invalid content")
	}

	user, podRow, err := s.requireAcceptedPodMember(ctx, userPublicID, podSlug)
	if err != nil {
		return sqlc.GetMessageByIDRow{}, s.rejectIfDomain(err, "send pod message rejected")
	}

	if err := s.validateReplyTarget(ctx, input.ReplyToMessageID, 0, podRow.ID); err != nil {
		return sqlc.GetMessageByIDRow{}, err
	}

	created, err := s.repo.CreatePodMessage(ctx, user.ID, podRow.ID, content, input.ReplyToMessageID)
	if err != nil {
		return sqlc.GetMessageByIDRow{}, logging.Unexpected(s.log, err, "send pod message failed")
	}

	msg, err := s.repo.GetMessageByID(ctx, created.ID)
	if err != nil {
		return sqlc.GetMessageByIDRow{}, logging.Unexpected(s.log, err, "send pod message: reload failed")
	}

	s.log.Info().
		Str("pod_slug", podSlug).
		Str("author", user.Username).
		Int64("message_id", msg.ID).
		Msg("pod message created")

	s.publish(RoomKeyPod(podSlug), RealtimeEvent{
		Type:    EventMessageCreated,
		Message: ptrMessage(toGetMessageResponse(msg)),
	})

	return msg, nil
}

func (s *Service) EditMessage(
	ctx context.Context,
	userPublicID uuid.UUID,
	messageID int64,
	input SendMessageInput,
) (sqlc.GetMessageByIDRow, error) {
	content, err := normalizeContent(input.Content)
	if err != nil {
		return sqlc.GetMessageByIDRow{}, logging.Reject(s.log, err, "edit message rejected: invalid content")
	}

	user, err := s.repo.GetUserByPublicID(ctx, userPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.GetMessageByIDRow{}, logging.Reject(s.log, apperrors.ErrNotFound, "edit message rejected: user not found")
		}
		return sqlc.GetMessageByIDRow{}, logging.Unexpected(s.log, err, "edit message: get user failed")
	}

	existing, err := s.repo.GetMessageByID(ctx, messageID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.GetMessageByIDRow{}, logging.Reject(s.log, ErrMessageNotFound, "edit message rejected: message not found")
		}
		return sqlc.GetMessageByIDRow{}, logging.Unexpected(s.log, err, "edit message: get message failed")
	}
	if existing.AuthorID != user.ID {
		return sqlc.GetMessageByIDRow{}, logging.Reject(s.log, ErrNotMessageAuthor, "edit message rejected: not author")
	}

	if err := s.ensureCanAccessMessage(ctx, user, existing); err != nil {
		return sqlc.GetMessageByIDRow{}, err
	}

	if _, err := s.repo.UpdateMessageContent(ctx, messageID, user.ID, content); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.GetMessageByIDRow{}, logging.Reject(s.log, ErrNotMessageAuthor, "edit message rejected: not author")
		}
		return sqlc.GetMessageByIDRow{}, logging.Unexpected(s.log, err, "edit message failed")
	}

	msg, err := s.repo.GetMessageByID(ctx, messageID)
	if err != nil {
		return sqlc.GetMessageByIDRow{}, logging.Unexpected(s.log, err, "edit message: reload failed")
	}

	s.log.Info().Int64("message_id", messageID).Str("author", user.Username).Msg("message edited")
	if room := roomKeyFromMessage(msg); room != "" {
		s.publish(room, RealtimeEvent{
			Type:    EventMessageUpdated,
			Message: ptrMessage(toGetMessageResponse(msg)),
		})
	}
	return msg, nil
}

func (s *Service) DeleteMessage(ctx context.Context, userPublicID uuid.UUID, messageID int64) error {
	user, err := s.repo.GetUserByPublicID(ctx, userPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return logging.Reject(s.log, apperrors.ErrNotFound, "delete message rejected: user not found")
		}
		return logging.Unexpected(s.log, err, "delete message: get user failed")
	}

	existing, err := s.repo.GetMessageByID(ctx, messageID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return logging.Reject(s.log, ErrMessageNotFound, "delete message rejected: message not found")
		}
		return logging.Unexpected(s.log, err, "delete message: get message failed")
	}
	if existing.AuthorID != user.ID {
		return logging.Reject(s.log, ErrNotMessageAuthor, "delete message rejected: not author")
	}

	if err := s.ensureCanAccessMessage(ctx, user, existing); err != nil {
		return err
	}

	rows, err := s.repo.DeleteMessage(ctx, messageID, user.ID)
	if err != nil {
		return logging.Unexpected(s.log, err, "delete message failed")
	}
	if rows == 0 {
		return logging.Reject(s.log, ErrNotMessageAuthor, "delete message rejected: not author")
	}

	s.log.Info().Int64("message_id", messageID).Str("author", user.Username).Msg("message deleted")
	if room := roomKeyFromMessage(existing); room != "" {
		id := messageID
		s.publish(room, RealtimeEvent{
			Type:      EventMessageDeleted,
			MessageID: &id,
		})
	}
	return nil
}

func (s *Service) requireEnrolled(ctx context.Context, userPublicID uuid.UUID, skillSlug string) (sqlc.User, error) {
	user, err := s.repo.GetUserByPublicID(ctx, userPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.User{}, apperrors.ErrNotFound
		}
		return sqlc.User{}, logging.Unexpected(s.log, err, "community access: get user failed")
	}

	if _, err := s.repo.GetUserSkillByUserAndSkillSlug(ctx, user.ID, skillSlug); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.User{}, ErrNotEnrolledInSkill
		}
		return sqlc.User{}, logging.Unexpected(s.log, err, "community access: check enrollment failed")
	}
	return user, nil
}

func (s *Service) requireAcceptedPodMember(
	ctx context.Context,
	userPublicID uuid.UUID,
	podSlug string,
) (sqlc.User, sqlc.GetPodBySlugRow, error) {
	user, err := s.repo.GetUserByPublicID(ctx, userPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.User{}, sqlc.GetPodBySlugRow{}, apperrors.ErrNotFound
		}
		return sqlc.User{}, sqlc.GetPodBySlugRow{}, logging.Unexpected(s.log, err, "pod chat access: get user failed")
	}

	podRow, err := s.repo.GetPodBySlug(ctx, podSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.User{}, sqlc.GetPodBySlugRow{}, ErrPodNotFound
		}
		return sqlc.User{}, sqlc.GetPodBySlugRow{}, logging.Unexpected(s.log, err, "pod chat access: get pod failed")
	}

	membership, err := s.repo.GetPodMembershipByPodAndUser(ctx, podRow.ID, user.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.User{}, sqlc.GetPodBySlugRow{}, ErrNotAcceptedPodMember
		}
		return sqlc.User{}, sqlc.GetPodBySlugRow{}, logging.Unexpected(s.log, err, "pod chat access: get membership failed")
	}
	if membership.Status != sqlc.UserPodMembershipStatusACCEPTED {
		return sqlc.User{}, sqlc.GetPodBySlugRow{}, ErrNotAcceptedPodMember
	}

	return user, podRow, nil
}

func (s *Service) rejectIfDomain(err error, msg string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, apperrors.ErrNotFound) ||
		errors.Is(err, ErrNotEnrolledInSkill) ||
		errors.Is(err, ErrNotAcceptedPodMember) ||
		errors.Is(err, ErrPodNotFound) ||
		errors.Is(err, ErrChannelNotFound) ||
		errors.Is(err, ErrMessageNotFound) ||
		errors.Is(err, ErrNotMessageAuthor) ||
		errors.Is(err, ErrInvalidMessageContent) ||
		errors.Is(err, ErrInvalidReplyTarget) ||
		errors.Is(err, ErrInvalidChatRoom) ||
		errors.Is(err, ErrCommunityNotFound) {
		return logging.Reject(s.log, err, msg)
	}
	return err
}

func (s *Service) AuthorizeRealtimeRoom(ctx context.Context, userPublicID uuid.UUID, room string) error {
	kind, a, b, err := parseRoomKey(room)
	if err != nil {
		return logging.Reject(s.log, ErrInvalidChatRoom, "chat ws rejected: invalid room")
	}
	switch kind {
	case "pod":
		_, _, err := s.requireAcceptedPodMember(ctx, userPublicID, a)
		return s.rejectIfDomain(err, "chat ws rejected")
	case "channel":
		_, err := s.requireEnrolled(ctx, userPublicID, a)
		if err != nil {
			return s.rejectIfDomain(err, "chat ws rejected")
		}
		if _, err := s.repo.GetChannelBySkillSlugAndChannelSlug(ctx, a, b); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return logging.Reject(s.log, ErrChannelNotFound, "chat ws rejected: channel not found")
			}
			return logging.Unexpected(s.log, err, "chat ws: get channel failed")
		}
		return nil
	default:
		return logging.Reject(s.log, ErrInvalidChatRoom, "chat ws rejected: invalid room")
	}
}

func (s *Service) validateReplyTarget(ctx context.Context, replyTo *int64, channelID, podID int64) error {
	if replyTo == nil {
		return nil
	}
	parent, err := s.repo.GetMessageByID(ctx, *replyTo)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return logging.Reject(s.log, ErrInvalidReplyTarget, "send message rejected: reply target not found")
		}
		return logging.Unexpected(s.log, err, "send message: get reply target failed")
	}
	if channelID > 0 {
		if !parent.CommunityChannelID.Valid || parent.CommunityChannelID.Int64 != channelID {
			return logging.Reject(s.log, ErrInvalidReplyTarget, "send message rejected: reply target not in channel")
		}
		return nil
	}
	if podID > 0 {
		if !parent.PodID.Valid || parent.PodID.Int64 != podID {
			return logging.Reject(s.log, ErrInvalidReplyTarget, "send message rejected: reply target not in pod")
		}
		return nil
	}
	return logging.Reject(s.log, ErrInvalidReplyTarget, "send message rejected: reply target invalid")
}

func (s *Service) publish(room string, event RealtimeEvent) {
	if s.hub == nil || room == "" {
		return
	}
	s.hub.Publish(room, event)
}

func parseRoomKey(room string) (kind, a, b string, err error) {
	parts := strings.Split(room, ":")
	if len(parts) < 2 {
		return "", "", "", ErrInvalidChatRoom
	}
	switch parts[0] {
	case "pod":
		if len(parts) != 2 || parts[1] == "" {
			return "", "", "", ErrInvalidChatRoom
		}
		return "pod", parts[1], "", nil
	case "channel":
		if len(parts) != 3 || parts[1] == "" || parts[2] == "" {
			return "", "", "", ErrInvalidChatRoom
		}
		return "channel", parts[1], parts[2], nil
	default:
		return "", "", "", ErrInvalidChatRoom
	}
}

func roomKeyFromMessage(msg sqlc.GetMessageByIDRow) string {
	if msg.PodID.Valid && msg.PodSlug.Valid {
		return RoomKeyPod(msg.PodSlug.String)
	}
	if msg.CommunityChannelID.Valid && msg.SkillSlug.Valid && msg.ChannelSlug.Valid {
		return RoomKeyChannel(msg.SkillSlug.String, msg.ChannelSlug.String)
	}
	return ""
}

func ptrMessage(m messageResponse) *messageResponse {
	return &m
}

func (s *Service) ensureCanAccessMessage(ctx context.Context, user sqlc.User, msg sqlc.GetMessageByIDRow) error {
	if msg.CommunityChannelID.Valid {
		skillSlug := ""
		if msg.SkillSlug.Valid {
			skillSlug = msg.SkillSlug.String
		}
		if skillSlug == "" {
			return logging.Reject(s.log, ErrChannelNotFound, "message access rejected: channel not found")
		}
		if _, err := s.repo.GetUserSkillByUserAndSkillSlug(ctx, user.ID, skillSlug); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return logging.Reject(s.log, ErrNotEnrolledInSkill, "message access rejected: not enrolled")
			}
			return logging.Unexpected(s.log, err, "message access: check enrollment failed")
		}
		return nil
	}

	if msg.PodID.Valid {
		podSlug := ""
		if msg.PodSlug.Valid {
			podSlug = msg.PodSlug.String
		}
		if podSlug == "" {
			return logging.Reject(s.log, ErrPodNotFound, "message access rejected: pod not found")
		}
		podRow, err := s.repo.GetPodBySlug(ctx, podSlug)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return logging.Reject(s.log, ErrPodNotFound, "message access rejected: pod not found")
			}
			return logging.Unexpected(s.log, err, "message access: get pod failed")
		}
		membership, err := s.repo.GetPodMembershipByPodAndUser(ctx, podRow.ID, user.ID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return logging.Reject(s.log, ErrNotAcceptedPodMember, "message access rejected: not accepted member")
			}
			return logging.Unexpected(s.log, err, "message access: get membership failed")
		}
		if membership.Status != sqlc.UserPodMembershipStatusACCEPTED {
			return logging.Reject(s.log, ErrNotAcceptedPodMember, "message access rejected: not accepted member")
		}
		return nil
	}

	return logging.Reject(s.log, ErrMessageNotFound, "message access rejected: message not found")
}

func normalizeContent(raw string) (string, error) {
	content := strings.TrimSpace(raw)
	if content == "" || len(content) > maxMessageContent {
		return "", ErrInvalidMessageContent
	}
	return content, nil
}

func normalizeLimit(limit int32) int32 {
	if limit <= 0 {
		return defaultMessageLimit
	}
	if limit > maxMessageLimit {
		return maxMessageLimit
	}
	return limit
}

func reverseChannelMessages(rows []sqlc.ListChannelMessagesRow) []sqlc.ListChannelMessagesRow {
	for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
		rows[i], rows[j] = rows[j], rows[i]
	}
	return rows
}

func reversePodMessages(rows []sqlc.ListPodMessagesRow) []sqlc.ListPodMessagesRow {
	for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
		rows[i], rows[j] = rows[j], rows[i]
	}
	return rows
}
