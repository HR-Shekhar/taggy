package skillrequest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/HR-Shekhar/taggy-backend/internal/aigen"
	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/openrouter"
	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/postgres/sqlc"
	apperrors "github.com/HR-Shekhar/taggy-backend/internal/shared/errors"
	"github.com/HR-Shekhar/taggy-backend/internal/shared/logging"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
)

const (
	similarMinScore      float32 = 0.3  // soft confirm (force allowed)
	similarHardBlockScore float32 = 0.72 // near-duplicate: roadmap already exists
	similarLimit         int32   = 5
	defaultListLimit     int32   = 50
)

type Generator interface {
	Available() bool
	EvaluateSkillRequest(ctx context.Context, name, description string) (openrouter.SkillRequestEvaluation, error)
	GenerateRoadmap(ctx context.Context, skillName, description, rationale, currentOutline string) ([]openrouter.MilestoneDraft, error)
}

type Notifier interface {
	NotifySkillRequestApproved(ctx context.Context, userID, skillID int64, skillSlug string)
	NotifySkillRequestRejected(ctx context.Context, userID int64, skillName, note string)
	NotifySkillRequestReady(ctx context.Context, userID int64, skillName string)
	NotifySkillRequestFailed(ctx context.Context, userID int64, skillName, note string)
}

type JobPool interface {
	Submit(job aigen.Job) error
}

type Service struct {
	repo      *Repository
	generator Generator
	notifier  Notifier
	pool      JobPool
	log       zerolog.Logger
}

func NewService(repo *Repository, generator Generator, notifier Notifier, pool JobPool, log zerolog.Logger) *Service {
	return &Service{repo: repo, generator: generator, notifier: notifier, pool: pool, log: log}
}

func (s *Service) GetUserPublicIDByUsername(ctx context.Context, username string) (uuid.UUID, error) {
	u, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return uuid.UUID{}, err
	}
	return u.PublicID, nil
}

func (s *Service) ListSimilar(ctx context.Context, query string) ([]SimilarSkill, error) {
	q := strings.TrimSpace(query)
	if len(q) < 2 {
		return nil, logging.Reject(s.log, ErrInvalidName, "similar skills rejected: query too short")
	}
	rows, err := s.repo.ListSimilarSkills(ctx, q, similarMinScore, similarLimit)
	if err != nil {
		return nil, logging.Unexpected(s.log, err, "list similar skills failed")
	}
	return mapSimilar(rows), nil
}

