package audio

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/livekit"
	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/postgres/sqlc"
	apperrors "github.com/HR-Shekhar/taggy-backend/internal/shared/errors"
	"github.com/HR-Shekhar/taggy-backend/internal/shared/logging"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
)

type Service struct {
	repo    *Repository
	livekit *livekit.TokenClient
	log     zerolog.Logger
}

func NewService(repo *Repository, livekitClient *livekit.TokenClient, log zerolog.Logger) *Service {
	return &Service{repo: repo, livekit: livekitClient, log: log}
}

func (s *Service) CreatePodRoom(
	ctx context.Context,
	userPublicID uuid.UUID,
	podSlug string,
	input CreateRoomInput,
) (sqlc.GetAudioRoomByPublicIDRow, error) {
	title, desc, maxParticipants, err := normalizeCreateInput(input)
	if err != nil {
		return sqlc.GetAudioRoomByPublicIDRow{}, logging.Reject(s.log, err, "create pod room rejected: invalid input")
	}

	user, podRow, err := s.requireAcceptedPodMember(ctx, userPublicID, podSlug)
	if err != nil {
		return sqlc.GetAudioRoomByPublicIDRow{}, err
	}

	if _, err := s.repo.GetActiveAudioRoomByPodID(ctx, podRow.ID); err == nil {
		return sqlc.GetAudioRoomByPublicIDRow{}, logging.Reject(s.log, ErrActiveRoomExists, "create pod room rejected: active room exists")
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return sqlc.GetAudioRoomByPublicIDRow{}, logging.Unexpected(s.log, err, "create pod room active check failed")
	}

	publicID := uuid.New()
	livekitName := fmt.Sprintf("taggy-pod-%s", publicID.String())

	room, err := s.repo.CreatePodRoomWithHost(ctx, user.ID, podRow.ID, publicID, title, desc, maxParticipants, livekitName)
	if err != nil {
		return sqlc.GetAudioRoomByPublicIDRow{}, logging.Unexpected(s.log, err, "create pod room failed")
	}

	detail, err := s.repo.GetAudioRoomByPublicID(ctx, room.PublicID)
	if err != nil {
		return sqlc.GetAudioRoomByPublicIDRow{}, logging.Unexpected(s.log, err, "create pod room reload failed")
	}

	s.log.Info().
		Str("pod_slug", podSlug).
		Str("room_id", room.PublicID.String()).
		Str("host", user.Username).
		Msg("pod audio room created")

	return detail, nil
}

func (s *Service) CreateChannelRoom(
	ctx context.Context,
	userPublicID uuid.UUID,
	skillSlug, channelSlug string,
	input CreateRoomInput,
) (sqlc.GetAudioRoomByPublicIDRow, error) {
	title, desc, maxParticipants, err := normalizeCreateInput(input)
	if err != nil {
		return sqlc.GetAudioRoomByPublicIDRow{}, logging.Reject(s.log, err, "create channel room rejected: invalid input")
	}

	user, err := s.requireEnrolled(ctx, userPublicID, skillSlug)
	if err != nil {
		return sqlc.GetAudioRoomByPublicIDRow{}, err
	}

	channel, err := s.repo.GetChannelBySkillSlugAndChannelSlug(ctx, skillSlug, channelSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.GetAudioRoomByPublicIDRow{}, logging.Reject(s.log, ErrChannelNotFound, "create channel room rejected: channel not found")
		}
		return sqlc.GetAudioRoomByPublicIDRow{}, logging.Unexpected(s.log, err, "create channel room channel lookup failed")
	}

	if _, err := s.repo.GetActiveAudioRoomByChannelID(ctx, channel.ID); err == nil {
		return sqlc.GetAudioRoomByPublicIDRow{}, logging.Reject(s.log, ErrActiveRoomExists, "create channel room rejected: active room exists")
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return sqlc.GetAudioRoomByPublicIDRow{}, logging.Unexpected(s.log, err, "create channel room active check failed")
	}

	publicID := uuid.New()
	livekitName := fmt.Sprintf("taggy-channel-%s", publicID.String())
	room, err := s.repo.CreateChannelRoomWithHost(ctx, user.ID, channel.ID, publicID, title, desc, maxParticipants, livekitName)
	if err != nil {
		return sqlc.GetAudioRoomByPublicIDRow{}, logging.Unexpected(s.log, err, "create channel room failed")
	}

	detail, err := s.repo.GetAudioRoomByPublicID(ctx, room.PublicID)
	if err != nil {
		return sqlc.GetAudioRoomByPublicIDRow{}, logging.Unexpected(s.log, err, "create channel room reload failed")
	}

	s.log.Info().
		Str("skill_slug", skillSlug).
		Str("channel_slug", channelSlug).
		Str("room_id", room.PublicID.String()).
		Str("host", user.Username).
		Msg("channel audio room created")

	return detail, nil
}

