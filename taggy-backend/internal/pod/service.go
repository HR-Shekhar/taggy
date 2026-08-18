package pod

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"strconv"
	"strings"

	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/postgres/sqlc"
	apperrors "github.com/HR-Shekhar/taggy-backend/internal/shared/errors"
	"github.com/HR-Shekhar/taggy-backend/internal/shared/logging"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
)

type Notifier interface {
	NotifyPodJoinRequest(ctx context.Context, ownerUserID int64, podID int64, podSlug, requesterUsername string)
	NotifyPodJoinAccepted(ctx context.Context, memberUserID int64, podID int64, podSlug string)
	NotifyPodJoinRejected(ctx context.Context, memberUserID int64, podID int64, podSlug string)
	NotifyPodMemberRemoved(ctx context.Context, memberUserID int64, podID int64, podSlug string)
}

type Service struct {
	repo     *Repository
	notifier Notifier
	log      zerolog.Logger
}

func NewService(repo *Repository, notifier Notifier, log zerolog.Logger) *Service {
	return &Service{repo: repo, notifier: notifier, log: log}
}

func (s *Service) GetUserPublicIDByUsername(ctx context.Context, username string) (uuid.UUID, error) {
	id, err := s.repo.GetUserPublicIDByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.UUID{}, err
		}
		return uuid.UUID{}, logging.Unexpected(s.log, err, "get user public id by username failed")
	}
	return id, nil
}

func (s *Service) CreatePod(
	ctx context.Context,
	userPublicID uuid.UUID,
	skillSlug string,
	input CreatePodInput,
) (sqlc.GetPodBySlugRow, error) {
	name := strings.TrimSpace(input.Name)
	if len(name) < 3 {
		return sqlc.GetPodBySlugRow{}, logging.Reject(s.log, ErrInvalidPodName, "create pod rejected: invalid name")
	}

	slug := normalizePodSlug(input.Slug)
	if err := validatePodSlug(slug); err != nil {
		return sqlc.GetPodBySlugRow{}, logging.Reject(s.log, err, "create pod rejected: invalid slug")
	}

	user, err := s.repo.GetUserByPublicID(ctx, userPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.GetPodBySlugRow{}, logging.Reject(s.log, apperrors.ErrNotFound, "create pod rejected: user not found")
		}
		return sqlc.GetPodBySlugRow{}, logging.Unexpected(s.log, err, "create pod: get user failed")
	}

	skill, err := s.repo.GetSkillBySlug(ctx, skillSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.GetPodBySlugRow{}, logging.Reject(s.log, apperrors.ErrNotFound, "create pod rejected: skill not found")
		}
		return sqlc.GetPodBySlugRow{}, logging.Unexpected(s.log, err, "create pod: get skill failed")
	}
	if !skill.IsActive {
		return sqlc.GetPodBySlugRow{}, logging.Reject(s.log, apperrors.ErrNotFound, "create pod rejected: skill inactive")
	}

	if _, err := s.repo.GetUserSkillByUserAndSkillSlug(ctx, user.ID, skill.Slug); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.GetPodBySlugRow{}, logging.Reject(s.log, ErrNotEnrolledInSkill, "create pod rejected: not enrolled in skill")
		}
		return sqlc.GetPodBySlugRow{}, logging.Unexpected(s.log, err, "create pod: check enrollment failed")
	}

	if _, err := s.repo.GetAcceptedPodMembershipByUserID(ctx, user.ID); err == nil {
		return sqlc.GetPodBySlugRow{}, logging.Reject(s.log, ErrAlreadyInActivePod, "create pod rejected: already in active pod")
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return sqlc.GetPodBySlugRow{}, logging.Unexpected(s.log, err, "create pod: check active membership failed")
	}

	slug, err = s.uniquePodSlug(ctx, slug)
	if err != nil {
		return sqlc.GetPodBySlugRow{}, err
	}

	desc := pgtype.Text{}
	if input.Description != nil && strings.TrimSpace(*input.Description) != "" {
		desc = pgtype.Text{String: strings.TrimSpace(*input.Description), Valid: true}
	}

	podRow, err := s.repo.CreatePodWithOwnerMembership(ctx, sqlc.CreatePodParams{
		Slug:        slug,
		Name:        name,
		Description: desc,
		OwnerID:     user.ID,
		SkillID:     skill.ID,
		MaxMembers:  7,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return sqlc.GetPodBySlugRow{}, logging.Reject(s.log, ErrPodSlugTaken, "create pod rejected: slug taken")
		}
		return sqlc.GetPodBySlugRow{}, logging.Unexpected(s.log, err, "create pod failed")
	}

	detail, err := s.repo.GetPodBySlug(ctx, podRow.Slug)
	if err != nil {
		return sqlc.GetPodBySlugRow{}, logging.Unexpected(s.log, err, "create pod: reload failed")
	}

	s.log.Info().
		Str("pod_slug", detail.Slug).
		Str("owner", user.Username).
		Msg("pod created")

	return detail, nil
}

