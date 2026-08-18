package progress

import (
	"context"
	"errors"
	"fmt"

	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

func (r *Repository) GetUserPublicIDByUsername(ctx context.Context, username string) (uuid.UUID, error) {
	user, err := r.queries.GetUserByUsername(ctx, username)
	if err != nil {
		return uuid.UUID{}, err
	}
	return user.PublicID, nil
}

func (r *Repository) GetSkillBySlug(ctx context.Context, slug string) (sqlc.Skill, error) {
	return r.queries.GetSkillBySlug(ctx, slug)
}

func (r *Repository) GetActiveUserSkillBySkillSlug(ctx context.Context, userID int64, skillSlug string) (sqlc.Userskill, error) {
	row, err := r.queries.GetUserSkillByUserAndSkillSlug(ctx, sqlc.GetUserSkillByUserAndSkillSlugParams{
		UserID: userID,
		Slug:   skillSlug,
	})
	if err != nil {
		return sqlc.Userskill{}, err
	}
	if row.Status != sqlc.StatusValueACTIVE {
		return sqlc.Userskill{}, pgx.ErrNoRows
	}
	return row, nil
}

func (r *Repository) ListStudySessions(ctx context.Context, userID int64, skillSlug *string) ([]sqlc.ListStudySessionsByUserIDRow, error) {
	if skillSlug != nil {
		rows, err := r.queries.ListStudySessionsByUserAndSkillSlug(ctx, sqlc.ListStudySessionsByUserAndSkillSlugParams{
			UserID: pgtype.Int8{Int64: userID, Valid: true},
			Slug:   *skillSlug,
		})
		if err != nil {
			return nil, err
		}
		out := make([]sqlc.ListStudySessionsByUserIDRow, 0, len(rows))
		for _, row := range rows {
			out = append(out, sqlc.ListStudySessionsByUserIDRow{
				ID:              row.ID,
				UserID:          row.UserID,
				SkillID:         row.SkillID,
				DurationMinutes: row.DurationMinutes,
				Notes:           row.Notes,
				StudiedAt:       row.StudiedAt,
				CreatedAt:       row.CreatedAt,
				SkillSlug:       row.SkillSlug,
			})
		}
		return out, nil
	}
	return r.queries.ListStudySessionsByUserID(ctx, pgtype.Int8{Int64: userID, Valid: true})
}

func (r *Repository) GetStreak(ctx context.Context, userID int64) (sqlc.Streak, error) {
	return r.queries.GetStreakByUserID(ctx, pgtype.Int8{Int64: userID, Valid: true})
}

func (r *Repository) SumStudyMinutes(ctx context.Context, userID int64) (int64, int64, int64, error) {
	total, err := r.queries.SumStudyMinutesByUserID(ctx, pgtype.Int8{Int64: userID, Valid: true})
	if err != nil {
		return 0, 0, 0, err
	}
	weekly, err := r.queries.SumStudyMinutesLast7Days(ctx, pgtype.Int8{Int64: userID, Valid: true})
	if err != nil {
		return 0, 0, 0, err
	}
	monthly, err := r.queries.SumStudyMinutesThisMonth(ctx, pgtype.Int8{Int64: userID, Valid: true})
	if err != nil {
		return 0, 0, 0, err
	}
	return total, weekly, monthly, nil
}

func (r *Repository) LogStudySessionWithStreak(
	ctx context.Context,
	session sqlc.CreateStudySessionParams,
	streakFn func(old *sqlc.Streak) (StreakWrite, error),
) (sqlc.StudySession, sqlc.Streak, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return sqlc.StudySession{}, sqlc.Streak{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	var oldStreak *sqlc.Streak
	streakRow, err := qtx.GetStreakByUserIDForUpdate(ctx, session.UserID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return sqlc.StudySession{}, sqlc.Streak{}, err
		}
	} else {
		oldStreak = &streakRow
	}

	streakWrite, err := streakFn(oldStreak)
	if err != nil {
		return sqlc.StudySession{}, sqlc.Streak{}, err
	}

	studySession, err := qtx.CreateStudySession(ctx, session)
	if err != nil {
		return sqlc.StudySession{}, sqlc.Streak{}, err
	}

	activityDate := pgtype.Date{
		Time:  truncateToUTCDate(streakWrite.LastActivityDate),
		Valid: true,
	}

	var updatedStreak sqlc.Streak
	if streakWrite.IsNew {
		updatedStreak, err = qtx.CreateStreak(ctx, sqlc.CreateStreakParams{
			UserID:           session.UserID,
			CurrentStreak:    streakWrite.CurrentStreak,
			LongestStreak:    streakWrite.LongestStreak,
			LastActivityDate: activityDate,
		})
	} else {
		updatedStreak, err = qtx.UpdateStreak(ctx, sqlc.UpdateStreakParams{
			UserID:           session.UserID,
			CurrentStreak:    streakWrite.CurrentStreak,
			LongestStreak:    streakWrite.LongestStreak,
			LastActivityDate: activityDate,
		})
	}
	if err != nil {
		return sqlc.StudySession{}, sqlc.Streak{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return sqlc.StudySession{}, sqlc.Streak{}, fmt.Errorf("commit transaction: %w", err)
	}

	return studySession, updatedStreak, nil
}
