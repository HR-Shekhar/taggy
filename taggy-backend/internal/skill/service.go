package skill

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/postgres/sqlc"
	apperrors "github.com/HR-Shekhar/taggy-backend/internal/shared/errors"
	"github.com/HR-Shekhar/taggy-backend/internal/shared/logging"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
)

type Notifier interface {
	NotifyMilestoneCompleted(ctx context.Context, userID, milestoneID int64, skillSlug, milestoneTitle string)
	NotifyMilestoneDue(ctx context.Context, userID, milestoneID int64, skillSlug, milestoneTitle, dueAtRFC3339 string)
	NotifyRoadmapUpdated(ctx context.Context, userID, skillID int64, skillSlug string, versionNumber int32)
	NotifyCommunityAnnouncement(ctx context.Context, userID, communityID int64, skillSlug, body string)
}

type Service struct {
	repo     *Repository
	notifier Notifier
	log      zerolog.Logger
}

func NewService(repo *Repository, notifier Notifier, log zerolog.Logger) *Service {
	return &Service{repo: repo, notifier: notifier, log: log}
}

func (s *Service) ListSkills(ctx context.Context) ([]sqlc.Skill, error) {
	skills, err := s.repo.ListActiveSkills(ctx)
	if err != nil {
		return nil, logging.Unexpected(s.log, err, "list skills failed")
	}

	s.log.Debug().Int("count", len(skills)).Msg("skills listed")
	return skills, nil
}

func (s *Service) GetSkillBySlug(ctx context.Context, slug string) (sqlc.Skill, sqlc.GetCommunityBySkillIDRow, error) {
	skill, err := s.repo.GetSkillBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.Skill{}, sqlc.GetCommunityBySkillIDRow{}, ErrSkillNotFound
		}
		return sqlc.Skill{}, sqlc.GetCommunityBySkillIDRow{}, logging.Unexpected(s.log, err, "get skill by slug failed")
	}

	if !skill.IsActive {
		return sqlc.Skill{}, sqlc.GetCommunityBySkillIDRow{}, ErrSkillNotFound
	}

	community, err := s.repo.GetCommunityBySkillID(ctx, skill.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.Skill{}, sqlc.GetCommunityBySkillIDRow{}, ErrCommunityNotFound
		}
		return sqlc.Skill{}, sqlc.GetCommunityBySkillIDRow{}, logging.Unexpected(s.log, err, "get community by skill id failed")
	}

	return skill, community, nil
}

