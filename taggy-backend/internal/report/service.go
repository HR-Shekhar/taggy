package report

import (
	"context"
	"errors"
	"strings"

	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/postgres/sqlc"
	apperrors "github.com/HR-Shekhar/taggy-backend/internal/shared/errors"
	"github.com/HR-Shekhar/taggy-backend/internal/shared/logging"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"
)

const (
	defaultLimit int32 = 50
	maxLimit     int32 = 100
)

type Service struct {
	repo *Repository
	log  zerolog.Logger
}

func NewService(repo *Repository, log zerolog.Logger) *Service {
	return &Service{repo: repo, log: log}
}

func (s *Service) GetUserPublicIDByUsername(ctx context.Context, username string) (uuid.UUID, error) {
	id, err := s.repo.GetUserPublicIDByUsername(ctx, username)
	if err != nil {
		return uuid.Nil, logging.Unexpected(s.log, err, "get user public id by username failed")
	}
	return id, nil
}

func (s *Service) Create(ctx context.Context, userPublicID uuid.UUID, input CreateInput) (sqlc.Report, error) {
	targetType, err := parseTargetType(input.TargetType)
	if err != nil {
		return sqlc.Report{}, logging.Reject(s.log, err, "create report rejected: invalid target type")
	}
	reason := strings.TrimSpace(input.Reason)
	if len(reason) < 3 || len(reason) > 2000 {
		return sqlc.Report{}, logging.Reject(s.log, ErrInvalidReason, "create report rejected: invalid reason")
	}
	if input.TargetID <= 0 {
		return sqlc.Report{}, logging.Reject(s.log, ErrInvalidTarget, "create report rejected: invalid target id")
	}

	user, err := s.repo.GetUserByPublicID(ctx, userPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.Report{}, logging.Reject(s.log, apperrors.ErrNotFound, "create report rejected: user not found")
		}
		return sqlc.Report{}, logging.Unexpected(s.log, err, "create report user lookup failed")
	}

	if targetType == sqlc.ReportTargetTypeUSER && input.TargetID == user.ID {
		return sqlc.Report{}, logging.Reject(s.log, ErrCannotReportSelf, "create report rejected: cannot report self")
	}

	if _, err := s.repo.GetOpenReportByReporterAndTarget(ctx, user.ID, targetType, input.TargetID); err == nil {
		return sqlc.Report{}, logging.Reject(s.log, ErrDuplicateOpen, "create report rejected: duplicate open")
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return sqlc.Report{}, logging.Unexpected(s.log, err, "create report duplicate check failed")
	}

	created, err := s.repo.CreateReport(ctx, sqlc.CreateReportParams{
		ReporterID: user.ID,
		TargetType: targetType,
		TargetID:   input.TargetID,
		Reason:     reason,
	})
	if err != nil {
		return sqlc.Report{}, logging.Unexpected(s.log, err, "create report failed")
	}

	s.log.Info().
		Int64("report_id", created.ID).
		Str("target_type", string(targetType)).
		Int64("target_id", input.TargetID).
		Str("reporter", user.Username).
		Msg("report created")

	return created, nil
}

func (s *Service) ListMine(ctx context.Context, userPublicID uuid.UUID, limit int32) ([]sqlc.Report, error) {
	user, err := s.repo.GetUserByPublicID(ctx, userPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, logging.Unexpected(s.log, err, "list reports user lookup failed")
	}
	rows, err := s.repo.ListReportsByReporterID(ctx, user.ID, normalizeLimit(limit))
	if err != nil {
		return nil, logging.Unexpected(s.log, err, "list reports failed")
	}
	return rows, nil
}

func parseTargetType(raw string) (sqlc.ReportTargetType, error) {
	t := sqlc.ReportTargetType(strings.ToUpper(strings.TrimSpace(raw)))
	switch t {
	case sqlc.ReportTargetTypeUSER,
		sqlc.ReportTargetTypePROPOSAL,
		sqlc.ReportTargetTypePOD,
		sqlc.ReportTargetTypeMESSAGE,
		sqlc.ReportTargetTypeAUDIOROOM,
		sqlc.ReportTargetTypeCOMMUNITYCHANNEL:
		return t, nil
	default:
		return "", ErrInvalidTarget
	}
}

func normalizeLimit(limit int32) int32 {
	if limit <= 0 {
		return defaultLimit
	}
	if limit > maxLimit {
		return maxLimit
	}
	return limit
}