func (s *Service) Create(ctx context.Context, userPublicID uuid.UUID, input CreateInput) (CreateResult, error) {
	name := strings.TrimSpace(input.Name)
	if len(name) < 3 || len(name) > 255 {
		return CreateResult{}, logging.Reject(s.log, ErrInvalidName, "create skill request rejected: invalid name")
	}
	desc := strings.TrimSpace(input.Description)
	if len(desc) > 4000 {
		return CreateResult{}, logging.Reject(s.log, ErrInvalidDescription, "create skill request rejected: description too long")
	}

	user, err := s.repo.GetUserByPublicID(ctx, userPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return CreateResult{}, apperrors.ErrNotFound
		}
		return CreateResult{}, logging.Unexpected(s.log, err, "create skill request: get user failed")
	}

	if _, err := s.repo.GetPendingByRequesterAndName(ctx, user.ID, name); err == nil {
		return CreateResult{}, logging.Reject(s.log, ErrDuplicatePending, "create skill request rejected: duplicate pending")
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return CreateResult{}, logging.Unexpected(s.log, err, "create skill request: pending check failed")
	}

	similarRows, err := s.repo.ListSimilarSkills(ctx, name, similarMinScore, similarLimit)
	if err != nil {
		return CreateResult{}, logging.Unexpected(s.log, err, "create skill request: similar search failed")
	}
	similar := mapSimilar(similarRows)

	// Hard block: name is nearly identical to an existing skill — don't generate a duplicate roadmap.
	var nearDupes []SimilarSkill
	for _, sRow := range similar {
		if sRow.Score >= similarHardBlockScore {
			nearDupes = append(nearDupes, sRow)
		}
	}
	if len(nearDupes) > 0 {
		return CreateResult{
			Similar:       nearDupes,
			AlreadyExists: true,
			Message:       "A roadmap for a very similar skill already exists. Open the existing skill instead of creating a duplicate.",
		}, nil
	}

	if len(similar) > 0 && !input.Force {
		return CreateResult{
			Similar:         similar,
			RequiresConfirm: true,
			Message:         "Similar skills found. Review them, or confirm if you still want a new roadmap.",
		}, nil
	}

	slug := slugify(name)
	if existing, err := s.repo.GetSkillBySlug(ctx, slug); err == nil && existing.ID > 0 {
		desc := (*string)(nil)
		if existing.Description.Valid {
			d := existing.Description.String
			desc = &d
		}
		return CreateResult{
			Similar: []SimilarSkill{{
				ID:          existing.ID,
				Name:        existing.Name,
				Slug:        existing.Slug,
				Description: desc,
				Score:       1,
			}},
			AlreadyExists: true,
			Message:       "A roadmap for this skill already exists. Open the existing skill instead of creating a duplicate.",
		}, nil
	} else if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return CreateResult{}, logging.Unexpected(s.log, err, "create skill request: slug check failed")
	}

	if s.generator == nil || !s.generator.Available() {
		return CreateResult{}, logging.Reject(s.log, ErrAIUnavailable, "create skill request rejected: ai unavailable")
	}
	if s.pool == nil {
		return CreateResult{}, logging.Reject(s.log, ErrAIUnavailable, "create skill request rejected: ai pool unavailable")
	}

	similarJSON, _ := json.Marshal(similar)
	created, err := s.repo.CreateRequest(ctx, sqlc.CreateSkillCreationRequestParams{
		RequesterID:     user.ID,
		Name:            name,
		SlugCandidate:   slug,
		Description:     pgtype.Text{String: desc, Valid: desc != ""},
		Status:          sqlc.CatalogRequestStatusGENERATING,
		SimilarSkills:   similarJSON,
		DraftMilestones: []byte("[]"),
	})
	if err != nil {
		return CreateResult{}, logging.Unexpected(s.log, err, "create skill request insert failed")
	}

	if err := s.enqueue(created.ID); err != nil {
		note := "AI queue is full; please retry"
		if _, ferr := s.repo.FailGenerating(ctx, created.ID, note); ferr != nil {
			s.log.Error().Err(ferr).Int64("request_id", created.ID).Msg("fail skill request after queue full failed")
		}
		return CreateResult{}, logging.Reject(s.log, ErrAIBusy, "create skill request rejected: ai queue full")
	}

	s.log.Info().
		Str("request_id", created.PublicID.String()).
		Str("user", user.Username).
		Str("name", name).
		Msg("skill creation request accepted for async generation")

	return CreateResult{Request: toView(created)}, nil
}

// RequeueGenerating re-enqueues rows left in GENERATING after a process restart.
func (s *Service) RequeueGenerating(ctx context.Context) {
	if s.pool == nil || s.generator == nil || !s.generator.Available() {
		return
	}
	ids, err := s.repo.ListGeneratingIDs(ctx)
	if err != nil {
		s.log.Error().Err(err).Msg("list generating skill requests failed")
		return
	}
	for _, id := range ids {
		if err := s.enqueue(id); err != nil {
			s.log.Warn().Err(err).Int64("id", id).Msg("requeue skill request failed")
		}
	}
	if len(ids) > 0 {
		s.log.Info().Int("count", len(ids)).Msg("requeued generating skill requests")
	}
}

func (s *Service) enqueue(id int64) error {
	return s.pool.Submit(func(ctx context.Context) {
		s.generateDraft(ctx, id)
	})
}

