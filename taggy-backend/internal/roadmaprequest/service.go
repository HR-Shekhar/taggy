package roadmaprequest

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

const defaultListLimit int32 = 50

type Generator interface {
	Available() bool
	GenerateRoadmap(ctx context.Context, skillName, description, rationale, currentOutline string) ([]openrouter.MilestoneDraft, error)
}

type Notifier interface {
	NotifyRoadmapRequestApproved(ctx context.Context, userID, skillID int64, skillSlug string, versionNumber int32)
	NotifyRoadmapRequestRejected(ctx context.Context, userID int64, skillSlug, note string)
	NotifyRoadmapUpdated(ctx context.Context, userID, skillID int64, skillSlug string, versionNumber int32)
	NotifyRoadmapRequestReady(ctx context.Context, userID, skillID int64, skillSlug string)
	NotifyRoadmapRequestFailed(ctx context.Context, userID int64, skillSlug, note string)
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

func (s *Service) Create(ctx context.Context, userPublicID uuid.UUID, skillSlug string, input CreateInput) (RequestView, error) {
	user, err := s.repo.GetUserByPublicID(ctx, userPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return RequestView{}, apperrors.ErrNotFound
		}
		return RequestView{}, logging.Unexpected(s.log, err, "create roadmap request: get user failed")
	}

	skill, err := s.repo.GetSkillBySlug(ctx, skillSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return RequestView{}, logging.Reject(s.log, ErrSkillNotFound, "create roadmap request rejected: skill not found")
		}
		return RequestView{}, logging.Unexpected(s.log, err, "create roadmap request: get skill failed")
	}

	if _, err := s.repo.GetUserSkill(ctx, user.ID, skill.Slug); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return RequestView{}, logging.Reject(s.log, ErrNotEnrolled, "create roadmap request rejected: not enrolled")
		}
		return RequestView{}, logging.Unexpected(s.log, err, "create roadmap request: enrollment check failed")
	}

	if _, err := s.repo.GetPending(ctx, user.ID, skill.ID); err == nil {
		return RequestView{}, logging.Reject(s.log, ErrDuplicatePending, "create roadmap request rejected: duplicate pending")
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return RequestView{}, logging.Unexpected(s.log, err, "create roadmap request: pending check failed")
	}

	active, err := s.repo.GetActiveVersion(ctx, skill.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return RequestView{}, logging.Reject(s.log, ErrNoActiveVersion, "create roadmap request rejected: no active version")
		}
		return RequestView{}, logging.Unexpected(s.log, err, "create roadmap request: active version failed")
	}

	if s.generator == nil || !s.generator.Available() {
		return RequestView{}, logging.Reject(s.log, ErrAIUnavailable, "create roadmap request rejected: ai unavailable")
	}
	if s.pool == nil {
		return RequestView{}, logging.Reject(s.log, ErrAIUnavailable, "create roadmap request rejected: ai pool unavailable")
	}

	rationale := strings.TrimSpace(input.Rationale)
	created, err := s.repo.Create(ctx, sqlc.CreateRoadmapEditRequestParams{
		SkillID:           skill.ID,
		RequesterID:       user.ID,
		Rationale:         pgtype.Text{String: rationale, Valid: rationale != ""},
		Status:            sqlc.CatalogRequestStatusGENERATING,
		BaseVersionNumber: active.VersionNumber,
		DraftMilestones:   []byte("[]"),
	})
	if err != nil {
		return RequestView{}, logging.Unexpected(s.log, err, "create roadmap request insert failed")
	}

	if err := s.enqueue(created.ID); err != nil {
		note := "AI queue is full; please retry"
		if _, ferr := s.repo.FailGenerating(ctx, created.ID, note); ferr != nil {
			s.log.Error().Err(ferr).Int64("request_id", created.ID).Msg("fail roadmap request after queue full failed")
		}
		return RequestView{}, logging.Reject(s.log, ErrAIBusy, "create roadmap request rejected: ai queue full")
	}

	s.log.Info().
		Str("request_id", created.PublicID.String()).
		Str("skill", skillSlug).
		Str("user", user.Username).
		Msg("roadmap edit request accepted for async generation")

	row, err := s.repo.GetByPublicID(ctx, created.PublicID)
	if err != nil {
		return RequestView{}, logging.Unexpected(s.log, err, "create roadmap request reload failed")
	}
	return toViewFromGet(row), nil
}

