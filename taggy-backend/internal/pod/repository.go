package pod

import (
	"context"
	"errors"
	"fmt"
	"time"

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

func (r *Repository) GetUserByUsername(ctx context.Context, username string) (sqlc.User, error) {
	return r.queries.GetUserByUsername(ctx, username)
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

func (r *Repository) GetUserSkillByUserAndSkillSlug(ctx context.Context, userID int64, skillSlug string) (sqlc.Userskill, error) {
	return r.queries.GetUserSkillByUserAndSkillSlug(ctx, sqlc.GetUserSkillByUserAndSkillSlugParams{
		UserID: userID,
		Slug:   skillSlug,
	})
}

func (r *Repository) CreatePod(ctx context.Context, arg sqlc.CreatePodParams) (sqlc.Pod, error) {
	return r.queries.CreatePod(ctx, arg)
}

func (r *Repository) GetPodBySlug(ctx context.Context, slug string) (sqlc.GetPodBySlugRow, error) {
	return r.queries.GetPodBySlug(ctx, slug)
}

func (r *Repository) PodSlugExists(ctx context.Context, slug string) (bool, error) {
	return r.queries.PodSlugExists(ctx, slug)
}

func (r *Repository) ListPodsBySkillSlug(ctx context.Context, skillSlug string) ([]sqlc.ListPodsBySkillSlugRow, error) {
	return r.queries.ListPodsBySkillSlug(ctx, skillSlug)
}

func (r *Repository) CountAcceptedPodMembers(ctx context.Context, podID int64) (int64, error) {
	return r.queries.CountAcceptedPodMembers(ctx, podID)
}

func (r *Repository) CreatePodMembership(ctx context.Context, arg sqlc.CreatePodMembershipParams) (sqlc.PodMembership, error) {
	return r.queries.CreatePodMembership(ctx, arg)
}

func (r *Repository) GetPodMembershipByPodAndUser(ctx context.Context, podID, userID int64) (sqlc.PodMembership, error) {
	return r.queries.GetPodMembershipByPodAndUser(ctx, sqlc.GetPodMembershipByPodAndUserParams{
		PodID:  podID,
		UserID: userID,
	})
}

func (r *Repository) GetAcceptedPodMembershipByUserID(ctx context.Context, userID int64) (sqlc.PodMembership, error) {
	return r.queries.GetAcceptedPodMembershipByUserID(ctx, userID)
}

func (r *Repository) UpdatePodMembershipStatus(
	ctx context.Context,
	id int64,
	status sqlc.UserPodMembershipStatus,
	joinedAt pgtype.Timestamptz,
) (sqlc.PodMembership, error) {
	return r.queries.UpdatePodMembershipStatus(ctx, sqlc.UpdatePodMembershipStatusParams{
		ID:       id,
		Status:   status,
		JoinedAt: joinedAt,
	})
}

func (r *Repository) UpdatePodMembershipRole(
	ctx context.Context,
	id int64,
	role sqlc.PodMemberRole,
) (sqlc.PodMembership, error) {
	return r.queries.UpdatePodMembershipRole(ctx, sqlc.UpdatePodMembershipRoleParams{
		ID:   id,
		Role: role,
	})
}

func (r *Repository) ListPodMembershipsByUserID(ctx context.Context, userID int64) ([]sqlc.ListPodMembershipsByUserIDRow, error) {
	return r.queries.ListPodMembershipsByUserID(ctx, userID)
}

func (r *Repository) ListAcceptedMembersByPodID(ctx context.Context, podID int64) ([]sqlc.ListAcceptedMembersByPodIDRow, error) {
	return r.queries.ListAcceptedMembersByPodID(ctx, podID)
}

func (r *Repository) ListPendingMembersByPodID(ctx context.Context, podID int64) ([]sqlc.ListPendingMembersByPodIDRow, error) {
	return r.queries.ListPendingMembersByPodID(ctx, podID)
}

func (r *Repository) ListAcceptedAdminsByPodID(ctx context.Context, podID int64) ([]sqlc.PodMembership, error) {
	return r.queries.ListAcceptedAdminsByPodID(ctx, podID)
}

func (r *Repository) ListAcceptedMembersExcludingUser(
	ctx context.Context,
	podID, userID int64,
) ([]sqlc.PodMembership, error) {
	return r.queries.ListAcceptedMembersExcludingUser(ctx, sqlc.ListAcceptedMembersExcludingUserParams{
		PodID:  podID,
		UserID: userID,
	})
}

// CreatePodWithOwnerMembership inserts pod + owner ACCEPTED/OWNER membership atomically.
func (r *Repository) CreatePodWithOwnerMembership(
	ctx context.Context,
	podParams sqlc.CreatePodParams,
) (sqlc.Pod, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return sqlc.Pod{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	podRow, err := qtx.CreatePod(ctx, podParams)
	if err != nil {
		return sqlc.Pod{}, err
	}

	now := pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
	_, err = qtx.CreatePodMembership(ctx, sqlc.CreatePodMembershipParams{
		PodID:    podRow.ID,
		UserID:   podParams.OwnerID,
		Status:   sqlc.UserPodMembershipStatusACCEPTED,
		Role:     sqlc.PodMemberRoleOWNER,
		JoinedAt: now,
	})
	if err != nil {
		return sqlc.Pod{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return sqlc.Pod{}, fmt.Errorf("commit transaction: %w", err)
	}

	return podRow, nil
}

// AcceptMembershipAtomically re-checks capacity and one-active-pod, then accepts as MEMBER.
func (r *Repository) AcceptMembershipAtomically(
	ctx context.Context,
	membershipID int64,
	podID int64,
	userID int64,
	maxMembers int32,
) (sqlc.PodMembership, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return sqlc.PodMembership{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	count, err := qtx.CountAcceptedPodMembers(ctx, podID)
	if err != nil {
		return sqlc.PodMembership{}, err
	}
	if count >= int64(maxMembers) {
		return sqlc.PodMembership{}, ErrPodFull
	}

	if _, err := qtx.GetAcceptedPodMembershipByUserID(ctx, userID); err == nil {
		return sqlc.PodMembership{}, ErrAlreadyInActivePod
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return sqlc.PodMembership{}, err
	}

	now := pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
	updated, err := qtx.UpdatePodMembershipStatusAndRole(ctx, sqlc.UpdatePodMembershipStatusAndRoleParams{
		ID:       membershipID,
		Status:   sqlc.UserPodMembershipStatusACCEPTED,
		Role:     sqlc.PodMemberRoleMEMBER,
		JoinedAt: now,
	})
	if err != nil {
		return sqlc.PodMembership{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return sqlc.PodMembership{}, fmt.Errorf("commit transaction: %w", err)
	}

	return updated, nil
}

// TransferOwnershipAtomically moves OWNER to newOwner and demotes previous owner to ADMIN.
func (r *Repository) TransferOwnershipAtomically(
	ctx context.Context,
	podID, previousOwnerMembershipID, newOwnerMembershipID, newOwnerUserID int64,
) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	if _, err := qtx.UpdatePodOwner(ctx, sqlc.UpdatePodOwnerParams{
		ID:      podID,
		OwnerID: newOwnerUserID,
	}); err != nil {
		return err
	}

	if _, err := qtx.UpdatePodMembershipRole(ctx, sqlc.UpdatePodMembershipRoleParams{
		ID:   newOwnerMembershipID,
		Role: sqlc.PodMemberRoleOWNER,
	}); err != nil {
		return err
	}

	if _, err := qtx.UpdatePodMembershipRole(ctx, sqlc.UpdatePodMembershipRoleParams{
		ID:   previousOwnerMembershipID,
		Role: sqlc.PodMemberRoleADMIN,
	}); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// OwnerLeaveAtomically transfers ownership to successor then marks owner LEFT,
// or deletes the pod when the owner is the sole accepted member.
func (r *Repository) OwnerLeaveAtomically(
	ctx context.Context,
	podID, ownerUserID, ownerMembershipID int64,
	successorUserID, successorMembershipID int64,
	deletePod bool,
) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	if deletePod {
		if err := qtx.DeletePod(ctx, podID); err != nil {
			return err
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit transaction: %w", err)
		}
		return nil
	}

	if _, err := qtx.UpdatePodOwner(ctx, sqlc.UpdatePodOwnerParams{
		ID:      podID,
		OwnerID: successorUserID,
	}); err != nil {
		return err
	}

	if _, err := qtx.UpdatePodMembershipRole(ctx, sqlc.UpdatePodMembershipRoleParams{
		ID:   successorMembershipID,
		Role: sqlc.PodMemberRoleOWNER,
	}); err != nil {
		return err
	}

	if _, err := qtx.UpdatePodMembershipStatusAndRole(ctx, sqlc.UpdatePodMembershipStatusAndRoleParams{
		ID:       ownerMembershipID,
		Status:   sqlc.UserPodMembershipStatusLEFT,
		Role:     sqlc.PodMemberRoleMEMBER,
		JoinedAt: pgtype.Timestamptz{},
	}); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// DeleteEmptyPod deletes the pod when the caller is owner and sole accepted member.
func (r *Repository) DeleteEmptyPod(ctx context.Context, podID, ownerUserID int64) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	count, err := qtx.CountAcceptedPodMembers(ctx, podID)
	if err != nil {
		return err
	}
	if count != 1 {
		return ErrPodNotEmpty
	}

	others, err := qtx.ListAcceptedMembersExcludingUser(ctx, sqlc.ListAcceptedMembersExcludingUserParams{
		PodID:  podID,
		UserID: ownerUserID,
	})
	if err != nil {
		return err
	}
	if len(others) > 0 {
		return ErrPodNotEmpty
	}

	if err := qtx.DeletePod(ctx, podID); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