func (s *Service) ListPodRooms(
	ctx context.Context,
	userPublicID uuid.UUID,
	podSlug string,
	statusFilter string,
) ([]sqlc.ListAudioRoomsByPodIDRow, error) {
	if _, _, err := s.requireAcceptedPodMember(ctx, userPublicID, podSlug); err != nil {
		return nil, err
	}

	podRow, err := s.repo.GetPodBySlug(ctx, podSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, logging.Reject(s.log, ErrPodNotFound, "list pod rooms rejected: pod not found")
		}
		return nil, logging.Unexpected(s.log, err, "list pod rooms pod lookup failed")
	}

	status, err := parseOptionalStatus(statusFilter)
	if err != nil {
		return nil, logging.Reject(s.log, err, "list pod rooms rejected: invalid status")
	}

	rows, err := s.repo.ListAudioRoomsByPodID(ctx, podRow.ID, status)
	if err != nil {
		return nil, logging.Unexpected(s.log, err, "list pod rooms failed")
	}
	return rows, nil
}

func (s *Service) ListChannelRooms(
	ctx context.Context,
	userPublicID uuid.UUID,
	skillSlug, channelSlug, statusFilter string,
) ([]sqlc.ListAudioRoomsByChannelIDRow, error) {
	if _, err := s.requireEnrolled(ctx, userPublicID, skillSlug); err != nil {
		return nil, err
	}

	channel, err := s.repo.GetChannelBySkillSlugAndChannelSlug(ctx, skillSlug, channelSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, logging.Reject(s.log, ErrChannelNotFound, "list channel rooms rejected: channel not found")
		}
		return nil, logging.Unexpected(s.log, err, "list channel rooms channel lookup failed")
	}

	status, err := parseOptionalStatus(statusFilter)
	if err != nil {
		return nil, logging.Reject(s.log, err, "list channel rooms rejected: invalid status")
	}

	rows, err := s.repo.ListAudioRoomsByChannelID(ctx, channel.ID, status)
	if err != nil {
		return nil, logging.Unexpected(s.log, err, "list channel rooms failed")
	}
	return rows, nil
}

func (s *Service) GetRoom(
	ctx context.Context,
	userPublicID uuid.UUID,
	roomPublicID uuid.UUID,
) (sqlc.GetAudioRoomByPublicIDRow, []sqlc.ListActiveAudioRoomParticipantsRow, error) {
	room, err := s.repo.GetAudioRoomByPublicID(ctx, roomPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.GetAudioRoomByPublicIDRow{}, nil, logging.Reject(s.log, ErrRoomNotFound, "get room rejected: not found")
		}
		return sqlc.GetAudioRoomByPublicIDRow{}, nil, logging.Unexpected(s.log, err, "get room lookup failed")
	}

	user, err := s.repo.GetUserByPublicID(ctx, userPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.GetAudioRoomByPublicIDRow{}, nil, logging.Reject(s.log, apperrors.ErrNotFound, "get room rejected: user not found")
		}
		return sqlc.GetAudioRoomByPublicIDRow{}, nil, logging.Unexpected(s.log, err, "get room user lookup failed")
	}

	if err := s.ensureCanAccessRoom(ctx, user, room); err != nil {
		return sqlc.GetAudioRoomByPublicIDRow{}, nil, err
	}

	participants, err := s.repo.ListActiveParticipants(ctx, room.ID)
	if err != nil {
		return sqlc.GetAudioRoomByPublicIDRow{}, nil, logging.Unexpected(s.log, err, "get room list participants failed")
	}

	return room, participants, nil
}

