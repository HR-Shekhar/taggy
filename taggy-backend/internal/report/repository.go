package report

import (
	"context"

	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool, queries: sqlc.New(pool)}
}

func (r *Repository) GetUserByPublicID(ctx context.Context, publicID uuid.UUID) (sqlc.User, error) {
	return r.queries.GetUserByPublicID(ctx, publicID)
}

func (r *Repository) GetUserPublicIDByUsername(ctx context.Context, username string) (uuid.UUID, error) {
	user, err := r.queries.GetUserByUsername(ctx, username)
	if err != nil {
		return uuid.UUID{}, err
	}
	return user.PublicID, nil
}

func (r *Repository) CreateReport(ctx context.Context, arg sqlc.CreateReportParams) (sqlc.Report, error) {
	return r.queries.CreateReport(ctx, arg)
}

func (r *Repository) GetOpenReportByReporterAndTarget(
	ctx context.Context,
	reporterID int64,
	targetType sqlc.ReportTargetType,
	targetID int64,
) (sqlc.Report, error) {
	return r.queries.GetOpenReportByReporterAndTarget(ctx, sqlc.GetOpenReportByReporterAndTargetParams{
		ReporterID: reporterID,
		TargetType: targetType,
		TargetID:   targetID,
	})
}

func (r *Repository) ListReportsByReporterID(ctx context.Context, reporterID int64, limit int32) ([]sqlc.Report, error) {
	return r.queries.ListReportsByReporterID(ctx, sqlc.ListReportsByReporterIDParams{
		ReporterID: reporterID,
		Limit:      limit,
	})
}