func (s *Service) JoinSkill(ctx context.Context, userPublicID uuid.UUID, skillSlug string) (sqlc.Userskill, sqlc.GetCommunityBySkillIDRow, sqlc.Skill, error) {
	user, err := s.repo.GetUserByPublicID(ctx, userPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.Userskill{}, sqlc.GetCommunityBySkillIDRow{}, sqlc.Skill{}, logging.Reject(s.log, apperrors.ErrNotFound, "join skill rejected: user not found")
		}
		return sqlc.Userskill{}, sqlc.GetCommunityBySkillIDRow{}, sqlc.Skill{}, logging.Unexpected(s.log, err, "get user by public id failed")
	}

	skill, err := s.repo.GetSkillBySlug(ctx, skillSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.Userskill{}, sqlc.GetCommunityBySkillIDRow{}, sqlc.Skill{}, logging.Reject(s.log, ErrSkillNotFound, "join skill rejected: skill not found")
		}
		return sqlc.Userskill{}, sqlc.GetCommunityBySkillIDRow{}, sqlc.Skill{}, logging.Unexpected(s.log, err, "get skill by slug failed")
	}

	if !skill.IsActive {
		return sqlc.Userskill{}, sqlc.GetCommunityBySkillIDRow{}, sqlc.Skill{}, logging.Reject(s.log, ErrSkillNotFound, "join skill rejected: skill inactive")
	}

	_, err = s.repo.GetUserSkillByUserAndSkillSlug(ctx, user.ID, skill.Slug)
	if err == nil {
		s.log.Warn().
			Str("user_id", user.PublicID.String()).
			Str("skill_slug", skill.Slug).
			Msg("join rejected: already enrolled")
		return sqlc.Userskill{}, sqlc.GetCommunityBySkillIDRow{}, sqlc.Skill{}, ErrAlreadyEnrolled
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return sqlc.Userskill{}, sqlc.GetCommunityBySkillIDRow{}, sqlc.Skill{}, logging.Unexpected(s.log, err, "get user skill failed")
	}

	if user.Subscription == sqlc.SubscriptionTierFREE {
		activeCount, err := s.repo.CountActiveUserSkillsByUserID(ctx, user.ID)
		if err != nil {
			return sqlc.Userskill{}, sqlc.GetCommunityBySkillIDRow{}, sqlc.Skill{}, logging.Unexpected(s.log, err, "count active user skills failed")
		}
		if activeCount >= 1 {
			s.log.Warn().
				Str("user_id", user.PublicID.String()).
				Msg("join rejected: free user active skill limit")
			return sqlc.Userskill{}, sqlc.GetCommunityBySkillIDRow{}, sqlc.Skill{}, ErrActiveSkillLimit
		}
	}

	roadmapVersion, err := s.repo.GetActiveRoadmapVersionBySkillID(ctx, skill.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.Userskill{}, sqlc.GetCommunityBySkillIDRow{}, sqlc.Skill{}, logging.Reject(s.log, ErrRoadmapNotFound, "join skill rejected: roadmap not found")
		}
		return sqlc.Userskill{}, sqlc.GetCommunityBySkillIDRow{}, sqlc.Skill{}, logging.Unexpected(s.log, err, "get active roadmap version failed")
	}

	milestones, err := s.repo.ListMilestonesByRoadmapVersionID(ctx, roadmapVersion.ID)
	if err != nil {
		return sqlc.Userskill{}, sqlc.GetCommunityBySkillIDRow{}, sqlc.Skill{}, logging.Unexpected(s.log, err, "list milestones for enroll failed")
	}

	milestoneIDs := make([]int64, 0, len(milestones))
	for _, m := range milestones {
		milestoneIDs = append(milestoneIDs, m.ID)
	}

	userSkill, err := s.repo.EnrollUserInSkill(ctx, user.ID, skill.ID, roadmapVersion.ID, milestoneIDs)
	if err != nil {
		return sqlc.Userskill{}, sqlc.GetCommunityBySkillIDRow{}, sqlc.Skill{}, logging.Unexpected(s.log, err, "enroll user in skill failed")
	}

	community, err := s.repo.GetCommunityBySkillID(ctx, skill.ID)
	if err != nil {
		return sqlc.Userskill{}, sqlc.GetCommunityBySkillIDRow{}, sqlc.Skill{}, logging.Unexpected(s.log, err, "get community after enroll failed")
	}

	s.log.Info().
		Str("user_id", user.PublicID.String()).
		Str("skill_slug", skill.Slug).
		Int64("user_skill_id", userSkill.ID).
		Msg("user joined skill")

	if s.notifier != nil {
		s.notifier.NotifyCommunityAnnouncement(
			ctx,
			user.ID,
			community.ID,
			skill.Slug,
			fmt.Sprintf("Welcome to the %s community. Open chat to meet other learners.", skill.Slug),
		)
	}

	return userSkill, community, skill, nil
}

func (s *Service) ListMySkills(ctx context.Context, userPublicID uuid.UUID) ([]sqlc.ListUserSkillsByUserIDRow, error) {
	user, err := s.repo.GetUserByPublicID(ctx, userPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, logging.Unexpected(s.log, err, "get user by public id failed")
	}

	rows, err := s.repo.ListUserSkillsByUserID(ctx, user.ID)
	if err != nil {
		return nil, logging.Unexpected(s.log, err, "list user skills failed")
	}
	return rows, nil
}