func (s *Service) generateDraft(ctx context.Context, id int64) {
	req, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			s.log.Error().Err(err).Int64("id", id).Msg("skill request generate: get failed")
		}
		return
	}
	if req.Status != sqlc.CatalogRequestStatusGENERATING {
		return
	}

	desc := ""
	if req.Description.Valid {
		desc = req.Description.String
	}

	eval, err := s.generator.EvaluateSkillRequest(ctx, req.Name, desc)
	if err != nil {
		s.log.Warn().Err(err).Int64("id", id).Str("name", req.Name).Msg("skill request evaluation failed")
		note := "AI review failed; please submit again"
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			note = "AI review timed out; please submit again"
		}
		if _, ferr := s.repo.FailGenerating(context.WithoutCancel(ctx), id, note); ferr != nil && !errors.Is(ferr, pgx.ErrNoRows) {
			s.log.Error().Err(ferr).Int64("id", id).Msg("mark skill request failed after eval error")
		}
		if s.notifier != nil {
			s.notifier.NotifySkillRequestFailed(context.WithoutCancel(ctx), req.RequesterID, req.Name, note)
		}
		return
	}
	if !eval.WorthConsidering {
		reason := eval.Reason
		if reason == "" {
			reason = "This request does not look like a skill worth building a Taggy roadmap for."
		}
		updated, rerr := s.repo.AutoRejectGenerating(context.WithoutCancel(ctx), id, reason)
		if rerr != nil && !errors.Is(rerr, pgx.ErrNoRows) {
			s.log.Error().Err(rerr).Int64("id", id).Msg("auto-reject skill request failed")
			return
		}
		if s.notifier != nil && rerr == nil {
			s.notifier.NotifySkillRequestRejected(context.WithoutCancel(ctx), updated.RequesterID, updated.Name, reason)
		}
		s.log.Info().
			Str("request_id", req.PublicID.String()).
			Str("name", req.Name).
			Str("reason", reason).
			Msg("skill creation request auto-rejected")
		return
	}

	drafts, err := s.generator.GenerateRoadmap(ctx, req.Name, desc, "", "")
	if err != nil {
		s.log.Warn().Err(err).Int64("id", id).Str("name", req.Name).Msg("skill request ai generation failed")
		note := "AI generation failed; please submit again"
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			note = "AI generation timed out; please submit again"
		}
		if _, ferr := s.repo.FailGenerating(context.WithoutCancel(ctx), id, note); ferr != nil && !errors.Is(ferr, pgx.ErrNoRows) {
			s.log.Error().Err(ferr).Int64("id", id).Msg("mark skill request failed")
		}
		if s.notifier != nil {
			s.notifier.NotifySkillRequestFailed(context.WithoutCancel(ctx), req.RequesterID, req.Name, note)
		}
		return
	}

	draftJSON, _ := json.Marshal(drafts)
	updated, err := s.repo.CompleteDraft(ctx, id, draftJSON)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return
		}
		s.log.Error().Err(err).Int64("id", id).Msg("complete skill request draft failed")
		return
	}

	bg := context.WithoutCancel(ctx)
	// Auto-publish the skill; the APPROVED row stays in the admin queue for audit.
	if _, aerr := s.approvePending(bg, updated, 0, "Auto-approved after AI generation"); aerr != nil {
		s.log.Error().Err(aerr).Int64("id", id).Msg("auto-approve skill request failed; left pending for admin")
		if s.notifier != nil {
			s.notifier.NotifySkillRequestReady(bg, updated.RequesterID, updated.Name)
		}
		return
	}
	s.log.Info().
		Str("request_id", updated.PublicID.String()).
		Int("milestones", len(drafts)).
		Msg("skill creation request auto-approved")
}

func (s *Service) ListMine(ctx context.Context, userPublicID uuid.UUID) ([]RequestView, error) {
	user, err := s.repo.GetUserByPublicID(ctx, userPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, logging.Unexpected(s.log, err, "list skill requests: get user failed")
	}
	rows, err := s.repo.ListByRequester(ctx, user.ID, defaultListLimit)
	if err != nil {
		return nil, logging.Unexpected(s.log, err, "list skill requests failed")
	}
	out := make([]RequestView, 0, len(rows))
	for _, row := range rows {
		out = append(out, toView(row))
	}
	return out, nil
}

func (s *Service) Cancel(ctx context.Context, userPublicID uuid.UUID, requestPublicID uuid.UUID) (RequestView, error) {
	user, err := s.repo.GetUserByPublicID(ctx, userPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return RequestView{}, apperrors.ErrNotFound
		}
		return RequestView{}, logging.Unexpected(s.log, err, "cancel skill request: get user failed")
	}
	row, err := s.repo.Cancel(ctx, requestPublicID, user.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return RequestView{}, logging.Reject(s.log, ErrRequestNotFound, "cancel skill request rejected: not found or not cancellable")
		}
		return RequestView{}, logging.Unexpected(s.log, err, "cancel skill request failed")
	}
	return toView(row), nil
}

func (s *Service) ListPendingAdmin(ctx context.Context) ([]RequestView, error) {
	rows, err := s.repo.ListAdmin(ctx, defaultListLimit)
	if err != nil {
		return nil, logging.Unexpected(s.log, err, "list admin skill requests failed")
	}
	out := make([]RequestView, 0, len(rows))
	for _, row := range rows {
		out = append(out, toView(row))
	}
	return out, nil
}

func (s *Service) Approve(ctx context.Context, adminPublicID, requestPublicID uuid.UUID) (RequestView, error) {
	admin, err := s.requireAdmin(ctx, adminPublicID)
	if err != nil {
		return RequestView{}, err
	}
	req, err := s.repo.GetByPublicID(ctx, requestPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return RequestView{}, ErrRequestNotFound
		}
		return RequestView{}, logging.Unexpected(s.log, err, "approve skill request: get failed")
	}
	if req.Status != sqlc.CatalogRequestStatusPENDING {
		return RequestView{}, logging.Reject(s.log, ErrNotPending, "approve skill request rejected: not pending")
	}

	view, err := s.approvePending(ctx, req, admin.ID, "")
	if err != nil {
		return RequestView{}, err
	}

	s.log.Info().
		Str("request_id", requestPublicID.String()).
		Str("admin", admin.Username).
		Msg("skill creation request approved")

	return view, nil
}

