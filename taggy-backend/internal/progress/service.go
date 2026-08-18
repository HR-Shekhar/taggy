package progress

import (
	"context"
	"errors"
	"time"

	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/postgres/sqlc"
	apperrors "github.com/HR-Shekhar/taggy-backend/internal/shared/errors"
	"github.com/HR-Shekhar/taggy-backend/internal/shared/logging"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
)

type Service struct {
	repo *Repository
	log  zerolog.Logger
}

func NewService(repo *Repository, log zerolog.Logger) *Service {
	return &Service{repo: repo, log: log}
}

func (s *Service) LogStudySession(
	ctx context.Context,
	userPublicID uuid.UUID,
	input LogStudySessionInput,
) (sqlc.StudySession, sqlc.Streak, error) {
	user, err := s.repo.GetUserByPublicID(ctx, userPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.StudySession{}, sqlc.Streak{}, logging.Reject(s.log, apperrors.ErrNotFound, "study log rejected: user not found")
		}
		return sqlc.StudySession{}, sqlc.Streak{}, logging.Unexpected(s.log, err, "get user by public id failed")
	}

	if _, err := s.repo.GetActiveUserSkillBySkillSlug(ctx, user.ID, input.SkillSlug); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.log.Warn().
				Str("user_id", user.PublicID.String()).
				Str("skill_slug", input.SkillSlug).
				Msg("study log rejected: not enrolled")
			return sqlc.StudySession{}, sqlc.Streak{}, ErrNotEnrolledInSkill
		}
		return sqlc.StudySession{}, sqlc.Streak{}, logging.Unexpected(s.log, err, "get active user skill failed")
	}

	skill, err := s.repo.GetSkillBySlug(ctx, input.SkillSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.StudySession{}, sqlc.Streak{}, logging.Reject(s.log, apperrors.ErrNotFound, "study log rejected: skill not found")
		}
		return sqlc.StudySession{}, sqlc.Streak{}, logging.Unexpected(s.log, err, "get skill by slug failed")
	}

	sessionParams := sqlc.CreateStudySessionParams{
		UserID:          pgtype.Int8{Int64: user.ID, Valid: true},
		SkillID:         pgtype.Int8{Int64: skill.ID, Valid: true},
		DurationMinutes: input.DurationMinutes,
		StudiedAt:       pgtype.Timestamptz{Time: input.StudiedAt, Valid: true},
	}
	if input.Notes != nil {
		sessionParams.Notes = pgtype.Text{String: *input.Notes, Valid: true}
	}

	studySession, streak, err := s.repo.LogStudySessionWithStreak(
		ctx,
		sessionParams,
		func(old *sqlc.Streak) (StreakWrite, error) {
			return computeStreakUpdate(old, input.StudiedAt), nil
		},
	)
	if err != nil {
		return sqlc.StudySession{}, sqlc.Streak{}, logging.Unexpected(s.log, err, "log study session with streak failed")
	}

	s.log.Info().
		Str("user_id", user.PublicID.String()).
		Str("skill_slug", input.SkillSlug).
		Int32("duration_minutes", input.DurationMinutes).
		Int32("current_streak", streak.CurrentStreak).
		Msg("study session logged")

	return studySession, streak, nil
}

func (s *Service) GetUserPublicIDByUsername(ctx context.Context, username string) (uuid.UUID, error) {
	id, err := s.repo.GetUserPublicIDByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, err
		}
		return uuid.Nil, logging.Unexpected(s.log, err, "get user public id by username failed")
	}
	return id, nil
}

func (s *Service) ListStudySessions(
	ctx context.Context,
	userPublicID uuid.UUID,
	skillSlug *string,
) ([]sqlc.ListStudySessionsByUserIDRow, error) {
	user, err := s.repo.GetUserByPublicID(ctx, userPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, logging.Unexpected(s.log, err, "get user by public id failed")
	}

	if skillSlug != nil {
		if _, err := s.repo.GetActiveUserSkillBySkillSlug(ctx, user.ID, *skillSlug); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrNotEnrolledInSkill
			}
			return nil, logging.Unexpected(s.log, err, "get active user skill failed")
		}
	}

	rows, err := s.repo.ListStudySessions(ctx, user.ID, skillSlug)
	if err != nil {
		return nil, logging.Unexpected(s.log, err, "list study sessions failed")
	}
	return rows, nil
}

