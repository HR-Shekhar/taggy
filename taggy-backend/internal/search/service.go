package search

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/postgres/sqlc"
	"github.com/HR-Shekhar/taggy-backend/internal/shared/logging"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

const (
	defaultLimit int32 = 20
	maxLimit     int32 = 50
	minQueryLen        = 2
	maxQueryLen        = 80
)

type Service struct {
	repo *Repository
	log  zerolog.Logger
}

func NewService(repo *Repository, log zerolog.Logger) *Service {
	return &Service{repo: repo, log: log}
}

func (s *Service) Search(ctx context.Context, input Input) (Result, error) {
	query := sanitizeQuery(input.Query)
	if utf8.RuneCountInString(query) < minQueryLen || utf8.RuneCountInString(query) > maxQueryLen {
		return Result{}, logging.Reject(s.log, ErrInvalidQuery, "search rejected: invalid query")
	}

	types, err := parseTypes(input.Types)
	if err != nil {
		return Result{}, logging.Reject(s.log, err, "search rejected: invalid type")
	}
	limit := normalizeLimit(input.Limit)

	g, gctx := errgroup.WithContext(ctx)

	var (
		skills      []SkillHit
		users       []UserHit
		communities []CommunityHit
	)

	if types["skills"] {
		g.Go(func() error {
			rows, err := s.repo.SearchSkills(gctx, query, limit)
			if err != nil {
				return logging.Unexpected(s.log, err, "search skills failed")
			}
			skills = mapSkillHits(rows)
			return nil
		})
	}
	if types["users"] {
		g.Go(func() error {
			rows, err := s.repo.SearchUsers(gctx, query, limit)
			if err != nil {
				return logging.Unexpected(s.log, err, "search users failed")
			}
			users = mapUserHits(rows)
			return nil
		})
	}
	if types["communities"] {
		g.Go(func() error {
			rows, err := s.repo.SearchCommunities(gctx, query, limit)
			if err != nil {
				return logging.Unexpected(s.log, err, "search communities failed")
			}
			communities = mapCommunityHits(rows)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return Result{}, err
	}

	out := Result{
		Skills:      skills,
		Users:       users,
		Communities: communities,
	}

	s.log.Info().
		Str("query", query).
		Int("skills", len(out.Skills)).
		Int("users", len(out.Users)).
		Int("communities", len(out.Communities)).
		Msg("search completed")

	return out, nil
}

func mapSkillHits(rows []sqlc.Skill) []SkillHit {
	out := make([]SkillHit, 0, len(rows))
	for _, row := range rows {
		out = append(out, SkillHit{
			ID:          row.ID,
			Name:        row.Name,
			Slug:        row.Slug,
			Description: textPtr(row.Description),
		})
	}
	return out
}

func mapUserHits(rows []sqlc.SearchUsersRow) []UserHit {
	out := make([]UserHit, 0, len(rows))
	for _, row := range rows {
		out = append(out, UserHit{
			PublicID:          row.PublicID.String(),
			Username:          row.Username,
			Name:              row.Name,
			ProfilePictureURL: textPtr(row.ProfilePictureUrl),
			Bio:               textPtr(row.Bio),
		})
	}
	return out
}

func mapCommunityHits(rows []sqlc.SearchCommunitiesRow) []CommunityHit {
	out := make([]CommunityHit, 0, len(rows))
	for _, row := range rows {
		out = append(out, CommunityHit{
			ID:          row.ID,
			Name:        row.Name,
			Description: textPtr(row.Description),
			SkillSlug:   row.SkillSlug,
			SkillName:   row.SkillName,
		})
	}
	return out
}

func sanitizeQuery(raw string) string {
	q := strings.TrimSpace(raw)
	q = strings.Map(func(r rune) rune {
		if r == '%' || r == '_' {
			return -1
		}
		return r
	}, q)
	return strings.Join(strings.Fields(q), " ")
}

func parseTypes(raw []string) (map[string]bool, error) {
	allowed := map[string]bool{
		"skills":      true,
		"users":       true,
		"communities": true,
	}
	if len(raw) == 0 {
		return map[string]bool{
			"skills":      true,
			"users":       true,
			"communities": true,
		}, nil
	}
	out := make(map[string]bool, len(raw))
	for _, t := range raw {
		t = strings.ToLower(strings.TrimSpace(t))
		if t == "" {
			continue
		}
		if !allowed[t] {
			return nil, ErrInvalidType
		}
		out[t] = true
	}
	if len(out) == 0 {
		return nil, ErrInvalidType
	}
	return out, nil
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

func textPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	s := t.String
	return &s
}