func (s *Service) ListPodsBySkill(ctx context.Context, skillSlug string) ([]sqlc.ListPodsBySkillSlugRow, error) {
	if _, err := s.repo.GetSkillBySlug(ctx, skillSlug); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, logging.Unexpected(s.log, err, "list pods by skill: get skill failed")
	}
	rows, err := s.repo.ListPodsBySkillSlug(ctx, skillSlug)
	if err != nil {
		return nil, logging.Unexpected(s.log, err, "list pods by skill failed")
	}
	return rows, nil
}

func (s *Service) GetPod(ctx context.Context, podSlug string) (
	sqlc.GetPodBySlugRow,
	[]sqlc.ListAcceptedMembersByPodIDRow,
	[]sqlc.ListPendingMembersByPodIDRow,
	int64,
	error,
) {
	podRow, err := s.repo.GetPodBySlug(ctx, podSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.GetPodBySlugRow{}, nil, nil, 0, ErrPodNotFound
		}
		return sqlc.GetPodBySlugRow{}, nil, nil, 0, logging.Unexpected(s.log, err, "get pod failed")
	}

	members, err := s.repo.ListAcceptedMembersByPodID(ctx, podRow.ID)
	if err != nil {
		return sqlc.GetPodBySlugRow{}, nil, nil, 0, logging.Unexpected(s.log, err, "get pod: list members failed")
	}

	pending, err := s.repo.ListPendingMembersByPodID(ctx, podRow.ID)
	if err != nil {
		return sqlc.GetPodBySlugRow{}, nil, nil, 0, logging.Unexpected(s.log, err, "get pod: list pending failed")
	}

	count, err := s.repo.CountAcceptedPodMembers(ctx, podRow.ID)
	if err != nil {
		return sqlc.GetPodBySlugRow{}, nil, nil, 0, logging.Unexpected(s.log, err, "get pod: count members failed")
	}

	return podRow, members, pending, count, nil
}

func (s *Service) ListMyPods(ctx context.Context, userPublicID uuid.UUID) ([]sqlc.ListPodMembershipsByUserIDRow, error) {
	user, err := s.repo.GetUserByPublicID(ctx, userPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, logging.Unexpected(s.log, err, "list my pods: get user failed")
	}
	rows, err := s.repo.ListPodMembershipsByUserID(ctx, user.ID)
	if err != nil {
		return nil, logging.Unexpected(s.log, err, "list my pods failed")
	}
	return rows, nil
}