func (s *Service) SwitchRoadmapVersion(
	ctx context.Context,
	userPublicID uuid.UUID,
	skillSlug string,
	versionNumber int32,
) (sqlc.Userskill, sqlc.GetRoadmapVersionBySkillSlugAndNumberRow, error) {
	userSkill, err := s.getActiveUserSkillBySlug(ctx, userPublicID, skillSlug)
	if err != nil {
		return sqlc.Userskill{}, sqlc.GetRoadmapVersionBySkillSlugAndNumberRow{}, err
	}

	version, err := s.repo.GetRoadmapVersionBySkillSlugAndNumber(ctx, skillSlug, versionNumber)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.Userskill{}, sqlc.GetRoadmapVersionBySkillSlugAndNumberRow{}, logging.Reject(s.log, ErrVersionNotFound, "switch version rejected: not found")
		}
		return sqlc.Userskill{}, sqlc.GetRoadmapVersionBySkillSlugAndNumberRow{}, logging.Unexpected(s.log, err, "get roadmap version for switch failed")
	}

	if version.Status == sqlc.CurrentStatusDRAFT {
		return sqlc.Userskill{}, sqlc.GetRoadmapVersionBySkillSlugAndNumberRow{}, logging.Reject(s.log, ErrVersionNotSelectable, "switch version rejected: draft")
	}

	if userSkill.RoadmapVersionID == version.ID {
		return sqlc.Userskill{}, sqlc.GetRoadmapVersionBySkillSlugAndNumberRow{}, logging.Reject(s.log, ErrAlreadyOnVersion, "switch version rejected: already on version")
	}

	completedSlugs, err := s.repo.ListCompletedMilestoneSlugsByUserSkillID(ctx, userSkill.ID)
	if err != nil {
		return sqlc.Userskill{}, sqlc.GetRoadmapVersionBySkillSlugAndNumberRow{}, logging.Unexpected(s.log, err, "list completed milestone slugs failed")
	}
	completedSet := make(map[string]struct{}, len(completedSlugs))
	for _, slug := range completedSlugs {
		completedSet[slug] = struct{}{}
	}

	milestones, err := s.repo.ListMilestonesByRoadmapVersionID(ctx, version.ID)
	if err != nil {
		return sqlc.Userskill{}, sqlc.GetRoadmapVersionBySkillSlugAndNumberRow{}, logging.Unexpected(s.log, err, "list milestones for switch failed")
	}

	updated, err := s.repo.SwitchUserSkillRoadmapVersion(ctx, userSkill.ID, version.ID, milestones, completedSet)
	if err != nil {
		return sqlc.Userskill{}, sqlc.GetRoadmapVersionBySkillSlugAndNumberRow{}, logging.Unexpected(s.log, err, "switch roadmap version failed")
	}

	s.log.Info().
		Str("user_id", userPublicID.String()).
		Str("skill_slug", skillSlug).
		Int32("version_number", versionNumber).
		Msg("user switched roadmap version")

	if s.notifier != nil {
		s.notifier.NotifyRoadmapUpdated(ctx, userSkill.UserID, userSkill.SkillID, skillSlug, versionNumber)
	}

	return updated, version, nil
}

func (s *Service) ListMilestones(
	ctx context.Context,
	userPublicID uuid.UUID,
	skillSlug string,
) ([]sqlc.ListMilestoneProgressByUserSkillIDRow, error) {
	userSkill, err := s.getActiveUserSkillBySlug(ctx, userPublicID, skillSlug)
	if err != nil {
		return nil, err
	}

	rows, err := s.repo.ListMilestoneProgressByUserSkillID(ctx, userSkill.ID)
	if err != nil {
		return nil, logging.Unexpected(s.log, err, "list milestone progress failed")
	}
	return rows, nil
}

