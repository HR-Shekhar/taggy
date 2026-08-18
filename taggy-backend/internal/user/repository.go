package user

import (
	"context"
	"fmt"

	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository wraps sqlc queries for the user profile module.
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

func (r *Repository) GetByPublicID(ctx context.Context, publicID uuid.UUID) (sqlc.User, error) {
	return r.queries.GetUserByPublicID(ctx, publicID)
}

func (r *Repository) GetByUsername(ctx context.Context, username string) (sqlc.User, error) {
	return r.queries.GetUserByUsername(ctx, username)
}

func (r *Repository) UpdateProfile(
	ctx context.Context,
	params sqlc.UpdateUserProfileParams,
) (sqlc.User, error) {
	return r.queries.UpdateUserProfile(ctx, params)
}

func (r *Repository) UsernameExists(ctx context.Context, username string) (bool, error) {
	return r.queries.UsernameExists(ctx, username)
}

func (r *Repository) UsernameHistoryExists(ctx context.Context, username string) (bool, error) {
	return r.queries.UsernameHistoryExists(ctx, username)
}

// ChangeUsername atomically archives the old username and sets the new one.
// Both steps must succeed or neither is persisted.
func (r *Repository) ChangeUsername(
	ctx context.Context,
	userID int64,
	oldUsername string,
	newUsername string,
) (sqlc.User, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return sqlc.User{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	if _, err := qtx.CreateUsernameHistory(ctx, sqlc.CreateUsernameHistoryParams{
		UserID:   userID,
		Username: oldUsername,
	}); err != nil {
		return sqlc.User{}, err
	}

	user, err := qtx.UpdateUsername(ctx, sqlc.UpdateUsernameParams{
		ID:       userID,
		Username: newUsername,
	})
	if err != nil {
		return sqlc.User{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return sqlc.User{}, fmt.Errorf("commit transaction: %w", err)
	}

	return user, nil
}