func (s *Service) RequestJoin(ctx context.Context, userPublicID uuid.UUID, podSlug string) (sqlc.PodMembership, error) {
	user, err := s.repo.GetUserByPublicID(ctx, userPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.PodMembership{}, logging.Reject(s.log, apperrors.ErrNotFound, "request join rejected: user not found")
		}
		return sqlc.PodMembership{}, logging.Unexpected(s.log, err, "request join: get user failed")
	}

	podRow, err := s.repo.GetPodBySlug(ctx, podSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.PodMembership{}, logging.Reject(s.log, ErrPodNotFound, "request join rejected: pod not found")
		}
		return sqlc.PodMembership{}, logging.Unexpected(s.log, err, "request join: get pod failed")
	}

	skill, err := s.repo.GetSkillBySlug(ctx, podRow.SkillSlug)
	if err != nil {
		return sqlc.PodMembership{}, logging.Unexpected(s.log, err, "request join: get skill failed")
	}
	if _, err := s.repo.GetUserSkillByUserAndSkillSlug(ctx, user.ID, skill.Slug); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.PodMembership{}, logging.Reject(s.log, ErrNotEnrolledInSkill, "request join rejected: not enrolled in skill")
		}
		return sqlc.PodMembership{}, logging.Unexpected(s.log, err, "request join: check enrollment failed")
	}

	if _, err := s.repo.GetAcceptedPodMembershipByUserID(ctx, user.ID); err == nil {
		return sqlc.PodMembership{}, logging.Reject(s.log, ErrAlreadyInActivePod, "request join rejected: already in active pod")
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return sqlc.PodMembership{}, logging.Unexpected(s.log, err, "request join: check active membership failed")
	}

	existing, err := s.repo.GetPodMembershipByPodAndUser(ctx, podRow.ID, user.ID)
	if err == nil {
		switch existing.Status {
		case sqlc.UserPodMembershipStatusACCEPTED:
			return sqlc.PodMembership{}, logging.Reject(s.log, ErrAlreadyMember, "request join rejected: already member")
		case sqlc.UserPodMembershipStatusPENDING:
			return sqlc.PodMembership{}, logging.Reject(s.log, ErrAlreadyPending, "request join rejected: already pending")
		case sqlc.UserPodMembershipStatusREJECTED, sqlc.UserPodMembershipStatusLEFT, sqlc.UserPodMembershipStatusREMOVED:
			updated, err := s.repo.UpdatePodMembershipStatus(ctx, existing.ID, sqlc.UserPodMembershipStatusPENDING, pgtype.Timestamptz{})
			if err != nil {
				return sqlc.PodMembership{}, logging.Unexpected(s.log, err, "request join: re-request failed")
			}
			s.notifyJoinRequest(ctx, podRow, user.Username)
			s.log.Info().Str("pod_slug", podSlug).Str("user", user.Username).Msg("pod join re-requested")
			return updated, nil
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return sqlc.PodMembership{}, logging.Unexpected(s.log, err, "request join: get membership failed")
	}

	membership, err := s.repo.CreatePodMembership(ctx, sqlc.CreatePodMembershipParams{
		PodID:    podRow.ID,
		UserID:   user.ID,
		Status:   sqlc.UserPodMembershipStatusPENDING,
		Role:     sqlc.PodMemberRoleMEMBER,
		JoinedAt: pgtype.Timestamptz{},
	})
	if err != nil {
		if isUniqueViolation(err) {
			return sqlc.PodMembership{}, logging.Reject(s.log, ErrAlreadyPending, "request join rejected: already pending")
		}
		return sqlc.PodMembership{}, logging.Unexpected(s.log, err, "request join: create membership failed")
	}

	s.notifyJoinRequest(ctx, podRow, user.Username)
	s.log.Info().Str("pod_slug", podSlug).Str("user", user.Username).Msg("pod join requested")
	return membership, nil
}

func (s *Service) AcceptMember(ctx context.Context, ownerPublicID uuid.UUID, podSlug, memberUsername string) (sqlc.PodMembership, error) {
	owner, membership, podRow, err := s.loadOwnerPendingMembership(ctx, ownerPublicID, podSlug, memberUsername)
	if err != nil {
		return sqlc.PodMembership{}, err
	}
	_ = owner

	updated, err := s.repo.AcceptMembershipAtomically(ctx, membership.ID, podRow.ID, membership.UserID, podRow.MaxMembers)
	if err != nil {
		if errors.Is(err, ErrPodFull) || errors.Is(err, ErrAlreadyInActivePod) {
			return sqlc.PodMembership{}, logging.Reject(s.log, err, "accept member rejected")
		}
		return sqlc.PodMembership{}, logging.Unexpected(s.log, err, "accept member failed")
	}

	if s.notifier != nil {
		s.notifier.NotifyPodJoinAccepted(ctx, membership.UserID, podRow.ID, podSlug)
	}

	s.log.Info().
		Str("pod_slug", podSlug).
		Str("member", memberUsername).
		Msg("pod member accepted")

	return updated, nil
}

