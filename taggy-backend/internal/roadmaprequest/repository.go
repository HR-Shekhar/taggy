package roadmaprequest

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
	return &Repository{pool: pool, queries: sqlc.New(pool)}
}

func (r *Repository) WithTx(ctx context.Context, fn func(q *sqlc.Queries) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := fn(r.queries.WithTx(tx)); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repository) GetUserByPublicID(ctx context.Context, id uuid.UUID) (sqlc.User, error) {
	return r.queries.GetUserByPublicID(ctx, id)
}

func (r *Repository) GetUserByUsername(ctx context.Context, username string) (sqlc.User, error) {
	return r.queries.GetUserByUsername(ctx, username)
}

func (r *Repository) GetSkillBySlug(ctx context.Context, slug string) (sqlc.Skill, error) {
	return r.queries.GetSkillBySlug(ctx, slug)
}

func (r *Repository) GetUserSkill(ctx context.Context, userID int64, skillSlug string) (sqlc.Userskill, error) {
	return r.queries.GetUserSkillByUserAndSkillSlug(ctx, sqlc.GetUserSkillByUserAndSkillSlugParams{
		UserID: userID,
		Slug:   skillSlug,
	})
}

func (r *Repository) GetActiveVersion(ctx context.Context, skillID int64) (sqlc.RoadmapVersion, error) {
	return r.queries.GetActiveCatalogRoadmapVersionBySkillID(ctx, skillID)
}

func (r *Repository) ListMilestonesByVersion(ctx context.Context, versionID int64) ([]sqlc.ListMilestonesByRoadmapVersionIDRow, error) {
	return r.queries.ListMilestonesByRoadmapVersionID(ctx, pgtype.Int8{Int64: versionID, Valid: true})
}

func (r *Repository) GetMaxVersion(ctx context.Context, roadmapID int64) (int32, error) {
	return r.queries.GetMaxRoadmapVersionNumber(ctx, pgtype.Int8{Int64: roadmapID, Valid: true})
}

func (r *Repository) ListEnrolledUserIDs(ctx context.Context, skillID int64) ([]int64, error) {
	return r.queries.ListUserIDsEnrolledInSkill(ctx, skillID)
}

func (r *Repository) Create(ctx context.Context, arg sqlc.CreateRoadmapEditRequestParams) (sqlc.RoadmapEditRequest, error) {
	return r.queries.CreateRoadmapEditRequest(ctx, arg)
}

func (r *Repository) GetByID(ctx context.Context, id int64) (sqlc.GetRoadmapEditRequestByIDRow, error) {
	return r.queries.GetRoadmapEditRequestByID(ctx, id)
}

func (r *Repository) GetByPublicID(ctx context.Context, id uuid.UUID) (sqlc.GetRoadmapEditRequestByPublicIDRow, error) {
	return r.queries.GetRoadmapEditRequestByPublicID(ctx, id)
}

func (r *Repository) ListGeneratingIDs(ctx context.Context) ([]int64, error) {
	return r.queries.ListGeneratingRoadmapEditRequests(ctx)
}

func (r *Repository) CompleteDraft(ctx context.Context, id int64, draftJSON []byte) (sqlc.RoadmapEditRequest, error) {
	return r.queries.CompleteRoadmapEditDraft(ctx, sqlc.CompleteRoadmapEditDraftParams{
		ID:              id,
		DraftMilestones: draftJSON,
	})
}

func (r *Repository) FailGenerating(ctx context.Context, id int64, note string) (sqlc.RoadmapEditRequest, error) {
	return r.queries.FailRoadmapEditRequest(ctx, sqlc.FailRoadmapEditRequestParams{
		ID:        id,
		AdminNote: pgtype.Text{String: note, Valid: note != ""},
	})
}

func (r *Repository) AutoRejectGenerating(ctx context.Context, id int64, note string) (sqlc.RoadmapEditRequest, error) {
	return r.queries.AutoRejectGeneratingRoadmapEditRequest(ctx, sqlc.AutoRejectGeneratingRoadmapEditRequestParams{
		ID:        id,
		AdminNote: pgtype.Text{String: note, Valid: note != ""},
	})
}

func (r *Repository) ListByRequester(ctx context.Context, requesterID int64, limit int32) ([]sqlc.ListRoadmapEditRequestsByRequesterRow, error) {
	return r.queries.ListRoadmapEditRequestsByRequester(ctx, sqlc.ListRoadmapEditRequestsByRequesterParams{
		RequesterID: requesterID,
		ResultLimit: limit,
	})
}

func (r *Repository) ListPending(ctx context.Context, limit int32) ([]sqlc.ListPendingRoadmapEditRequestsRow, error) {
	return r.queries.ListPendingRoadmapEditRequests(ctx, limit)
}

func (r *Repository) Cancel(ctx context.Context, publicID uuid.UUID, requesterID int64) (sqlc.RoadmapEditRequest, error) {
	return r.queries.CancelRoadmapEditRequest(ctx, sqlc.CancelRoadmapEditRequestParams{
		PublicID:    publicID,
		RequesterID: requesterID,
	})
}

func (r *Repository) GetPending(ctx context.Context, requesterID, skillID int64) (sqlc.RoadmapEditRequest, error) {
	return r.queries.GetPendingRoadmapEditByRequesterAndSkill(ctx, sqlc.GetPendingRoadmapEditByRequesterAndSkillParams{
		RequesterID: requesterID,
		SkillID:     skillID,
	})
}

func (r *Repository) Reject(ctx context.Context, id, reviewerID int64, note *string) (sqlc.RoadmapEditRequest, error) {
	noteText := pgtype.Text{}
	if note != nil && *note != "" {
		noteText = pgtype.Text{String: *note, Valid: true}
	}
	return r.queries.RejectRoadmapEditRequest(ctx, sqlc.RejectRoadmapEditRequestParams{
		ID:         id,
		ReviewedBy: pgtype.Int8{Int64: reviewerID, Valid: true},
		AdminNote:  noteText,
	})
}