func (s *Service) JoinRoom(
	ctx context.Context,
	userPublicID uuid.UUID,
	roomPublicID uuid.UUID,
) (JoinRoomResult, error) {
	if s.livekit == nil || !s.livekit.Configured() {
		return JoinRoomResult{}, logging.Reject(s.log, ErrLiveKitNotConfigured, "join room rejected: livekit not configured")
	}

	user, err := s.repo.GetUserByPublicID(ctx, userPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return JoinRoomResult{}, logging.Reject(s.log, apperrors.ErrNotFound, "join room rejected: user not found")
		}
		return JoinRoomResult{}, logging.Unexpected(s.log, err, "join room user lookup failed")
	}

	room, err := s.repo.GetAudioRoomByPublicID(ctx, roomPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return JoinRoomResult{}, logging.Reject(s.log, ErrRoomNotFound, "join room rejected: not found")
		}
		return JoinRoomResult{}, logging.Unexpected(s.log, err, "join room lookup failed")
	}
	if room.Status != sqlc.AudioRoomStatusACTIVE {
		return JoinRoomResult{}, logging.Reject(s.log, ErrRoomNotActive, "join room rejected: not active")
	}

	if err := s.ensureCanAccessRoom(ctx, user, room); err != nil {
		return JoinRoomResult{}, err
	}

	role := sqlc.AudioRoomParticipantRoleSPEAKER
	if room.HostID == user.ID {
		role = sqlc.AudioRoomParticipantRoleHOST
	} else if room.MaxParticipants.Valid {
		count, err := s.repo.CountActiveParticipants(ctx, room.ID)
		if err != nil {
			return JoinRoomResult{}, logging.Unexpected(s.log, err, "join room count participants failed")
		}
		participants, err := s.repo.ListActiveParticipants(ctx, room.ID)
		if err != nil {
			return JoinRoomResult{}, logging.Unexpected(s.log, err, "join room list participants failed")
		}
		alreadyIn := false
		for _, p := range participants {
			if p.UserID == user.ID {
				alreadyIn = true
				break
			}
		}
		if !alreadyIn && count >= int64(room.MaxParticipants.Int32) {
			return JoinRoomResult{}, logging.Reject(s.log, ErrRoomFull, "join room rejected: room full")
		}
	}

	participant, err := s.repo.UpsertParticipant(ctx, room.ID, user.ID, role)
	if err != nil {
		return JoinRoomResult{}, logging.Unexpected(s.log, err, "join room upsert participant failed")
	}

	token, err := s.livekit.MintJoinToken(livekit.MintParams{
		Identity:   user.PublicID.String(),
		Name:       user.Username,
		RoomName:   room.LivekitRoomName,
		CanPublish: true,
	})
	if err != nil {
		if errors.Is(err, livekit.ErrNotConfigured) {
			return JoinRoomResult{}, logging.Reject(s.log, ErrLiveKitNotConfigured, "join room rejected: livekit not configured")
		}
		return JoinRoomResult{}, logging.Unexpected(s.log, err, "join room mint token failed")
	}

	s.log.Info().
		Str("room_id", room.PublicID.String()).
		Str("user", user.Username).
		Str("role", string(participant.Role)).
		Msg("joined audio room")

	return JoinRoomResult{
		RoomID:          room.PublicID.String(),
		LiveKitURL:      s.livekit.URL(),
		LiveKitRoomName: room.LivekitRoomName,
		Token:           token,
		Role:            string(participant.Role),
	}, nil
}

func (s *Service) LeaveRoom(ctx context.Context, userPublicID uuid.UUID, roomPublicID uuid.UUID) error {
	user, err := s.repo.GetUserByPublicID(ctx, userPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return logging.Reject(s.log, apperrors.ErrNotFound, "leave room rejected: user not found")
		}
		return logging.Unexpected(s.log, err, "leave room user lookup failed")
	}

	room, err := s.repo.GetAudioRoomByPublicID(ctx, roomPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return logging.Reject(s.log, ErrRoomNotFound, "leave room rejected: not found")
		}
		return logging.Unexpected(s.log, err, "leave room lookup failed")
	}

	if err := s.ensureCanAccessRoom(ctx, user, room); err != nil {
		return err
	}

	if _, err := s.repo.LeaveParticipant(ctx, room.ID, user.ID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return logging.Reject(s.log, ErrNotParticipant, "leave room rejected: not participant")
		}
		return logging.Unexpected(s.log, err, "leave room failed")
	}

	s.log.Info().Str("room_id", room.PublicID.String()).Str("user", user.Username).Msg("left audio room")
	return nil
}