func (s *Service) RejectMember(ctx context.Context, ownerPublicID uuid.UUID, podSlug, memberUsername string) (sqlc.PodMembership, error) {
	_, membership, podRow, err := s.loadOwnerPendingMembership(ctx, ownerPublicID, podSlug, memberUsername)
	if err != nil {
		return sqlc.PodMembership{}, err
	}

	updated, err := s.repo.UpdatePodMembershipStatus(ctx, membership.ID, sqlc.UserPodMembershipStatusREJECTED, pgtype.Timestamptz{})
	if err != nil {
		return sqlc.PodMembership{}, logging.Unexpected(s.log, err, "reject member failed")
	}

	if s.notifier != nil {
		s.notifier.NotifyPodJoinRejected(ctx, membership.UserID, podRow.ID, podSlug)
	}

	s.log.Info().Str("pod_slug", podSlug).Str("member", memberUsername).Msg("pod member rejected")
	return updated, nil
}

func (s *Service) SetMemberRole(
	ctx context.Context,
	ownerPublicID uuid.UUID,
	podSlug, memberUsername, roleRaw string,
) (sqlc.PodMembership, error) {
	role, err := parseAssignableRole(roleRaw)
	if err != nil {
		return sqlc.PodMembership{}, logging.Reject(s.log, err, "set member role rejected: invalid role")
	}

	owner, err := s.repo.GetUserByPublicID(ctx, ownerPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.PodMembership{}, logging.Reject(s.log, apperrors.ErrNotFound, "set member role rejected: user not found")
		}
		return sqlc.PodMembership{}, logging.Unexpected(s.log, err, "set member role: get owner failed")
	}

	podRow, err := s.repo.GetPodBySlug(ctx, podSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.PodMembership{}, logging.Reject(s.log, ErrPodNotFound, "set member role rejected: pod not found")
		}
		return sqlc.PodMembership{}, logging.Unexpected(s.log, err, "set member role: get pod failed")
	}
	if podRow.OwnerID != owner.ID {
		return sqlc.PodMembership{}, logging.Reject(s.log, ErrNotPodOwner, "set member role rejected: not owner")
	}

	member, err := s.repo.GetUserByUsername(ctx, memberUsername)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.PodMembership{}, logging.Reject(s.log, ErrMembershipNotFound, "set member role rejected: membership not found")
		}
		return sqlc.PodMembership{}, logging.Unexpected(s.log, err, "set member role: get member failed")
	}
	if member.ID == owner.ID {
		return sqlc.PodMembership{}, logging.Reject(s.log, ErrCannotChangeOwnRole, "set member role rejected: cannot change own role")
	}

	membership, err := s.repo.GetPodMembershipByPodAndUser(ctx, podRow.ID, member.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.PodMembership{}, logging.Reject(s.log, ErrMembershipNotFound, "set member role rejected: membership not found")
		}
		return sqlc.PodMembership{}, logging.Unexpected(s.log, err, "set member role: get membership failed")
	}
	if membership.Status != sqlc.UserPodMembershipStatusACCEPTED {
		return sqlc.PodMembership{}, logging.Reject(s.log, ErrNotAcceptedMember, "set member role rejected: not accepted member")
	}

	ownerMembership, err := s.repo.GetPodMembershipByPodAndUser(ctx, podRow.ID, owner.ID)
	if err != nil {
		return sqlc.PodMembership{}, logging.Unexpected(s.log, err, "set member role: get owner membership failed")
	}

	if role == sqlc.PodMemberRoleOWNER {
		if err := s.repo.TransferOwnershipAtomically(
			ctx,
			podRow.ID,
			ownerMembership.ID,
			membership.ID,
			member.ID,
		); err != nil {
			return sqlc.PodMembership{}, logging.Unexpected(s.log, err, "set member role: transfer ownership failed")
		}
		updated, err := s.repo.GetPodMembershipByPodAndUser(ctx, podRow.ID, member.ID)
		if err != nil {
			return sqlc.PodMembership{}, logging.Unexpected(s.log, err, "set member role: reload membership failed")
		}
		s.log.Info().
			Str("pod_slug", podSlug).
			Str("new_owner", memberUsername).
			Msg("pod ownership transferred")
		return updated, nil
	}

	updated, err := s.repo.UpdatePodMembershipRole(ctx, membership.ID, role)
	if err != nil {
		return sqlc.PodMembership{}, logging.Unexpected(s.log, err, "set member role failed")
	}

	s.log.Info().
		Str("pod_slug", podSlug).
		Str("member", memberUsername).
		Str("role", string(role)).
		Msg("pod member role updated")

	return updated, nil
}

