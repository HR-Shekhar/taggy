package skill

import (
	"context"
	"fmt"

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

func (r *Repository) ListActiveSkills(ctx context.Context) ([]sqlc.Skill, error) {
	return r.queries.ListActiveSkills(ctx)
}

func (r *Repository) GetSkillBySlug(ctx context.Context, slug string) (sqlc.Skill, error) {
	return r.queries.GetSkillBySlug(ctx, slug)
}

func (r *Repository) GetCommunityBySkillID(ctx context.Context, skillID int64) (sqlc.GetCommunityBySkillIDRow, error) {
	return r.queries.GetCommunityBySkillID(ctx, skillID)
}

func (r *Repository) GetActiveRoadmapVersionBySkillID(ctx context.Context, skillID int64) (sqlc.RoadmapVersion, error) {
	return r.queries.GetCurrentRoadmapVersionBySkillID(ctx, skillID)
}

func (r *Repository) GetRoadmapVersionBySkillSlugAndNumber(ctx context.Context, skillSlug string, versionNumber int32) (sqlc.GetRoadmapVersionBySkillSlugAndNumberRow, error) {
	return r.queries.GetRoadmapVersionBySkillSlugAndNumber(ctx, sqlc.GetRoadmapVersionBySkillSlugAndNumberParams{
		Slug:          skillSlug,
		VersionNumber: versionNumber,
	})
}

func (r *Repository) ListCompletedMilestoneSlugsByUserSkillID(ctx context.Context, userSkillID int64) ([]string, error) {
	return r.queries.ListCompletedMilestoneSlugsByUserSkillID(ctx, pgtype.Int8{Int64: userSkillID, Valid: true})
}

// SwitchUserSkillRoadmapVersion remaps enrollment to another published version.
// Completed milestones that share a slug are preserved as COMPLETED.
func (r *Repository) SwitchUserSkillRoadmapVersion(
	ctx context.Context,
	userSkillID int64,
	newVersionID int64,
	milestones []sqlc.ListMilestonesByRoadmapVersionIDRow,
	completedSlugs map[string]struct{},
) (sqlc.Userskill, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return sqlc.Userskill{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	if err := qtx.DeleteMilestoneProgressByUserSkillID(ctx, pgtype.Int8{Int64: userSkillID, Valid: true}); err != nil {
		return sqlc.Userskill{}, err
	}

	userSkill, err := qtx.UpdateUserSkillRoadmapVersion(ctx, sqlc.UpdateUserSkillRoadmapVersionParams{
		ID:               userSkillID,
		RoadmapVersionID: newVersionID,
	})
	if err != nil {
		return sqlc.Userskill{}, err
	}

	for _, m := range milestones {
		status := sqlc.ProgressStatusNOTSTARTED
		if _, ok := completedSlugs[m.Slug]; ok {
			status = sqlc.ProgressStatusCOMPLETED
		}

		progress, err := qtx.CreateMilestoneProgress(ctx, sqlc.CreateMilestoneProgressParams{
			UserSkillID: pgtype.Int8{Int64: userSkill.ID, Valid: true},
			MilestoneID: pgtype.Int8{Int64: m.ID, Valid: true},
			Status:      status,
		})
		if err != nil {
			return sqlc.Userskill{}, err
		}
		if status == sqlc.ProgressStatusCOMPLETED {
			if _, err := qtx.CompleteMilestoneProgress(ctx, progress.ID); err != nil {
				return sqlc.Userskill{}, err
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return sqlc.Userskill{}, fmt.Errorf("commit transaction: %w", err)
	}

	return userSkill, nil
}

func (r *Repository) ListMilestonesByRoadmapVersionID(ctx context.Context, versionID int64) ([]sqlc.ListMilestonesByRoadmapVersionIDRow, error) {
	return r.queries.ListMilestonesByRoadmapVersionID(ctx, pgtype.Int8{Int64: versionID, Valid: true})
}

func (r *Repository) CountActiveUserSkillsByUserID(ctx context.Context, userID int64) (int64, error) {
	return r.queries.CountActiveUserSkillsByUserID(ctx, userID)
}

func (r *Repository) GetUserSkillByUserAndSkillSlug(ctx context.Context, userID int64, skillSlug string) (sqlc.Userskill, error) {
	return r.queries.GetUserSkillByUserAndSkillSlug(ctx, sqlc.GetUserSkillByUserAndSkillSlugParams{
		UserID: userID,
		Slug:   skillSlug,
	})
}

func (r *Repository) ListUserSkillsByUserID(ctx context.Context, userID int64) ([]sqlc.ListUserSkillsByUserIDRow, error) {
	return r.queries.ListUserSkillsByUserID(ctx, userID)
}

func (r *Repository) ListMilestoneProgressByUserSkillID(ctx context.Context, userSkillID int64) ([]sqlc.ListMilestoneProgressByUserSkillIDRow, error) {
	return r.queries.ListMilestoneProgressByUserSkillID(ctx, pgtype.Int8{Int64: userSkillID, Valid: true})
}

func (r *Repository) GetMilestoneProgressBySlug(ctx context.Context, userSkillID int64, milestoneSlug string) (sqlc.GetMilestoneProgressBySlugRow, error) {
	return r.queries.GetMilestoneProgressBySlug(ctx, sqlc.GetMilestoneProgressBySlugParams{
		UserSkillID: pgtype.Int8{Int64: userSkillID, Valid: true},
		Slug:        milestoneSlug,
	})
}

func (r *Repository) CountIncompleteMilestonesBefore(ctx context.Context, userSkillID int64, orderIndex int32) (int64, error) {
	return r.queries.CountIncompleteMilestonesBefore(ctx, sqlc.CountIncompleteMilestonesBeforeParams{
		UserSkillID: pgtype.Int8{Int64: userSkillID, Valid: true},
		OrderIndex:  orderIndex,
	})
}

func (r *Repository) CountIncompleteTopicMilestonesBefore(ctx context.Context, userSkillID int64, orderIndex int32) (int64, error) {
	return r.queries.CountIncompleteTopicMilestonesBefore(ctx, sqlc.CountIncompleteTopicMilestonesBeforeParams{
		UserSkillID: pgtype.Int8{Int64: userSkillID, Valid: true},
		OrderIndex:  orderIndex,
	})
}

func (r *Repository) CountIncompleteTopicsInChapter(ctx context.Context, userSkillID int64, chapter string) (int64, error) {
	return r.queries.CountIncompleteTopicsInChapter(ctx, sqlc.CountIncompleteTopicsInChapterParams{
		UserSkillID: pgtype.Int8{Int64: userSkillID, Valid: true},
		Chapter:     pgtype.Text{String: chapter, Valid: chapter != ""},
	})
}

func (r *Repository) CountIncompleteChaptersBefore(ctx context.Context, userSkillID int64, orderIndex int32) (int64, error) {
	return r.queries.CountIncompleteChaptersBefore(ctx, sqlc.CountIncompleteChaptersBeforeParams{
		UserSkillID: pgtype.Int8{Int64: userSkillID, Valid: true},
		OrderIndex:  orderIndex,
	})
}

func (r *Repository) GetMilestoneBySlugAndRoadmapVersion(ctx context.Context, roadmapVersionID int64, slug string) (sqlc.GetMilestoneBySlugAndRoadmapVersionRow, error) {
	return r.queries.GetMilestoneBySlugAndRoadmapVersion(ctx, sqlc.GetMilestoneBySlugAndRoadmapVersionParams{
		RoadmapVersionID: pgtype.Int8{Int64: roadmapVersionID, Valid: true},
		Slug:             slug,
	})
}

func (r *Repository) CompleteMilestoneProgress(ctx context.Context, progressID int64) (sqlc.UserMilestoneProgress, error) {
	return r.queries.CompleteMilestoneProgress(ctx, progressID)
}

func (r *Repository) PostponeMilestoneProgress(ctx context.Context, progressID int64, postponedUntil pgtype.Timestamptz) (sqlc.UserMilestoneProgress, error) {
	return r.queries.PostponeMilestoneProgress(ctx, sqlc.PostponeMilestoneProgressParams{
		ID:             progressID,
		PostponedUntil: postponedUntil,
	})
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

// EnrollUserInSkill atomically creates userskill and milestone progress rows.
func (r *Repository) EnrollUserInSkill(
	ctx context.Context,
	userID int64,
	skillID int64,
	roadmapVersionID int64,
	milestoneIDs []int64,
) (sqlc.Userskill, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return sqlc.Userskill{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	userSkill, err := qtx.CreateUserSkill(ctx, sqlc.CreateUserSkillParams{
		UserID:           userID,
		SkillID:          skillID,
		RoadmapVersionID: roadmapVersionID,
		Status:           sqlc.StatusValueACTIVE,
	})
	if err != nil {
		return sqlc.Userskill{}, err
	}

	for _, milestoneID := range milestoneIDs {
		_, err := qtx.CreateMilestoneProgress(ctx, sqlc.CreateMilestoneProgressParams{
			UserSkillID: pgtype.Int8{Int64: userSkill.ID, Valid: true},
			MilestoneID: pgtype.Int8{Int64: milestoneID, Valid: true},
			Status:      sqlc.ProgressStatusNOTSTARTED,
		})
		if err != nil {
			return sqlc.Userskill{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return sqlc.Userskill{}, fmt.Errorf("commit transaction: %w", err)
	}

	return userSkill, nil
}