func (s *Service) UpdateMilestone(
	ctx context.Context,
	userPublicID uuid.UUID,
	skillSlug string,
	milestoneSlug string,
	input UpdateMilestoneInput,
) (sqlc.GetMilestoneProgressBySlugRow, error) {
	userSkill, err := s.getActiveUserSkillBySlug(ctx, userPublicID, skillSlug)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) || errors.Is(err, ErrUserSkillNotFound) {
			return sqlc.GetMilestoneProgressBySlugRow{}, logging.Reject(s.log, err, "update milestone rejected: user skill not found")
		}
		return sqlc.GetMilestoneProgressBySlugRow{}, err
	}

	progress, err := s.repo.GetMilestoneProgressBySlug(ctx, userSkill.ID, milestoneSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.GetMilestoneProgressBySlugRow{}, logging.Reject(s.log, ErrMilestoneNotFound, "update milestone rejected: milestone not found")
		}
		return sqlc.GetMilestoneProgressBySlugRow{}, logging.Unexpected(s.log, err, "get milestone progress failed")
	}

	if progress.Status == sqlc.ProgressStatusCOMPLETED {
		return sqlc.GetMilestoneProgressBySlugRow{}, logging.Reject(s.log, ErrMilestoneAlreadyComplete, "update milestone rejected: already complete")
	}

	switch input.Action {
	case MilestoneActionComplete:
		kind := progress.Kind
		if kind == "" {
			kind = "TOPIC"
		}

		if kind == "CHAPTER" {
			chapterName := ""
			if progress.Chapter.Valid {
				chapterName = progress.Chapter.String
			}
			if chapterName == "" {
				chapterName = progress.Title
			}
			incompleteSubs, err := s.repo.CountIncompleteTopicsInChapter(ctx, userSkill.ID, chapterName)
			if err != nil {
				return sqlc.GetMilestoneProgressBySlugRow{}, logging.Unexpected(s.log, err, "count incomplete chapter topics failed")
			}
			if incompleteSubs > 0 {
				return sqlc.GetMilestoneProgressBySlugRow{}, logging.Reject(s.log, ErrSubtopicsIncomplete, "complete chapter rejected: subtopics incomplete")
			}
			incompleteChapters, err := s.repo.CountIncompleteChaptersBefore(ctx, userSkill.ID, progress.OrderIndex)
			if err != nil {
				return sqlc.GetMilestoneProgressBySlugRow{}, logging.Unexpected(s.log, err, "count incomplete chapters failed")
			}
			if incompleteChapters > 0 {
				return sqlc.GetMilestoneProgressBySlugRow{}, logging.Reject(s.log, ErrMilestoneOutOfOrder, "complete chapter rejected: previous chapters incomplete")
			}
		} else {
			// Subtopics: only prior TOPIC milestones must be done (parent CHAPTER can wait).
			incompleteBefore, err := s.repo.CountIncompleteTopicMilestonesBefore(ctx, userSkill.ID, progress.OrderIndex)
			if err != nil {
				return sqlc.GetMilestoneProgressBySlugRow{}, logging.Unexpected(s.log, err, "count incomplete topic milestones failed")
			}
			if incompleteBefore > 0 {
				s.log.Warn().
					Int64("user_skill_id", userSkill.ID).
					Str("milestone_slug", milestoneSlug).
					Msg("milestone complete rejected: previous subtopics incomplete")
				return sqlc.GetMilestoneProgressBySlugRow{}, ErrMilestoneOutOfOrder
			}
		}

		_, err = s.repo.CompleteMilestoneProgress(ctx, progress.ID)
		if err != nil {
			return sqlc.GetMilestoneProgressBySlugRow{}, logging.Unexpected(s.log, err, "complete milestone progress failed")
		}

		s.log.Info().
			Int64("user_skill_id", userSkill.ID).
			Str("milestone_slug", milestoneSlug).
			Str("kind", kind).
			Msg("milestone completed")

		if s.notifier != nil && progress.MilestoneID.Valid {
			s.notifier.NotifyMilestoneCompleted(
				ctx,
				userSkill.UserID,
				progress.MilestoneID.Int64,
				skillSlug,
				progress.Title,
			)
		}

	case MilestoneActionPostpone:
		if input.PostponedUntil == nil {
			return sqlc.GetMilestoneProgressBySlugRow{}, logging.Reject(s.log, apperrors.ErrBadRequest, "postpone rejected: missing postponed_until")
		}
		if input.PostponedUntil.Before(time.Now()) {
			return sqlc.GetMilestoneProgressBySlugRow{}, logging.Reject(s.log, apperrors.ErrBadRequest, "postpone rejected: postponed_until in past")
		}

		_, err = s.repo.PostponeMilestoneProgress(ctx, progress.ID, pgtype.Timestamptz{
			Time:  *input.PostponedUntil,
			Valid: true,
		})
		if err != nil {
			return sqlc.GetMilestoneProgressBySlugRow{}, logging.Unexpected(s.log, err, "postpone milestone progress failed")
		}

		s.log.Info().
			Int64("user_skill_id", userSkill.ID).
			Str("milestone_slug", milestoneSlug).
			Time("postponed_until", *input.PostponedUntil).
			Msg("milestone postponed")

		if s.notifier != nil && progress.MilestoneID.Valid {
			s.notifier.NotifyMilestoneDue(
				ctx,
				userSkill.UserID,
				progress.MilestoneID.Int64,
				skillSlug,
				progress.Title,
				input.PostponedUntil.UTC().Format(time.RFC3339),
			)
		}

	default:
		return sqlc.GetMilestoneProgressBySlugRow{}, logging.Reject(s.log, ErrInvalidMilestoneAction, "update milestone rejected: invalid action")
	}

	updated, err := s.repo.GetMilestoneProgressBySlug(ctx, userSkill.ID, milestoneSlug)
	if err != nil {
		return sqlc.GetMilestoneProgressBySlugRow{}, logging.Unexpected(s.log, err, "reload milestone progress failed")
	}

	return updated, nil
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

func (s *Service) getActiveUserSkillBySlug(ctx context.Context, userPublicID uuid.UUID, skillSlug string) (sqlc.Userskill, error) {
	user, err := s.repo.GetUserByPublicID(ctx, userPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.Userskill{}, apperrors.ErrNotFound
		}
		return sqlc.Userskill{}, logging.Unexpected(s.log, err, "get user by public id failed")
	}

	userSkill, err := s.repo.GetUserSkillByUserAndSkillSlug(ctx, user.ID, skillSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.Userskill{}, ErrUserSkillNotFound
		}
		return sqlc.Userskill{}, logging.Unexpected(s.log, err, "get user skill by slug failed")
	}

	if userSkill.Status != sqlc.StatusValueACTIVE {
		return sqlc.Userskill{}, ErrUserSkillNotFound
	}

	return userSkill, nil
}