func (s *Service) LeavePod(ctx context.Context, userPublicID uuid.UUID, podSlug string) error {
	user, err := s.repo.GetUserByPublicID(ctx, userPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return logging.Reject(s.log, apperrors.ErrNotFound, "leave pod rejected: user not found")
		}
		return logging.Unexpected(s.log, err, "leave pod: get user failed")
	}

	podRow, err := s.repo.GetPodBySlug(ctx, podSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return logging.Reject(s.log, ErrPodNotFound, "leave pod rejected: pod not found")
		}
		return logging.Unexpected(s.log, err, "leave pod: get pod failed")
	}

	membership, err := s.repo.GetPodMembershipByPodAndUser(ctx, podRow.ID, user.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return logging.Reject(s.log, ErrMembershipNotFound, "leave pod rejected: membership not found")
		}
		return logging.Unexpected(s.log, err, "leave pod: get membership failed")
	}

	if podRow.OwnerID == user.ID {
		return s.ownerLeave(ctx, podRow, user, membership)
	}

	if membership.Status != sqlc.UserPodMembershipStatusACCEPTED && membership.Status != sqlc.UserPodMembershipStatusPENDING {
		return logging.Reject(s.log, ErrNotAcceptedMember, "leave pod rejected: not accepted member")
	}

	_, err = s.repo.UpdatePodMembershipStatus(ctx, membership.ID, sqlc.UserPodMembershipStatusLEFT, pgtype.Timestamptz{})
	if err != nil {
		return logging.Unexpected(s.log, err, "leave pod failed")
	}

	s.log.Info().Str("pod_slug", podSlug).Str("user", user.Username).Msg("left pod")
	return nil
}

func (s *Service) ownerLeave(
	ctx context.Context,
	podRow sqlc.GetPodBySlugRow,
	owner sqlc.User,
	ownerMembership sqlc.PodMembership,
) error {
	if ownerMembership.Status != sqlc.UserPodMembershipStatusACCEPTED {
		return logging.Reject(s.log, ErrNotAcceptedMember, "owner leave rejected: not accepted member")
	}

	others, err := s.repo.ListAcceptedMembersExcludingUser(ctx, podRow.ID, owner.ID)
	if err != nil {
		return logging.Unexpected(s.log, err, "owner leave: list members failed")
	}

	if len(others) == 0 {
		if err := s.repo.OwnerLeaveAtomically(ctx, podRow.ID, owner.ID, ownerMembership.ID, 0, 0, true); err != nil {
			return logging.Unexpected(s.log, err, "owner leave: delete empty pod failed")
		}
		s.log.Info().Str("pod_slug", podRow.Slug).Str("user", owner.Username).Msg("owner left empty pod; pod deleted")
		return nil
	}

	admins, err := s.repo.ListAcceptedAdminsByPodID(ctx, podRow.ID)
	if err != nil {
		return logging.Unexpected(s.log, err, "owner leave: list admins failed")
	}

	var successorUserID, successorMembershipID int64
	if len(admins) > 0 {
		pick, err := pickRandomIndex(len(admins))
		if err != nil {
			return logging.Unexpected(s.log, err, "owner leave: pick admin successor failed")
		}
		successorUserID = admins[pick].UserID
		successorMembershipID = admins[pick].ID
	} else {
		pick, err := pickRandomIndex(len(others))
		if err != nil {
			return logging.Unexpected(s.log, err, "owner leave: pick member successor failed")
		}
		successorUserID = others[pick].UserID
		successorMembershipID = others[pick].ID
	}

	if err := s.repo.OwnerLeaveAtomically(
		ctx,
		podRow.ID,
		owner.ID,
		ownerMembership.ID,
		successorUserID,
		successorMembershipID,
		false,
	); err != nil {
		return logging.Unexpected(s.log, err, "owner leave: transfer ownership failed")
	}

	s.log.Info().
		Str("pod_slug", podRow.Slug).
		Str("user", owner.Username).
		Int64("new_owner_user_id", successorUserID).
		Msg("owner left after transferring ownership")
	return nil
}