func (s *Service) approvePending(ctx context.Context, req sqlc.SkillCreationRequest, reviewerID int64, adminNote string) (RequestView, error) {
	var drafts []MilestoneDraft
	if err := json.Unmarshal(req.DraftMilestones, &drafts); err != nil || len(drafts) == 0 {
		return RequestView{}, logging.Reject(s.log, ErrAIFailed, "approve skill request rejected: invalid draft")
	}

	slug := req.SlugCandidate
	if _, err := s.repo.GetSkillBySlug(ctx, slug); err == nil {
		slug = fmt.Sprintf("%s-%d", slug, time.Now().Unix()%100000)
	} else if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return RequestView{}, logging.Unexpected(s.log, err, "approve skill request: slug check failed")
	}

	var notePtr *string
	if strings.TrimSpace(adminNote) != "" {
		n := strings.TrimSpace(adminNote)
		notePtr = &n
	}

	var skillID int64
	err := s.repo.WithTx(ctx, func(q *sqlc.Queries) error {
		skill, err := q.CreateSkill(ctx, sqlc.CreateSkillParams{
			Name:        req.Name,
			Slug:        slug,
			Description: req.Description,
		})
		if err != nil {
			return err
		}
		skillID = skill.ID

		roadmap, err := q.CreateRoadmap(ctx, skill.ID)
		if err != nil {
			return err
		}

		now := pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
		version, err := q.CreateRoadmapVersion(ctx, sqlc.CreateRoadmapVersionParams{
			RoadmapID:     pgtype.Int8{Int64: roadmap.ID, Valid: true},
			VersionNumber: 1,
			Status:        sqlc.CurrentStatusACTIVE,
			GeneratedBy:   "AI",
			PublishedAt:   now,
		})
		if err != nil {
			return err
		}
		if err := q.SetRoadmapCurrentVersion(ctx, sqlc.SetRoadmapCurrentVersionParams{
			ID:               roadmap.ID,
			CurrentVersionID: pgtype.Int8{Int64: version.ID, Valid: true},
		}); err != nil {
			return err
		}

		for _, m := range drafts {
			hours := int32(m.EstimatedHours)
			kind := m.Kind
			if kind == "" {
				kind = "TOPIC"
			}
			chapter := m.Chapter
			_, err := q.CreateMilestone(ctx, sqlc.CreateMilestoneParams{
				RoadmapVersionID: pgtype.Int8{Int64: version.ID, Valid: true},
				Title:            m.Title,
				Description:      pgtype.Text{String: m.Description, Valid: m.Description != ""},
				EstimatedHours:   pgtype.Int4{Int32: hours, Valid: true},
				OrderIndex:       int32(m.OrderIndex),
				Difficulty:       pgtype.Text{String: m.Difficulty, Valid: true},
				Slug:             m.Slug,
				Chapter:          pgtype.Text{String: chapter, Valid: chapter != ""},
				Kind:             kind,
			})
			if err != nil {
				return err
			}
		}

		communityName := req.Name + " Community"
		community, err := q.CreateCommunity(ctx, sqlc.CreateCommunityParams{
			SkillID:     skill.ID,
			Name:        communityName,
			Description: pgtype.Text{String: "Official community for " + req.Name + " learners on Taggy.", Valid: true},
		})
		if err != nil {
			return err
		}

		channels := []struct{ name, slug, desc string }{
			{"General", "general", "General discussion."},
			{"Resources", "resources", "Share tutorials and resources."},
			{"Projects", "projects", "Show builds and get feedback."},
		}
		for _, ch := range channels {
			if _, err := q.CreateCommunityChannel(ctx, sqlc.CreateCommunityChannelParams{
				CommunityID: community.ID,
				Name:        ch.name,
				Slug:        ch.slug,
				Description: pgtype.Text{String: ch.desc, Valid: true},
			}); err != nil {
				return err
			}
		}

		reviewedBy := pgtype.Int8{}
		if reviewerID > 0 {
			reviewedBy = pgtype.Int8{Int64: reviewerID, Valid: true}
		}
		_, err = q.ApproveSkillCreationRequest(ctx, sqlc.ApproveSkillCreationRequestParams{
			ID:             req.ID,
			ReviewedBy:     reviewedBy,
			CreatedSkillID: pgtype.Int8{Int64: skill.ID, Valid: true},
			AdminNote:      textPtr(notePtr),
		})
		return err
	})
	if err != nil {
		return RequestView{}, logging.Unexpected(s.log, err, "approve skill request transaction failed")
	}

	updated, err := s.repo.GetByPublicID(ctx, req.PublicID)
	if err != nil {
		return RequestView{}, logging.Unexpected(s.log, err, "approve skill request reload failed")
	}

	if s.notifier != nil {
		s.notifier.NotifySkillRequestApproved(ctx, req.RequesterID, skillID, slug)
	}

	return toView(updated), nil
}

