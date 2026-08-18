package search

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/HR-Shekhar/taggy-backend/internal/shared/logging"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
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

	var out Result
	if types["skills"] {
		rows, err := s.repo.SearchSkills(ctx, query, limit)
		if err != nil {
			return Result{}, logging.Unexpected(s.log, err, "search skills failed")
		}
		out.Skills = make([]SkillHit, 0, len(rows))
		for _, row := range rows {
			out.Skills = append(out.Skills, SkillHit{
				ID:          row.ID,
				Name:        row.Name,
				Slug:        row.Slug,
				Description: textPtr(row.Description),
			})
		}
	}
	if types["users"] {
		rows, err := s.repo.SearchUsers(ctx, query, limit)
		if err != nil {
			return Result{}, logging.Unexpected(s.log, err, "search users failed")
		}
		out.Users = make([]UserHit, 0, len(rows))
		for _, row := range rows {
			out.Users = append(out.Users, UserHit{
				PublicID:          row.PublicID.String(),
				Username:          row.Username,
				Name:              row.Name,
				ProfilePictureURL: textPtr(row.ProfilePictureUrl),
				Bio:               textPtr(row.Bio),
			})
		}
	}
	if types["communities"] {
		rows, err := s.repo.SearchCommunities(ctx, query, limit)
		if err != nil {
			return Result{}, logging.Unexpected(s.log, err, "search communities failed")
		}
		out.Communities = make([]CommunityHit, 0, len(rows))
		for _, row := range rows {
			out.Communities = append(out.Communities, CommunityHit{
				ID:          row.ID,
				Name:        row.Name,
				Description: textPtr(row.Description),
				SkillSlug:   row.SkillSlug,
				SkillName:   row.SkillName,
			})
		}
	}

	s.log.Info().
		Str("query", query).
		Int("skills", len(out.Skills)).
		Int("users", len(out.Users)).
		Int("communities", len(out.Communities)).
		Msg("search completed")

	return out, nil
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
