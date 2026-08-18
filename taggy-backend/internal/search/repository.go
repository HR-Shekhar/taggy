package search

import (
	"context"

	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/postgres/sqlc"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool, queries: sqlc.New(pool)}
}

func (r *Repository) SearchSkills(ctx context.Context, query string, limit int32) ([]sqlc.Skill, error) {
	return r.queries.SearchSkills(ctx, sqlc.SearchSkillsParams{Query: query, Lim: limit})
}

func (r *Repository) SearchUsers(ctx context.Context, query string, limit int32) ([]sqlc.SearchUsersRow, error) {
	return r.queries.SearchUsers(ctx, sqlc.SearchUsersParams{Query: query, Lim: limit})
}

func (r *Repository) SearchCommunities(ctx context.Context, query string, limit int32) ([]sqlc.SearchCommunitiesRow, error) {
	return r.queries.SearchCommunities(ctx, sqlc.SearchCommunitiesParams{Query: query, Lim: limit})
}
