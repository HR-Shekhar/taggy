package audio

import (
	"context"
	"fmt"
	"time"

	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

func (r *Repository) GetUserByPublicID(ctx context.Context, publicID uuid.UUID) (sqlc.User, error) {
	return r.queries.GetUserByPublicID(ctx, publicID)
}

func (r *Repository) GetPodBySlug(ctx context.Context, slug string) (sqlc.GetPodBySlugRow, error) {
	return r.queries.GetPodBySlug(ctx, slug)
}

func (r *Repository) GetPodMembershipByPodAndUser(ctx context.Context, podID, userID int64) (sqlc.PodMembership, error) {
	return r.queries.GetPodMembershipByPodAndUser(ctx, sqlc.GetPodMembershipByPodAndUserParams{
		PodID:  podID,
		UserID: userID,
	})
}

func (r *Repository) GetUserSkillByUserAndSkillSlug(ctx context.Context, userID int64, skillSlug string) (sqlc.Userskill, error) {
	return r.queries.GetUserSkillByUserAndSkillSlug(ctx, sqlc.GetUserSkillByUserAndSkillSlugParams{
		UserID: userID,
		Slug:   skillSlug,
	})
}

func (r *Repository) GetChannelBySkillSlugAndChannelSlug(
	ctx context.Context,
	skillSlug, channelSlug string,
) (sqlc.GetChannelBySkillSlugAndChannelSlugRow, error) {
	return r.queries.GetChannelBySkillSlugAndChannelSlug(ctx, sqlc.GetChannelBySkillSlugAndChannelSlugParams{
		SkillSlug:   skillSlug,
		ChannelSlug: channelSlug,
	})
}

func (r *Repository) GetActiveAudioRoomByPodID(ctx context.Context, podID int64) (sqlc.AudioRoom, error) {
	return r.queries.GetActiveAudioRoomByPodID(ctx, pgtype.Int8{Int64: podID, Valid: true})
}

func (r *Repository) GetActiveAudioRoomByChannelID(ctx context.Context, channelID int64) (sqlc.AudioRoom, error) {
	return r.queries.GetActiveAudioRoomByChannelID(ctx, pgtype.Int8{Int64: channelID, Valid: true})
}

func (r *Repository) GetAudioRoomByPublicID(ctx context.Context, publicID uuid.UUID) (sqlc.GetAudioRoomByPublicIDRow, error) {
	return r.queries.GetAudioRoomByPublicID(ctx, publicID)
}

func (r *Repository) ListAudioRoomsByPodID(
	ctx context.Context,
	podID int64,
	status sqlc.NullAudioRoomStatus,
) ([]sqlc.ListAudioRoomsByPodIDRow, error) {
	return r.queries.ListAudioRoomsByPodID(ctx, sqlc.ListAudioRoomsByPodIDParams{
		PodID:  pgtype.Int8{Int64: podID, Valid: true},
		Status: status,
	})
}

func (r *Repository) ListAudioRoomsByChannelID(
	ctx context.Context,
	channelID int64,
	status sqlc.NullAudioRoomStatus,
) ([]sqlc.ListAudioRoomsByChannelIDRow, error) {
	return r.queries.ListAudioRoomsByChannelID(ctx, sqlc.ListAudioRoomsByChannelIDParams{
		ChannelID: pgtype.Int8{Int64: channelID, Valid: true},
		Status:    status,
	})
}

func (r *Repository) CountActiveParticipants(ctx context.Context, roomID int64) (int64, error) {
	return r.queries.CountActiveParticipants(ctx, roomID)
}

func (r *Repository) ListActiveParticipants(ctx context.Context, roomID int64) ([]sqlc.ListActiveAudioRoomParticipantsRow, error) {
	return r.queries.ListActiveAudioRoomParticipants(ctx, roomID)
}

func (r *Repository) UpsertParticipant(
	ctx context.Context,
	roomID, userID int64,
	role sqlc.AudioRoomParticipantRole,
) (sqlc.AudioRoomParticipant, error) {
	return r.queries.UpsertAudioRoomParticipant(ctx, sqlc.UpsertAudioRoomParticipantParams{
		AudioRoomID: roomID,
		UserID:      userID,
		JoinedAt:    pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		Role:        role,
	})
}

func (r *Repository) LeaveParticipant(ctx context.Context, roomID, userID int64) (sqlc.AudioRoomParticipant, error) {
	return r.queries.LeaveAudioRoomParticipant(ctx, sqlc.LeaveAudioRoomParticipantParams{
		AudioRoomID: roomID,
		UserID:      userID,
	})
}

func (r *Repository) EndAudioRoom(ctx context.Context, roomID int64) (sqlc.AudioRoom, error) {
	return r.queries.EndAudioRoom(ctx, roomID)
}

// EndAllActiveRooms marks every ACTIVE room ENDED and sets left_at on open participants.
func (r *Repository) EndAllActiveRooms(ctx context.Context) (roomsEnded, participantsLeft int64, err error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, 0, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	participantsLeft, err = qtx.LeaveAllActiveAudioParticipants(ctx)
	if err != nil {
		return 0, 0, err
	}

	roomsEnded, err = qtx.EndAllActiveAudioRooms(ctx)
	if err != nil {
		return 0, 0, err
	}

	if err := tx.Commit(ctx); err != nil {
		return 0, 0, fmt.Errorf("commit transaction: %w", err)
	}
	return roomsEnded, participantsLeft, nil
}

func (r *Repository) ListStaleEmptyActiveAudioRooms(
	ctx context.Context,
	emptyBefore pgtype.Timestamptz,
) ([]sqlc.ListStaleEmptyActiveAudioRoomsRow, error) {
	return r.queries.ListStaleEmptyActiveAudioRooms(ctx, emptyBefore)
}

func (r *Repository) ListActiveAudioRooms(ctx context.Context) ([]sqlc.ListActiveAudioRoomsRow, error) {
	return r.queries.ListActiveAudioRooms(ctx)
}

func (r *Repository) DeleteAudioRoom(ctx context.Context, roomID int64) error {
	return r.queries.DeleteAudioRoom(ctx, roomID)
}

// CreatePodRoomWithHost creates an ACTIVE pod room and inserts the host participant.
func (r *Repository) CreatePodRoomWithHost(
	ctx context.Context,
	hostID, podID int64,
	publicID uuid.UUID,
	title string,
	description pgtype.Text,
	maxParticipants pgtype.Int4,
	livekitRoomName string,
) (sqlc.AudioRoom, error) {
	return r.createRoomWithHost(ctx, sqlc.CreateAudioRoomParams{
		PublicID:           publicID,
		RoomType:           sqlc.AudioRoomTypePOD,
		PodID:              pgtype.Int8{Int64: podID, Valid: true},
		CommunityChannelID: pgtype.Int8{},
		HostID:             hostID,
		Title:              title,
		Description:        description,
		LivekitRoomName:    livekitRoomName,
		Status:             sqlc.AudioRoomStatusACTIVE,
		ActualStartTime:    pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		MaxParticipants:    maxParticipants,
	})
}

func (r *Repository) CreateChannelRoomWithHost(
	ctx context.Context,
	hostID, channelID int64,
	publicID uuid.UUID,
	title string,
	description pgtype.Text,
	maxParticipants pgtype.Int4,
	livekitRoomName string,
) (sqlc.AudioRoom, error) {
	return r.createRoomWithHost(ctx, sqlc.CreateAudioRoomParams{
		PublicID:           publicID,
		RoomType:           sqlc.AudioRoomTypeCOMMUNITYCHANNEL,
		PodID:              pgtype.Int8{},
		CommunityChannelID: pgtype.Int8{Int64: channelID, Valid: true},
		HostID:             hostID,
		Title:              title,
		Description:        description,
		LivekitRoomName:    livekitRoomName,
		Status:             sqlc.AudioRoomStatusACTIVE,
		ActualStartTime:    pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		MaxParticipants:    maxParticipants,
	})
}

func (r *Repository) createRoomWithHost(ctx context.Context, params sqlc.CreateAudioRoomParams) (sqlc.AudioRoom, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return sqlc.AudioRoom{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	room, err := qtx.CreateAudioRoom(ctx, params)
	if err != nil {
		return sqlc.AudioRoom{}, err
	}

	_, err = qtx.UpsertAudioRoomParticipant(ctx, sqlc.UpsertAudioRoomParticipantParams{
		AudioRoomID: room.ID,
		UserID:      params.HostID,
		JoinedAt:    pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		Role:        sqlc.AudioRoomParticipantRoleHOST,
	})
	if err != nil {
		return sqlc.AudioRoom{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return sqlc.AudioRoom{}, fmt.Errorf("commit transaction: %w", err)
	}
	return room, nil
}