// RequeueGenerating re-enqueues GENERATING rows after restart.
func (s *Service) RequeueGenerating(ctx context.Context) {
	if s.pool == nil || s.generator == nil || !s.generator.Available() {
		return
	}
	ids, err := s.repo.ListGeneratingIDs(ctx)
	if err != nil {
		s.log.Error().Err(err).Msg("list generating roadmap requests failed")
		return
	}
	for _, id := range ids {
		if err := s.enqueue(id); err != nil {
			s.log.Warn().Err(err).Int64("id", id).Msg("requeue roadmap request failed")
		}
	}
	if len(ids) > 0 {
		s.log.Info().Int("count", len(ids)).Msg("requeued generating roadmap requests")
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
			s.log.Error().Err(err).Int64("id", id).Msg("roadmap request generate: get failed")
		}
		return
	}
	if req.Status != sqlc.CatalogRequestStatusGENERATING {
		return
	}

	skill, err := s.repo.GetSkillBySlug(ctx, req.SkillSlug)
	if err != nil {
		s.log.Error().Err(err).Int64("id", id).Msg("roadmap request generate: skill lookup failed")
		s.failRequest(ctx, id, req.RequesterID, req.SkillSlug, "AI generation failed; please submit again")
		return
	}
	desc := ""
	if skill.Description.Valid {
		desc = skill.Description.String
	}
	rationale := ""
	if req.Rationale.Valid {
		rationale = req.Rationale.String
	}

	currentOutline := ""
	if active, aerr := s.repo.GetActiveVersion(ctx, skill.ID); aerr == nil {
		if ms, merr := s.repo.ListMilestonesByVersion(ctx, active.ID); merr == nil && len(ms) > 0 {
			var b strings.Builder
			for _, m := range ms {
				kind := m.Kind
				if kind == "" {
					kind = "TOPIC"
				}
				hours := 0
				if m.EstimatedHours.Valid {
					hours = int(m.EstimatedHours.Int32)
				}
				b.WriteString(fmt.Sprintf("- [%s] %s (%dh)", kind, m.Title, hours))
				if m.Description.Valid && m.Description.String != "" {
					b.WriteString(": ")
					b.WriteString(m.Description.String)
				}
				b.WriteByte('\n')
			}
			currentOutline = b.String()
		}
	}

	drafts, err := s.generator.GenerateRoadmap(ctx, skill.Name, desc, rationale, currentOutline)
	if err != nil {
		s.log.Warn().Err(err).Int64("id", id).Str("skill", req.SkillSlug).Msg("roadmap edit ai generation failed")
		note := "AI generation failed; please submit again"
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			note = "AI generation timed out; please submit again"
		}
		s.failRequest(ctx, id, req.RequesterID, req.SkillSlug, note)
		return
	}

	draftJSON, _ := json.Marshal(drafts)
	updated, err := s.repo.CompleteDraft(ctx, id, draftJSON)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return
		}
		s.log.Error().Err(err).Int64("id", id).Msg("complete roadmap request draft failed")
		return
	}

	if s.notifier != nil {
		s.notifier.NotifyRoadmapRequestReady(context.WithoutCancel(ctx), updated.RequesterID, updated.SkillID, req.SkillSlug)
	}
	s.log.Info().
		Str("request_id", updated.PublicID.String()).
		Str("skill", req.SkillSlug).
		Int("milestones", len(drafts)).
		Msg("roadmap edit draft ready for admin review")
}

func (s *Service) failRequest(ctx context.Context, id, requesterID int64, skillSlug, note string) {
	bg := context.WithoutCancel(ctx)
	if _, ferr := s.repo.FailGenerating(bg, id, note); ferr != nil && !errors.Is(ferr, pgx.ErrNoRows) {
		s.log.Error().Err(ferr).Int64("id", id).Msg("mark roadmap request failed")
	}
	if s.notifier != nil {
		s.notifier.NotifyRoadmapRequestFailed(bg, requesterID, skillSlug, note)
	}
}

func (s *Service) ListMine(ctx context.Context, userPublicID uuid.UUID) ([]RequestView, error) {
	user, err := s.repo.GetUserByPublicID(ctx, userPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, logging.Unexpected(s.log, err, "list roadmap requests: get user failed")
	}
	rows, err := s.repo.ListByRequester(ctx, user.ID, defaultListLimit)
	if err != nil {
		return nil, logging.Unexpected(s.log, err, "list roadmap requests failed")
	}
	out := make([]RequestView, 0, len(rows))
	for _, row := range rows {
		out = append(out, toViewFromList(row))
	}
	return out, nil
}

func (s *Service) Cancel(ctx context.Context, userPublicID, requestPublicID uuid.UUID) (RequestView, error) {
	user, err := s.repo.GetUserByPublicID(ctx, userPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return RequestView{}, apperrors.ErrNotFound
		}
		return RequestView{}, logging.Unexpected(s.log, err, "cancel roadmap request: get user failed")
	}
	_, err = s.repo.Cancel(ctx, requestPublicID, user.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return RequestView{}, logging.Reject(s.log, ErrRequestNotFound, "cancel roadmap request rejected")
		}
		return RequestView{}, logging.Unexpected(s.log, err, "cancel roadmap request failed")
	}
	row, err := s.repo.GetByPublicID(ctx, requestPublicID)
	if err != nil {
		return RequestView{}, logging.Unexpected(s.log, err, "cancel roadmap request reload failed")
	}
	return toViewFromGet(row), nil
}

