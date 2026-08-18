package roadmap

import (
	"context"

	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/postgres/sqlc"
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

func (r *Repository) GetRoadmapBySkillSlug(ctx context.Context, slug string) (sqlc.GetRoadmapBySkillSlugRow, error) {
	return r.queries.GetRoadmapBySkillSlug(ctx, slug)
}

func (r *Repository) ListVersionsBySkillSlug(ctx context.Context, slug string) ([]sqlc.ListRoadmapVersionsBySkillSlugRow, error) {
	return r.queries.ListRoadmapVersionsBySkillSlug(ctx, slug)
}

func (r *Repository) GetVersionBySkillSlugAndNumber(ctx context.Context, slug string, versionNumber int32) (sqlc.GetRoadmapVersionBySkillSlugAndNumberRow, error) {
	return r.queries.GetRoadmapVersionBySkillSlugAndNumber(ctx, sqlc.GetRoadmapVersionBySkillSlugAndNumberParams{
		Slug:          slug,
		VersionNumber: versionNumber,
	})
}

func (r *Repository) ListMilestonesByRoadmapVersionID(ctx context.Context, versionID int64) ([]sqlc.ListMilestonesByRoadmapVersionIDRow, error) {
	return r.queries.ListMilestonesByRoadmapVersionID(ctx, pgtype.Int8{Int64: versionID, Valid: true})
}

func (r *Repository) GetCurrentRoadmapVersionBySkillID(ctx context.Context, skillID int64) (sqlc.RoadmapVersion, error) {
	return r.queries.GetCurrentRoadmapVersionBySkillID(ctx, skillID)
}