func (s *Service) DeletePod(ctx context.Context, ownerPublicID uuid.UUID, podSlug string) error {
	owner, err := s.repo.GetUserByPublicID(ctx, ownerPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return logging.Reject(s.log, apperrors.ErrNotFound, "delete pod rejected: user not found")
		}
		return logging.Unexpected(s.log, err, "delete pod: get user failed")
	}

	podRow, err := s.repo.GetPodBySlug(ctx, podSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return logging.Reject(s.log, ErrPodNotFound, "delete pod rejected: pod not found")
		}
		return logging.Unexpected(s.log, err, "delete pod: get pod failed")
	}
	if podRow.OwnerID != owner.ID {
		return logging.Reject(s.log, ErrNotPodOwner, "delete pod rejected: not owner")
	}

	if err := s.repo.DeleteEmptyPod(ctx, podRow.ID, owner.ID); err != nil {
		if errors.Is(err, ErrPodNotEmpty) {
			return logging.Reject(s.log, err, "delete pod rejected: not empty")
		}
		return logging.Unexpected(s.log, err, "delete pod failed")
	}

	s.log.Info().Str("pod_slug", podSlug).Str("owner", owner.Username).Msg("empty pod deleted")
	return nil
}

func (s *Service) RemoveMember(ctx context.Context, ownerPublicID uuid.UUID, podSlug, memberUsername string) error {
	owner, err := s.repo.GetUserByPublicID(ctx, ownerPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return logging.Reject(s.log, apperrors.ErrNotFound, "remove member rejected: user not found")
		}
		return logging.Unexpected(s.log, err, "remove member: get owner failed")
	}

	podRow, err := s.repo.GetPodBySlug(ctx, podSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return logging.Reject(s.log, ErrPodNotFound, "remove member rejected: pod not found")
		}
		return logging.Unexpected(s.log, err, "remove member: get pod failed")
	}
	if podRow.OwnerID != owner.ID {
		return logging.Reject(s.log, ErrNotPodOwner, "remove member rejected: not owner")
	}

	member, err := s.repo.GetUserByUsername(ctx, memberUsername)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return logging.Reject(s.log, ErrMembershipNotFound, "remove member rejected: membership not found")
		}
		return logging.Unexpected(s.log, err, "remove member: get member failed")
	}
	if member.ID == owner.ID {
		return logging.Reject(s.log, ErrCannotRemoveOwner, "remove member rejected: cannot remove owner")
	}

	membership, err := s.repo.GetPodMembershipByPodAndUser(ctx, podRow.ID, member.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return logging.Reject(s.log, ErrMembershipNotFound, "remove member rejected: membership not found")
		}
		return logging.Unexpected(s.log, err, "remove member: get membership failed")
	}
	if membership.Status != sqlc.UserPodMembershipStatusACCEPTED {
		return logging.Reject(s.log, ErrNotAcceptedMember, "remove member rejected: not accepted member")
	}

	_, err = s.repo.UpdatePodMembershipStatus(ctx, membership.ID, sqlc.UserPodMembershipStatusREMOVED, pgtype.Timestamptz{})
	if err != nil {
		return logging.Unexpected(s.log, err, "remove member failed")
	}

	if s.notifier != nil {
		s.notifier.NotifyPodMemberRemoved(ctx, member.ID, podRow.ID, podSlug)
	}

	s.log.Info().Str("pod_slug", podSlug).Str("member", memberUsername).Msg("pod member removed")
	return nil
}