func (s *Service) Reject(ctx context.Context, adminPublicID, requestPublicID uuid.UUID, note string) (RequestView, error) {
	admin, err := s.requireAdmin(ctx, adminPublicID)
	if err != nil {
		return RequestView{}, err
	}
	req, err := s.repo.GetByPublicID(ctx, requestPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return RequestView{}, ErrRequestNotFound
		}
		return RequestView{}, logging.Unexpected(s.log, err, "reject skill request: get failed")
	}
	if req.Status != sqlc.CatalogRequestStatusPENDING {
		return RequestView{}, logging.Reject(s.log, ErrNotPending, "reject skill request rejected: not pending")
	}

	note = strings.TrimSpace(note)
	var notePtr *string
	if note != "" {
		notePtr = &note
	}
	updated, err := s.repo.Reject(ctx, req.ID, admin.ID, notePtr)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return RequestView{}, ErrNotPending
		}
		return RequestView{}, logging.Unexpected(s.log, err, "reject skill request failed")
	}

	if s.notifier != nil {
		s.notifier.NotifySkillRequestRejected(ctx, req.RequesterID, req.Name, note)
	}

	s.log.Info().
		Str("request_id", requestPublicID.String()).
		Str("admin", admin.Username).
		Msg("skill creation request rejected")

	return toView(updated), nil
}

func (s *Service) requireAdmin(ctx context.Context, publicID uuid.UUID) (sqlc.User, error) {
	user, err := s.repo.GetUserByPublicID(ctx, publicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.User{}, apperrors.ErrNotFound
		}
		return sqlc.User{}, logging.Unexpected(s.log, err, "admin check: get user failed")
	}
	if !user.IsAdmin() {
		return sqlc.User{}, logging.Reject(s.log, ErrNotAdmin, "admin check rejected")
	}
	return user, nil
}

func mapSimilar(rows []sqlc.ListSimilarSkillsRow) []SimilarSkill {
	out := make([]SimilarSkill, 0, len(rows))
	for _, r := range rows {
		item := SimilarSkill{
			ID:    r.ID,
			Name:  r.Name,
			Slug:  r.Slug,
			Score: r.Score,
		}
		if r.Description.Valid {
			item.Description = &r.Description.String
		}
		out = append(out, item)
	}
	return out
}

func toView(row sqlc.SkillCreationRequest) RequestView {
	view := RequestView{
		ID:            row.PublicID.String(),
		Name:          row.Name,
		SlugCandidate: row.SlugCandidate,
		Status:        string(row.Status),
		CreatedAt:     formatTime(row.CreatedAt),
		UpdatedAt:     formatTime(row.UpdatedAt),
	}
	if row.Description.Valid {
		view.Description = &row.Description.String
	}
	if row.AdminNote.Valid {
		view.AdminNote = &row.AdminNote.String
	}
	if row.CreatedSkillID.Valid {
		view.CreatedSkillID = &row.CreatedSkillID.Int64
	}
	_ = json.Unmarshal(row.SimilarSkills, &view.SimilarSkills)
	if view.SimilarSkills == nil {
		view.SimilarSkills = []SimilarSkill{}
	}
	_ = json.Unmarshal(row.DraftMilestones, &view.DraftMilestones)
	if view.DraftMilestones == nil {
		view.DraftMilestones = []MilestoneDraft{}
	}
	return view
}

func formatTime(t pgtype.Timestamptz) string {
	if !t.Valid {
		return ""
	}
	return t.Time.UTC().Format(time.RFC3339)
}

func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	prev := false
	for _, r := range s {
		ok := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		if ok {
			b.WriteRune(r)
			prev = false
			continue
		}
		if !prev && b.Len() > 0 {
			b.WriteByte('-')
			prev = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if out == "" {
		return "skill"
	}
	if len(out) > 60 {
		out = strings.Trim(out[:60], "-")
	}
	return out
}
