package roadmap

import (
	"context"
	"errors"

	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/postgres/sqlc"
	"github.com/HR-Shekhar/taggy-backend/internal/shared/logging"
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

type RoadmapOverview struct {
	SkillSlug string
	SkillName string
	Versions  []sqlc.ListRoadmapVersionsBySkillSlugRow
}

type VersionDetail struct {
	Version    sqlc.GetRoadmapVersionBySkillSlugAndNumberRow
	Milestones []sqlc.ListMilestonesByRoadmapVersionIDRow
}

func (s *Service) GetRoadmapOverview(ctx context.Context, skillSlug string) (RoadmapOverview, error) {
	roadmap, err := s.repo.GetRoadmapBySkillSlug(ctx, skillSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return RoadmapOverview{}, logging.Reject(s.log, ErrRoadmapNotFound, "roadmap overview rejected: not found")
		}
		return RoadmapOverview{}, logging.Unexpected(s.log, err, "get roadmap by skill slug failed")
	}

	versions, err := s.repo.ListVersionsBySkillSlug(ctx, skillSlug)
	if err != nil {
		return RoadmapOverview{}, logging.Unexpected(s.log, err, "list roadmap versions failed")
	}

	return RoadmapOverview{
		SkillSlug: roadmap.SkillSlug,
		SkillName: roadmap.SkillName,
		Versions:  versions,
	}, nil
}

func (s *Service) ListVersions(ctx context.Context, skillSlug string) ([]sqlc.ListRoadmapVersionsBySkillSlugRow, error) {
	_, err := s.repo.GetRoadmapBySkillSlug(ctx, skillSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, logging.Reject(s.log, ErrRoadmapNotFound, "list versions rejected: roadmap not found")
		}
		return nil, logging.Unexpected(s.log, err, "get roadmap for list versions failed")
	}

	versions, err := s.repo.ListVersionsBySkillSlug(ctx, skillSlug)
	if err != nil {
		return nil, logging.Unexpected(s.log, err, "list roadmap versions failed")
	}
	return versions, nil
}

func (s *Service) GetVersionDetail(ctx context.Context, skillSlug string, versionNumber int32) (VersionDetail, error) {
	version, err := s.repo.GetVersionBySkillSlugAndNumber(ctx, skillSlug, versionNumber)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return VersionDetail{}, logging.Reject(s.log, ErrVersionNotFound, "version detail rejected: not found")
		}
		return VersionDetail{}, logging.Unexpected(s.log, err, "get roadmap version failed")
	}

	milestones, err := s.repo.ListMilestonesByRoadmapVersionID(ctx, version.ID)
	if err != nil {
		return VersionDetail{}, logging.Unexpected(s.log, err, "list milestones for version failed")
	}

	return VersionDetail{Version: version, Milestones: milestones}, nil
}

func formatTime(t pgtype.Timestamptz) string {
	if !t.Valid {
		return ""
	}
	return t.Time.UTC().Format("2006-01-02T15:04:05Z07:00")
}

func optionalTime(t pgtype.Timestamptz) *string {
	if !t.Valid {
		return nil
	}
	s := formatTime(t)
	return &s
}