func (s *Service) loadOwnerPendingMembership(
	ctx context.Context,
	ownerPublicID uuid.UUID,
	podSlug, memberUsername string,
) (sqlc.User, sqlc.PodMembership, sqlc.GetPodBySlugRow, error) {
	owner, err := s.repo.GetUserByPublicID(ctx, ownerPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.User{}, sqlc.PodMembership{}, sqlc.GetPodBySlugRow{}, logging.Reject(s.log, apperrors.ErrNotFound, "owner action rejected: user not found")
		}
		return sqlc.User{}, sqlc.PodMembership{}, sqlc.GetPodBySlugRow{}, logging.Unexpected(s.log, err, "owner action: get owner failed")
	}

	podRow, err := s.repo.GetPodBySlug(ctx, podSlug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.User{}, sqlc.PodMembership{}, sqlc.GetPodBySlugRow{}, logging.Reject(s.log, ErrPodNotFound, "owner action rejected: pod not found")
		}
		return sqlc.User{}, sqlc.PodMembership{}, sqlc.GetPodBySlugRow{}, logging.Unexpected(s.log, err, "owner action: get pod failed")
	}
	if podRow.OwnerID != owner.ID {
		return sqlc.User{}, sqlc.PodMembership{}, sqlc.GetPodBySlugRow{}, logging.Reject(s.log, ErrNotPodOwner, "owner action rejected: not owner")
	}

	member, err := s.repo.GetUserByUsername(ctx, memberUsername)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.User{}, sqlc.PodMembership{}, sqlc.GetPodBySlugRow{}, logging.Reject(s.log, ErrMembershipNotFound, "owner action rejected: membership not found")
		}
		return sqlc.User{}, sqlc.PodMembership{}, sqlc.GetPodBySlugRow{}, logging.Unexpected(s.log, err, "owner action: get member failed")
	}

	membership, err := s.repo.GetPodMembershipByPodAndUser(ctx, podRow.ID, member.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.User{}, sqlc.PodMembership{}, sqlc.GetPodBySlugRow{}, logging.Reject(s.log, ErrMembershipNotFound, "owner action rejected: membership not found")
		}
		return sqlc.User{}, sqlc.PodMembership{}, sqlc.GetPodBySlugRow{}, logging.Unexpected(s.log, err, "owner action: get membership failed")
	}
	if membership.Status != sqlc.UserPodMembershipStatusPENDING {
		return sqlc.User{}, sqlc.PodMembership{}, sqlc.GetPodBySlugRow{}, logging.Reject(s.log, ErrMembershipNotPending, "owner action rejected: membership not pending")
	}

	return owner, membership, podRow, nil
}

func parseAssignableRole(raw string) (sqlc.PodMemberRole, error) {
	switch strings.ToUpper(strings.TrimSpace(raw)) {
	case string(sqlc.PodMemberRoleOWNER):
		return sqlc.PodMemberRoleOWNER, nil
	case string(sqlc.PodMemberRoleADMIN):
		return sqlc.PodMemberRoleADMIN, nil
	case string(sqlc.PodMemberRoleMEMBER):
		return sqlc.PodMemberRoleMEMBER, nil
	default:
		return "", ErrInvalidMemberRole
	}
}

func pickRandomIndex(n int) (int, error) {
	if n <= 0 {
		return 0, errors.New("empty candidate list")
	}
	v, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		return 0, err
	}
	return int(v.Int64()), nil
}

func (s *Service) uniquePodSlug(ctx context.Context, base string) (string, error) {
	base = normalizePodSlug(base)
	if err := validatePodSlug(base); err != nil {
		return "", logging.Reject(s.log, err, "unique pod slug: invalid base")
	}
	candidate := base
	for i := 2; i < 1000; i++ {
		exists, err := s.repo.PodSlugExists(ctx, candidate)
		if err != nil {
			return "", logging.Unexpected(s.log, err, "unique pod slug: exists check failed")
		}
		if !exists {
			return candidate, nil
		}
		suffix := "-" + strconv.Itoa(i)
		trimmed := base
		if len(trimmed)+len(suffix) > 60 {
			trimmed = strings.Trim(trimmed[:60-len(suffix)], "-")
			if len(trimmed) < 3 {
				trimmed = "pod"
			}
		}
		candidate = trimmed + suffix
	}
	return "", logging.Reject(s.log, ErrPodSlugTaken, "unique pod slug: exhausted")
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func (s *Service) notifyJoinRequest(ctx context.Context, podRow sqlc.GetPodBySlugRow, requesterUsername string) {
	if s.notifier == nil {
		return
	}
	s.notifier.NotifyPodJoinRequest(ctx, podRow.OwnerID, podRow.ID, podRow.Slug, requesterUsername)
}