func (s *Service) EndRoom(ctx context.Context, userPublicID uuid.UUID, roomPublicID uuid.UUID) error {
	user, err := s.repo.GetUserByPublicID(ctx, userPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return logging.Reject(s.log, apperrors.ErrNotFound, "end room rejected: user not found")
		}
		return logging.Unexpected(s.log, err, "end room user lookup failed")
	}

	room, err := s.repo.GetAudioRoomByPublicID(ctx, roomPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return logging.Reject(s.log, ErrRoomNotFound, "end room rejected: not found")
		}
		return logging.Unexpected(s.log, err, "end room lookup failed")
	}
	if room.HostID != user.ID {
		return logging.Reject(s.log, ErrNotRoomHost, "end room rejected: not host")
	}
	if room.Status != sqlc.AudioRoomStatusACTIVE {
		return logging.Reject(s.log, ErrRoomNotActive, "end room rejected: not active")
	}

	livekitName := room.LivekitRoomName
	if err := s.repo.DeleteAudioRoom(ctx, room.ID); err != nil {
		return logging.Unexpected(s.log, err, "end room delete failed")
	}

	s.deleteLiveKitRoom(ctx, livekitName)

	s.log.Info().Str("room_id", room.PublicID.String()).Str("host", user.Username).Msg("audio room ended and deleted")
	return nil
}

// EndAllActiveRooms deletes every ACTIVE room during process shutdown (no host check)
// and best-effort deletes matching LiveKit rooms.
func (s *Service) EndAllActiveRooms(ctx context.Context) error {
	active, err := s.repo.ListActiveAudioRooms(ctx)
	if err != nil {
		return logging.Unexpected(s.log, err, "end all rooms list active failed")
	}

	for _, room := range active {
		s.deleteLiveKitRoom(ctx, room.LivekitRoomName)
		if err := s.repo.DeleteAudioRoom(ctx, room.ID); err != nil {
			s.log.Error().
				Err(err).
				Str("room_id", room.PublicID.String()).
				Msg("delete audio room on shutdown failed")
			continue
		}
	}

	s.log.Info().Int("rooms_deleted", len(active)).Msg("deleted all active audio rooms for shutdown")
	return nil
}

func (s *Service) deleteLiveKitRoom(ctx context.Context, livekitRoomName string) {
	if s.livekit == nil || !s.livekit.Configured() || strings.TrimSpace(livekitRoomName) == "" {
		return
	}
	if err := s.livekit.DeleteRoom(ctx, livekitRoomName); err != nil {
		s.log.Warn().
			Err(err).
			Str("livekit_room_name", livekitRoomName).
			Msg("livekit DeleteRoom failed (continuing)")
		return
	}
	s.log.Info().
		Str("livekit_room_name", livekitRoomName).
		Msg("livekit room deleted")
}

func (s *Service) requireEnrolled(ctx context.Context, userPublicID uuid.UUID, skillSlug string) (sqlc.User, error) {
	user, err := s.repo.GetUserByPublicID(ctx, userPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.User{}, logging.Reject(s.log, apperrors.ErrNotFound, "audio access rejected: user not found")
		}
		return sqlc.User{}, logging.Unexpected(s.log, err, "audio access user lookup failed")
	}
	if _, err := s.repo.GetUserSkillByUserAndSkillSlug(ctx, user.ID, skillSlug); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.User{}, logging.Reject(s.log, ErrNotEnrolledInSkill, "audio access rejected: not enrolled")
		}
		return sqlc.User{}, logging.Unexpected(s.log, err, "audio access enrollment lookup failed")
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
			return sqlc.User{}, sqlc.GetPodBySlugRow{}, logging.Reject(s.log, apperrors.ErrNotFound, "audio pod access rejected: user not found")
		}
		return sqlc.User{}, sqlc.GetPodBySlugRow{}, logging.Unexpected(s.log, err, "audio pod access user lookup failed")
	}

	podRow, err := s.repo.GetPodBySlug(ctx, podSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.User{}, sqlc.GetPodBySlugRow{}, logging.Reject(s.log, ErrPodNotFound, "audio pod access rejected: pod not found")
		}
		return sqlc.User{}, sqlc.GetPodBySlugRow{}, logging.Unexpected(s.log, err, "audio pod access pod lookup failed")
	}

	membership, err := s.repo.GetPodMembershipByPodAndUser(ctx, podRow.ID, user.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.User{}, sqlc.GetPodBySlugRow{}, logging.Reject(s.log, ErrNotAcceptedPodMember, "audio pod access rejected: not member")
		}
		return sqlc.User{}, sqlc.GetPodBySlugRow{}, logging.Unexpected(s.log, err, "audio pod access membership lookup failed")
	}
	if membership.Status != sqlc.UserPodMembershipStatusACCEPTED {
		return sqlc.User{}, sqlc.GetPodBySlugRow{}, logging.Reject(s.log, ErrNotAcceptedPodMember, "audio pod access rejected: not accepted")
	}

	return user, podRow, nil
}

