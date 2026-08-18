package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/postgres/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository wraps sqlc queries for the auth module.
// It contains no business logic — only database access.
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

func (r *Repository) EmailExists(ctx context.Context, email string) (bool, error) {
	return r.queries.EmailExists(ctx, email)
}

func (r *Repository) UsernameExists(ctx context.Context, username string) (bool, error) {
	return r.queries.UsernameExists(ctx, username)
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (sqlc.User, error) {
	return r.queries.GetUserByEmail(ctx, email)
}

func (r *Repository) GetUserByID(ctx context.Context, id int64) (sqlc.User, error) {
	return r.queries.GetUserByID(ctx, id)
}

func (r *Repository) GetUserByPublicID(ctx context.Context, publicID uuid.UUID) (sqlc.User, error) {
	return r.queries.GetUserByPublicID(ctx, publicID)
}

func (r *Repository) GetIdentityByEmail(ctx context.Context, email string) (sqlc.UserIdentity, error) {
	return r.queries.GetIdentityByEmail(ctx, email)
}

// CreateUserWithLocalIdentity inserts a user and their local identity in one transaction.
// If either insert fails, both are rolled back.
func (r *Repository) CreateUserWithLocalIdentity(
	ctx context.Context,
	userParams sqlc.CreateUserParams,
	passwordHash string,
) (sqlc.User, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return sqlc.User{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	user, err := qtx.CreateUser(ctx, userParams)
	if err != nil {
		return sqlc.User{}, err
	}

	_, err = qtx.CreateIdentity(ctx, sqlc.CreateIdentityParams{
		UserID:       user.ID,
		Provider:     sqlc.ProviderNameLocal,
		PasswordHash: pgtype.Text{String: passwordHash, Valid: true},
	})
	if err != nil {
		return sqlc.User{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return sqlc.User{}, fmt.Errorf("commit transaction: %w", err)
	}

	return user, nil
}

// CreateUserWithLocalIdentityAndOTP inserts user, local identity, and verification OTP
// in one transaction so registration never leaves an account without a stored OTP.
func (r *Repository) CreateUserWithLocalIdentityAndOTP(
	ctx context.Context,
	userParams sqlc.CreateUserParams,
	passwordHash string,
	otpHash string,
	expiresAt time.Time,
) (sqlc.User, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return sqlc.User{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	user, err := qtx.CreateUser(ctx, userParams)
	if err != nil {
		return sqlc.User{}, err
	}

	_, err = qtx.CreateIdentity(ctx, sqlc.CreateIdentityParams{
		UserID:       user.ID,
		Provider:     sqlc.ProviderNameLocal,
		PasswordHash: pgtype.Text{String: passwordHash, Valid: true},
	})
	if err != nil {
		return sqlc.User{}, err
	}

	_, err = qtx.CreateEmailVerificationOTP(ctx, sqlc.CreateEmailVerificationOTPParams{
		UserID:    user.ID,
		OtpHash:   otpHash,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return sqlc.User{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return sqlc.User{}, fmt.Errorf("commit transaction: %w", err)
	}

	return user, nil
}

func (r *Repository) CreateSession(
	ctx context.Context,
	params sqlc.CreateSessionParams,
) (sqlc.UserSession, error) {
	return r.queries.CreateSession(ctx, params)
}

func (r *Repository) GetSessionByRefreshHash(
	ctx context.Context,
	refreshTokenHash string,
) (sqlc.UserSession, error) {
	return r.queries.GetSessionByRefreshHash(ctx, refreshTokenHash)
}

func (r *Repository) RotateRefreshToken(
	ctx context.Context,
	params sqlc.RotateRefreshTokenParams,
) (sqlc.UserSession, error) {
	return r.queries.RotateRefreshToken(ctx, params)
}

func (r *Repository) DeleteSession(ctx context.Context, refreshTokenHash string) error {
	return r.queries.DeleteSession(ctx, refreshTokenHash)
}

func (r *Repository) DeleteAllSessions(ctx context.Context, userID int64) error {
	return r.queries.DeleteAllSessions(ctx, userID)
}

func (r *Repository) VerifyUserEmail(ctx context.Context, userID int64) (sqlc.User, error) {
	return r.queries.VerifyUserEmail(ctx, userID)
}

func (r *Repository) CreateEmailVerificationOTP(
	ctx context.Context,
	userID int64,
	otpHash string,
	expiresAt time.Time,
) (sqlc.EmailVerificationOtp, error) {
	return r.queries.CreateEmailVerificationOTP(ctx, sqlc.CreateEmailVerificationOTPParams{
		UserID:    userID,
		OtpHash:   otpHash,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
}

func (r *Repository) GetActiveEmailVerificationOTPByUserID(ctx context.Context, userID int64) (sqlc.EmailVerificationOtp, error) {
	return r.queries.GetActiveEmailVerificationOTPByUserID(ctx, userID)
}

func (r *Repository) ConsumeEmailVerificationOTP(ctx context.Context, id int64) (sqlc.EmailVerificationOtp, error) {
	return r.queries.ConsumeEmailVerificationOTP(ctx, id)
}

func (r *Repository) InvalidateActiveEmailVerificationOTPs(ctx context.Context, userID int64) error {
	return r.queries.InvalidateActiveEmailVerificationOTPs(ctx, userID)
}

func (r *Repository) GetIdentityByProviderUserID(
	ctx context.Context,
	provider sqlc.ProviderName,
	providerUserID string,
) (sqlc.UserIdentity, error) {
	return r.queries.GetIdentityByProviderUserID(ctx, sqlc.GetIdentityByProviderUserIDParams{
		Provider:       provider,
		ProviderUserID: pgtype.Text{String: providerUserID, Valid: true},
	})
}

// CreateUserWithGoogleIdentity creates a Google-only account in one transaction.
func (r *Repository) CreateUserWithGoogleIdentity(
	ctx context.Context,
	userParams sqlc.CreateUserParams,
	providerUserID string,
) (sqlc.User, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return sqlc.User{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	user, err := qtx.CreateUser(ctx, userParams)
	if err != nil {
		return sqlc.User{}, err
	}

	_, err = qtx.CreateIdentity(ctx, sqlc.CreateIdentityParams{
		UserID:         user.ID,
		Provider:       sqlc.ProviderNameGoogle,
		ProviderUserID: pgtype.Text{String: providerUserID, Valid: true},
	})
	if err != nil {
		return sqlc.User{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return sqlc.User{}, fmt.Errorf("commit transaction: %w", err)
	}

	return user, nil
}

func (r *Repository) CreateGoogleIdentity(ctx context.Context, userID int64, providerUserID string) (sqlc.UserIdentity, error) {
	return r.queries.CreateIdentity(ctx, sqlc.CreateIdentityParams{
		UserID:         userID,
		Provider:       sqlc.ProviderNameGoogle,
		ProviderUserID: pgtype.Text{String: providerUserID, Valid: true},
	})
}

// VerifyEmailWithOTP marks an OTP consumed and sets users.email_verified atomically.
func (r *Repository) VerifyEmailWithOTP(ctx context.Context, userID int64, otpID int64) (sqlc.User, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return sqlc.User{}, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	if _, err := qtx.ConsumeEmailVerificationOTP(ctx, otpID); err != nil {
		return sqlc.User{}, err
	}

	user, err := qtx.VerifyUserEmail(ctx, userID)
	if err != nil {
		return sqlc.User{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return sqlc.User{}, fmt.Errorf("commit transaction: %w", err)
	}

	return user, nil
}

// IssueEmailVerificationOTP invalidates prior codes and stores a new OTP hash.
func (r *Repository) IssueEmailVerificationOTP(
	ctx context.Context,
	userID int64,
	otpHash string,
	expiresAt time.Time,
) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)

	if err := qtx.InvalidateActiveEmailVerificationOTPs(ctx, userID); err != nil {
		return err
	}

	_, err = qtx.CreateEmailVerificationOTP(ctx, sqlc.CreateEmailVerificationOTPParams{
		UserID:    userID,
		OtpHash:   otpHash,
		ExpiresAt: pgtype.Timestamptz{Time: expiresAt, Valid: true},
	})
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *Repository) GetUserIsAdminByPublicID(ctx context.Context, publicID uuid.UUID) (bool, error) {
	return r.queries.GetUserIsAdminByPublicID(ctx, publicID)
}

func (r *Repository) SetUsersAdminByUsernames(ctx context.Context, usernames []string) error {
	return r.queries.SetUsersAdminByUsernames(ctx, usernames)
}
