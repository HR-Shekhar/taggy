package community

import (
	"context"

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

func (r *Repository) GetCommunityBySkillSlug(ctx context.Context, skillSlug string) (sqlc.GetCommunityBySkillSlugRow, error) {
	return r.queries.GetCommunityBySkillSlug(ctx, skillSlug)
}

func (r *Repository) GetUserSkillByUserAndSkillSlug(ctx context.Context, userID int64, skillSlug string) (sqlc.Userskill, error) {
	return r.queries.GetUserSkillByUserAndSkillSlug(ctx, sqlc.GetUserSkillByUserAndSkillSlugParams{
		UserID: userID,
		Slug:   skillSlug,
	})
}

func (r *Repository) ListChannelsByCommunityID(ctx context.Context, communityID int64) ([]sqlc.ListChannelsByCommunityIDRow, error) {
	return r.queries.ListChannelsByCommunityID(ctx, communityID)
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

func (r *Repository) GetPodBySlug(ctx context.Context, slug string) (sqlc.GetPodBySlugRow, error) {
	return r.queries.GetPodBySlug(ctx, slug)
}

func (r *Repository) GetPodMembershipByPodAndUser(ctx context.Context, podID, userID int64) (sqlc.PodMembership, error) {
	return r.queries.GetPodMembershipByPodAndUser(ctx, sqlc.GetPodMembershipByPodAndUserParams{
		PodID:  podID,
		UserID: userID,
	})
}

func (r *Repository) CreateChannelMessage(
	ctx context.Context,
	authorID, channelID int64,
	content string,
	replyToMessageID *int64,
) (sqlc.Message, error) {
	return r.queries.CreateChannelMessage(ctx, sqlc.CreateChannelMessageParams{
		AuthorID:           authorID,
		CommunityChannelID: pgtype.Int8{Int64: channelID, Valid: true},
		Content:            content,
		ReplyToMessageID:   optionalInt8(replyToMessageID),
	})
}

func (r *Repository) CreatePodMessage(
	ctx context.Context,
	authorID, podID int64,
	content string,
	replyToMessageID *int64,
) (sqlc.Message, error) {
	return r.queries.CreatePodMessage(ctx, sqlc.CreatePodMessageParams{
		AuthorID:         authorID,
		PodID:            pgtype.Int8{Int64: podID, Valid: true},
		Content:          content,
		ReplyToMessageID: optionalInt8(replyToMessageID),
	})
}

func (r *Repository) ListChannelMessages(
	ctx context.Context,
	channelID, beforeID int64,
	limit int32,
) ([]sqlc.ListChannelMessagesRow, error) {
	return r.queries.ListChannelMessages(ctx, sqlc.ListChannelMessagesParams{
		ChannelID:    pgtype.Int8{Int64: channelID, Valid: true},
		BeforeID:     beforeID,
		MessageLimit: limit,
	})
}

func (r *Repository) ListPodMessages(
	ctx context.Context,
	podID, beforeID int64,
	limit int32,
) ([]sqlc.ListPodMessagesRow, error) {
	return r.queries.ListPodMessages(ctx, sqlc.ListPodMessagesParams{
		PodID:        pgtype.Int8{Int64: podID, Valid: true},
		BeforeID:     beforeID,
		MessageLimit: limit,
	})
}

func (r *Repository) GetMessageByID(ctx context.Context, id int64) (sqlc.GetMessageByIDRow, error) {
	return r.queries.GetMessageByID(ctx, id)
}

func (r *Repository) UpdateMessageContent(ctx context.Context, id, authorID int64, content string) (sqlc.Message, error) {
	return r.queries.UpdateMessageContent(ctx, sqlc.UpdateMessageContentParams{
		ID:       id,
		Content:  content,
		AuthorID: authorID,
	})
}

func (r *Repository) DeleteMessage(ctx context.Context, id, authorID int64) (int64, error) {
	return r.queries.DeleteMessage(ctx, sqlc.DeleteMessageParams{
		ID:       id,
		AuthorID: authorID,
	})
}

func optionalInt8(v *int64) pgtype.Int8 {
	if v == nil {
		return pgtype.Int8{}
	}
	return pgtype.Int8{Int64: *v, Valid: true}
}