func (s *Service) ensureCanAccessRoom(ctx context.Context, user sqlc.User, room sqlc.GetAudioRoomByPublicIDRow) error {
	switch room.RoomType {
	case sqlc.AudioRoomTypePOD:
		if !room.PodSlug.Valid {
			return logging.Reject(s.log, ErrPodNotFound, "audio room access rejected: pod not found")
		}
		podRow, err := s.repo.GetPodBySlug(ctx, room.PodSlug.String)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return logging.Reject(s.log, ErrPodNotFound, "audio room access rejected: pod not found")
			}
			return logging.Unexpected(s.log, err, "audio room access pod lookup failed")
		}
		membership, err := s.repo.GetPodMembershipByPodAndUser(ctx, podRow.ID, user.ID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return logging.Reject(s.log, ErrNotAcceptedPodMember, "audio room access rejected: not member")
			}
			return logging.Unexpected(s.log, err, "audio room access membership lookup failed")
		}
		if membership.Status != sqlc.UserPodMembershipStatusACCEPTED {
			return logging.Reject(s.log, ErrNotAcceptedPodMember, "audio room access rejected: not accepted")
		}
		return nil

	case sqlc.AudioRoomTypeCOMMUNITYCHANNEL:
		if !room.SkillSlug.Valid {
			return logging.Reject(s.log, ErrChannelNotFound, "audio room access rejected: channel not found")
		}
		if _, err := s.repo.GetUserSkillByUserAndSkillSlug(ctx, user.ID, room.SkillSlug.String); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return logging.Reject(s.log, ErrNotEnrolledInSkill, "audio room access rejected: not enrolled")
			}
			return logging.Unexpected(s.log, err, "audio room access enrollment lookup failed")
		}
		return nil

	default:
		return logging.Reject(s.log, ErrRoomNotFound, "audio room access rejected: unknown room type")
	}
}

func normalizeCreateInput(input CreateRoomInput) (string, pgtype.Text, pgtype.Int4, error) {
	title := strings.TrimSpace(input.Title)
	if len(title) < 3 || len(title) > 255 {
		return "", pgtype.Text{}, pgtype.Int4{}, ErrInvalidRoomTitle
	}

	desc := pgtype.Text{}
	if input.Description != nil && strings.TrimSpace(*input.Description) != "" {
		desc = pgtype.Text{String: strings.TrimSpace(*input.Description), Valid: true}
	}

	max := pgtype.Int4{}
	if input.MaxParticipants != nil {
		if *input.MaxParticipants < 1 {
			return "", pgtype.Text{}, pgtype.Int4{}, ErrInvalidRoomTitle
		}
		max = pgtype.Int4{Int32: *input.MaxParticipants, Valid: true}
	}

	return title, desc, max, nil
}

func parseOptionalStatus(raw string) (sqlc.NullAudioRoomStatus, error) {
	raw = strings.TrimSpace(strings.ToUpper(raw))
	if raw == "" {
		// Default list to live rooms only — ended rooms are deleted and shouldn't surface.
		return sqlc.NullAudioRoomStatus{
			AudioRoomStatus: sqlc.AudioRoomStatusACTIVE,
			Valid:           true,
		}, nil
	}
	switch sqlc.AudioRoomStatus(raw) {
	case sqlc.AudioRoomStatusACTIVE,
		sqlc.AudioRoomStatusENDED,
		sqlc.AudioRoomStatusSCHEDULED,
		sqlc.AudioRoomStatusCANCELLED:
		return sqlc.NullAudioRoomStatus{AudioRoomStatus: sqlc.AudioRoomStatus(raw), Valid: true}, nil
	default:
		return sqlc.NullAudioRoomStatus{}, apperrors.ErrBadRequest
	}
}