func (s *Service) ListPendingAdmin(ctx context.Context) ([]RequestView, error) {
	rows, err := s.repo.ListPending(ctx, defaultListLimit)
	if err != nil {
		return nil, logging.Unexpected(s.log, err, "list pending roadmap requests failed")
	}
	out := make([]RequestView, 0, len(rows))
	for _, row := range rows {
		out = append(out, toViewFromPending(row))
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
		return RequestView{}, logging.Unexpected(s.log, err, "approve roadmap request: get failed")
	}
	if req.Status != sqlc.CatalogRequestStatusPENDING {
		return RequestView{}, logging.Reject(s.log, ErrNotPending, "approve roadmap request rejected: not pending")
	}

	var drafts []MilestoneDraft
	if err := json.Unmarshal(req.DraftMilestones, &drafts); err != nil || len(drafts) == 0 {
		return RequestView{}, logging.Reject(s.log, ErrAIFailed, "approve roadmap request rejected: invalid draft")
	}

	active, err := s.repo.GetActiveVersion(ctx, req.SkillID)
	if err != nil {
		return RequestView{}, logging.Unexpected(s.log, err, "approve roadmap request: active version failed")
	}
	if !active.RoadmapID.Valid {
		return RequestView{}, logging.Reject(s.log, ErrNoActiveVersion, "approve roadmap request rejected: missing roadmap")
	}

	maxVer, err := s.repo.GetMaxVersion(ctx, active.RoadmapID.Int64)
	if err != nil {
		return RequestView{}, logging.Unexpected(s.log, err, "approve roadmap request: max version failed")
	}
	nextVer := maxVer + 1

	var versionID int64
	err = s.repo.WithTx(ctx, func(q *sqlc.Queries) error {
		if err := q.ArchiveActiveRoadmapVersions(ctx, active.RoadmapID); err != nil {
			return err
		}
		now := pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
		version, err := q.CreateRoadmapVersion(ctx, sqlc.CreateRoadmapVersionParams{
			RoadmapID:     active.RoadmapID,
			VersionNumber: nextVer,
			Status:        sqlc.CurrentStatusACTIVE,
			GeneratedBy:   "AI",
			PublishedAt:   now,
		})
		if err != nil {
			return err
		}
		versionID = version.ID
		if err := q.SetRoadmapCurrentVersion(ctx, sqlc.SetRoadmapCurrentVersionParams{
			ID:               active.RoadmapID.Int64,
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
			if _, err := q.CreateMilestone(ctx, sqlc.CreateMilestoneParams{
				RoadmapVersionID: pgtype.Int8{Int64: version.ID, Valid: true},
				Title:            m.Title,
				Description:      pgtype.Text{String: m.Description, Valid: m.Description != ""},
				EstimatedHours:   pgtype.Int4{Int32: hours, Valid: true},
				OrderIndex:       int32(m.OrderIndex),
				Difficulty:       pgtype.Text{String: m.Difficulty, Valid: true},
				Slug:             m.Slug,
				Chapter:          pgtype.Text{String: chapter, Valid: chapter != ""},
				Kind:             kind,
			}); err != nil {
				return err
			}
		}
		_, err = q.ApproveRoadmapEditRequest(ctx, sqlc.ApproveRoadmapEditRequestParams{
			ID:               req.ID,
			ReviewedBy:       pgtype.Int8{Int64: admin.ID, Valid: true},
			CreatedVersionID: pgtype.Int8{Int64: version.ID, Valid: true},
			AdminNote:        pgtype.Text{},
		})
		return err
	})
	if err != nil {
		return RequestView{}, logging.Unexpected(s.log, err, "approve roadmap request transaction failed")
	}

	updated, err := s.repo.GetByPublicID(ctx, requestPublicID)
	if err != nil {
		return RequestView{}, logging.Unexpected(s.log, err, "approve roadmap request reload failed")
	}

	if s.notifier != nil {
		s.notifier.NotifyRoadmapRequestApproved(ctx, req.RequesterID, req.SkillID, req.SkillSlug, nextVer)
		userIDs, err := s.repo.ListEnrolledUserIDs(ctx, req.SkillID)
		if err == nil {
			for _, uid := range userIDs {
				s.notifier.NotifyRoadmapUpdated(ctx, uid, req.SkillID, req.SkillSlug, nextVer)
			}
		}
	}

	s.log.Info().
		Str("request_id", requestPublicID.String()).
		Str("admin", admin.Username).
		Str("skill", req.SkillSlug).
		Int32("version", nextVer).
		Int64("version_id", versionID).
		Msg("roadmap edit request approved")

	return toViewFromGet(updated), nil
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
		return RequestView{}, logging.Unexpected(s.log, err, "reject roadmap request: get failed")
	}
	if req.Status != sqlc.CatalogRequestStatusPENDING {
		return RequestView{}, logging.Reject(s.log, ErrNotPending, "reject roadmap request rejected: not pending")
	}
	note = strings.TrimSpace(note)
	var notePtr *string
	if note != "" {
		notePtr = &note
	}
	if _, err := s.repo.Reject(ctx, req.ID, admin.ID, notePtr); err != nil {
		return RequestView{}, logging.Unexpected(s.log, err, "reject roadmap request failed")
	}
	updated, err := s.repo.GetByPublicID(ctx, requestPublicID)
	if err != nil {
		return RequestView{}, logging.Unexpected(s.log, err, "reject roadmap request reload failed")
	}
	if s.notifier != nil {
		s.notifier.NotifyRoadmapRequestRejected(ctx, req.RequesterID, req.SkillSlug, note)
	}
	s.log.Info().Str("request_id", requestPublicID.String()).Str("admin", admin.Username).Msg("roadmap edit request rejected")
	return toViewFromGet(updated), nil
}