func (s *Service) GetStreak(ctx context.Context, userPublicID uuid.UUID) (sqlc.Streak, error) {
	user, err := s.repo.GetUserByPublicID(ctx, userPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.Streak{}, apperrors.ErrNotFound
		}
		return sqlc.Streak{}, logging.Unexpected(s.log, err, "get user by public id failed")
	}

	streak, err := s.repo.GetStreak(ctx, user.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.Streak{
				UserID:        pgtype.Int8{Int64: user.ID, Valid: true},
				CurrentStreak: 0,
				LongestStreak: 0,
				FreezeCount:   0,
			}, nil
		}
		return sqlc.Streak{}, logging.Unexpected(s.log, err, "get streak failed")
	}

	return streak, nil
}

func (s *Service) GetProgressSummary(ctx context.Context, userPublicID uuid.UUID) (ProgressSummary, error) {
	user, err := s.repo.GetUserByPublicID(ctx, userPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ProgressSummary{}, apperrors.ErrNotFound
		}
		return ProgressSummary{}, logging.Unexpected(s.log, err, "get user by public id failed")
	}

	total, weekly, monthly, err := s.repo.SumStudyMinutes(ctx, user.ID)
	if err != nil {
		return ProgressSummary{}, logging.Unexpected(s.log, err, "sum study minutes failed")
	}

	streak, err := s.GetStreak(ctx, userPublicID)
	if err != nil {
		return ProgressSummary{}, err
	}

	return ProgressSummary{
		TotalMinutes:   total,
		WeeklyMinutes:  weekly,
		MonthlyMinutes: monthly,
		CurrentStreak:  streak.CurrentStreak,
		LongestStreak:  streak.LongestStreak,
	}, nil
}

// ProgressSummary aggregates study hours and streak for the dashboard.
type ProgressSummary struct {
	TotalMinutes   int64
	WeeklyMinutes  int64
	MonthlyMinutes int64
	CurrentStreak  int32
	LongestStreak  int32
}

// computeStreakUpdate is pure business logic for streak rules.
//
// Rules (calendar days in UTC):
//   - First ever session → streak = 1
//   - Another session same day → streak unchanged
//   - Session on consecutive day → streak + 1
//   - Gap > 1 day → streak resets to 1
func computeStreakUpdate(old *sqlc.Streak, studiedAt time.Time) StreakWrite {
	studyDate := truncateToUTCDate(studiedAt)

	if old == nil || !old.LastActivityDate.Valid {
		return StreakWrite{
			CurrentStreak:    1,
			LongestStreak:    1,
			LastActivityDate: studyDate,
			IsNew:            true,
		}
	}

	lastDate := truncateToUTCDate(old.LastActivityDate.Time)

	if sameUTCDate(studyDate, lastDate) {
		return StreakWrite{
			CurrentStreak:    old.CurrentStreak,
			LongestStreak:    old.LongestStreak,
			LastActivityDate: studyDate,
			IsNew:            false,
		}
	}

	var newCurrent int32
	if isPreviousUTCDate(lastDate, studyDate) {
		newCurrent = old.CurrentStreak + 1
	} else {
		newCurrent = 1
	}

	newLongest := old.LongestStreak
	if newCurrent > newLongest {
		newLongest = newCurrent
	}

	return StreakWrite{
		CurrentStreak:    newCurrent,
		LongestStreak:    newLongest,
		LastActivityDate: studyDate,
		IsNew:            false,
	}
}

func truncateToUTCDate(t time.Time) time.Time {
	utc := t.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}

func sameUTCDate(a, b time.Time) bool {
	return truncateToUTCDate(a).Equal(truncateToUTCDate(b))
}

func isPreviousUTCDate(previous, current time.Time) bool {
	prev := truncateToUTCDate(previous)
	curr := truncateToUTCDate(current)
	return curr.AddDate(0, 0, -1).Equal(prev)
}
