package skillrequest

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

func (r *Repository) ListSimilarSkills(ctx context.Context, query string, minScore float32, limit int32) ([]sqlc.ListSimilarSkillsRow, error) {
	return r.queries.ListSimilarSkills(ctx, sqlc.ListSimilarSkillsParams{
		Query:       query,
		MinScore:    minScore,
		ResultLimit: limit,
	})
}

func (r *Repository) GetPendingByRequesterAndName(ctx context.Context, requesterID int64, name string) (sqlc.SkillCreationRequest, error) {
	return r.queries.GetPendingSkillCreationByRequesterAndName(ctx, sqlc.GetPendingSkillCreationByRequesterAndNameParams{
		RequesterID: requesterID,
		Lower:       name,
	})
}

func (r *Repository) CreateRequest(ctx context.Context, arg sqlc.CreateSkillCreationRequestParams) (sqlc.SkillCreationRequest, error) {
	return r.queries.CreateSkillCreationRequest(ctx, arg)
}

func (r *Repository) GetByID(ctx context.Context, id int64) (sqlc.SkillCreationRequest, error) {
	return r.queries.GetSkillCreationRequestByID(ctx, id)
}

func (r *Repository) GetByPublicID(ctx context.Context, id uuid.UUID) (sqlc.SkillCreationRequest, error) {
	return r.queries.GetSkillCreationRequestByPublicID(ctx, id)
}

func (r *Repository) ListGeneratingIDs(ctx context.Context) ([]int64, error) {
	return r.queries.ListGeneratingSkillCreationRequests(ctx)
}

func (r *Repository) CompleteDraft(ctx context.Context, id int64, draftJSON []byte) (sqlc.SkillCreationRequest, error) {
	return r.queries.CompleteSkillCreationDraft(ctx, sqlc.CompleteSkillCreationDraftParams{
		ID:              id,
		DraftMilestones: draftJSON,
	})
}

func (r *Repository) FailGenerating(ctx context.Context, id int64, note string) (sqlc.SkillCreationRequest, error) {
	return r.queries.FailSkillCreationRequest(ctx, sqlc.FailSkillCreationRequestParams{
		ID:        id,
		AdminNote: pgtype.Text{String: note, Valid: note != ""},
	})
}

func (r *Repository) ListByRequester(ctx context.Context, requesterID int64, limit int32) ([]sqlc.SkillCreationRequest, error) {
	return r.queries.ListSkillCreationRequestsByRequester(ctx, sqlc.ListSkillCreationRequestsByRequesterParams{
		RequesterID: requesterID,
		ResultLimit: limit,
	})
}

func (r *Repository) ListPending(ctx context.Context, limit int32) ([]sqlc.SkillCreationRequest, error) {
	return r.queries.ListPendingSkillCreationRequests(ctx, limit)
}

func (r *Repository) Cancel(ctx context.Context, publicID uuid.UUID, requesterID int64) (sqlc.SkillCreationRequest, error) {
	return r.queries.CancelSkillCreationRequest(ctx, sqlc.CancelSkillCreationRequestParams{
		PublicID:    publicID,
		RequesterID: requesterID,
	})
}

func (r *Repository) Approve(ctx context.Context, id, reviewerID, skillID int64, note *string) (sqlc.SkillCreationRequest, error) {
	return r.queries.ApproveSkillCreationRequest(ctx, sqlc.ApproveSkillCreationRequestParams{
		ID:             id,
		ReviewedBy:     pgtype.Int8{Int64: reviewerID, Valid: true},
		CreatedSkillID: pgtype.Int8{Int64: skillID, Valid: true},
		AdminNote:      textPtr(note),
	})
}

func (r *Repository) Reject(ctx context.Context, id, reviewerID int64, note *string) (sqlc.SkillCreationRequest, error) {
	return r.queries.RejectSkillCreationRequest(ctx, sqlc.RejectSkillCreationRequestParams{
		ID:         id,
		ReviewedBy: pgtype.Int8{Int64: reviewerID, Valid: true},
		AdminNote:  textPtr(note),
	})
}

func textPtr(s *string) pgtype.Text {
	if s == nil || *s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}