func (s *Service) requireAdmin(ctx context.Context, publicID uuid.UUID) (sqlc.User, error) {
	user, err := s.repo.GetUserByPublicID(ctx, publicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.User{}, apperrors.ErrNotFound
		}
		return sqlc.User{}, logging.Unexpected(s.log, err, "admin check failed")
	}
	if !user.IsAdmin() {
		return sqlc.User{}, logging.Reject(s.log, ErrNotAdmin, "admin check rejected")
	}
	return user, nil
}

func toViewFromGet(row sqlc.GetRoadmapEditRequestByPublicIDRow) RequestView {
	view := RequestView{
		ID:                row.PublicID.String(),
		SkillSlug:         row.SkillSlug,
		SkillName:         row.SkillName,
		Status:            string(row.Status),
		BaseVersionNumber: row.BaseVersionNumber,
		CreatedAt:         formatTime(row.CreatedAt),
		UpdatedAt:         formatTime(row.UpdatedAt),
	}
	if row.Rationale.Valid {
		view.Rationale = &row.Rationale.String
	}
	if row.AdminNote.Valid {
		view.AdminNote = &row.AdminNote.String
	}
	if row.CreatedVersionID.Valid {
		view.CreatedVersionID = &row.CreatedVersionID.Int64
	}
	_ = json.Unmarshal(row.DraftMilestones, &view.DraftMilestones)
	if view.DraftMilestones == nil {
		view.DraftMilestones = []MilestoneDraft{}
	}
	return view
}

func toViewFromList(row sqlc.ListRoadmapEditRequestsByRequesterRow) RequestView {
	return toViewFromGet(sqlc.GetRoadmapEditRequestByPublicIDRow{
		ID:                row.ID,
		PublicID:          row.PublicID,
		SkillID:           row.SkillID,
		RequesterID:       row.RequesterID,
		Rationale:         row.Rationale,
		Status:            row.Status,
		BaseVersionNumber: row.BaseVersionNumber,
		DraftMilestones:   row.DraftMilestones,
		AdminNote:         row.AdminNote,
		ReviewedBy:        row.ReviewedBy,
		ReviewedAt:        row.ReviewedAt,
		CreatedVersionID:  row.CreatedVersionID,
		CreatedAt:         row.CreatedAt,
		UpdatedAt:         row.UpdatedAt,
		SkillSlug:         row.SkillSlug,
		SkillName:         row.SkillName,
	})
}

func toViewFromPending(row sqlc.ListPendingRoadmapEditRequestsRow) RequestView {
	return toViewFromGet(sqlc.GetRoadmapEditRequestByPublicIDRow{
		ID:                row.ID,
		PublicID:          row.PublicID,
		SkillID:           row.SkillID,
		RequesterID:       row.RequesterID,
		Rationale:         row.Rationale,
		Status:            row.Status,
		BaseVersionNumber: row.BaseVersionNumber,
		DraftMilestones:   row.DraftMilestones,
		AdminNote:         row.AdminNote,
		ReviewedBy:        row.ReviewedBy,
		ReviewedAt:        row.ReviewedAt,
		CreatedVersionID:  row.CreatedVersionID,
		CreatedAt:         row.CreatedAt,
		UpdatedAt:         row.UpdatedAt,
		SkillSlug:         row.SkillSlug,
		SkillName:         row.SkillName,
	})
}

func formatTime(t pgtype.Timestamptz) string {
	if !t.Valid {
		return ""
	}
	return t.Time.UTC().Format(time.RFC3339)
}
