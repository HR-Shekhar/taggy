package auth

import (
	"context"
	"errors"
	"net/netip"
	"strings"
	"time"

	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/email"
	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/postgres/sqlc"
	"github.com/HR-Shekhar/taggy-backend/internal/security/jwt"
	"github.com/HR-Shekhar/taggy-backend/internal/security/otp"
	"github.com/HR-Shekhar/taggy-backend/internal/security/password"
	"github.com/HR-Shekhar/taggy-backend/internal/security/token"
	apperrors "github.com/HR-Shekhar/taggy-backend/internal/shared/errors"
	"github.com/HR-Shekhar/taggy-backend/internal/shared/logging"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
)

const refreshTokenTTL = 7 * 24 * time.Hour

type Service struct {
	repo      *Repository
	passwords *password.Service
	jwt       *jwt.Service
	tokens    *token.Service
	otp       *otp.Service
	mailer    email.Sender
	google    *GoogleOAuth
	log       zerolog.Logger
	otpTTL    time.Duration
}

func NewService(
	repo *Repository,
	passwords *password.Service,
	jwt *jwt.Service,
	tokens *token.Service,
	otpService *otp.Service,
	mailer email.Sender,
	google *GoogleOAuth,
	log zerolog.Logger,
	otpTTL time.Duration,
) *Service {
	return &Service{
		repo:      repo,
		passwords: passwords,
		jwt:       jwt,
		tokens:    tokens,
		otp:       otpService,
		mailer:    mailer,
		google:    google,
		log:       log,
		otpTTL:    otpTTL,
	}
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (PendingSignup, string, error) {
	if err := s.repo.DeleteExpiredPendingRegistrations(ctx); err != nil {
		return PendingSignup{}, "", logging.Unexpected(s.log, err, "register expired pending cleanup failed")
	}

	emailExists, err := s.repo.EmailExists(ctx, input.Email)
	if err != nil {
		return PendingSignup{}, "", logging.Unexpected(s.log, err, "register email check failed")
	}
	if emailExists {
		s.log.Warn().Str("email", input.Email).Msg("register rejected: email in use")
		return PendingSignup{}, "", ErrEmailInUse
	}

	usernameExists, err := s.repo.UsernameExists(ctx, input.Username)
	if err != nil {
		return PendingSignup{}, "", logging.Unexpected(s.log, err, "register username check failed")
	}
	if usernameExists {
		s.log.Warn().Str("username", input.Username).Msg("register rejected: username in use")
		return PendingSignup{}, "", ErrUsernameInUse
	}

	taken, err := s.repo.GetActivePendingRegistrationByUsername(ctx, input.Username)
	if err == nil && !strings.EqualFold(taken.Email, input.Email) {
		s.log.Warn().Str("username", input.Username).Msg("register rejected: username pending")
		return PendingSignup{}, "", ErrUsernameInUse
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return PendingSignup{}, "", logging.Unexpected(s.log, err, "register pending username check failed")
	}

	passwordHash, err := s.passwords.Hash(input.Password)
	if err != nil {
		return PendingSignup{}, "", logging.Unexpected(s.log, err, "register password hash failed")
	}

	code, err := s.otp.Generate()
	if err != nil {
		return PendingSignup{}, "", logging.Unexpected(s.log, err, "register otp generate failed")
	}

	expiresAt := time.Now().Add(s.otpTTL)
	pending, err := s.repo.UpsertPendingRegistration(
		ctx,
		input.Email,
		input.Username,
		input.Name,
		passwordHash,
		code.Hash,
		expiresAt,
	)
	if err != nil {
		if isUniqueViolation(err) {
			return PendingSignup{}, "", logging.Reject(s.log, apperrors.ErrConflict, "register rejected: conflict")
		}
		return PendingSignup{}, "", logging.Unexpected(s.log, err, "register pending upsert failed")
	}

	s.log.Info().
		Str("email", pending.Email).
		Str("username", pending.Username).
		Msg("pending registration stored; waiting for otp")

	if err := s.mailer.SendVerificationOTP(ctx, pending.Email, code.PlainText); err != nil {
		s.log.Error().
			Str("email", pending.Email).
			Err(err).
			Msg("verification email failed after register; user can resend")
	}

	return PendingSignup{
		Email:    pending.Email,
		Username: pending.Username,
		Name:     pending.Name,
	}, code.PlainText, nil
}

func (s *Service) Login(ctx context.Context, input LoginInput) (TokenPair, error) {
	identity, err := s.repo.GetIdentityByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.log.Warn().Str("email", input.Email).Msg("login failed: identity not found")
			return TokenPair{}, ErrInvalidCredentials
		}
		return TokenPair{}, logging.Unexpected(s.log, err, "login identity lookup failed")
	}

	if !identity.PasswordHash.Valid {
		return TokenPair{}, logging.Reject(s.log, ErrInvalidCredentials, "login rejected: no password")
	}

	if err := s.passwords.Verify(input.Password, identity.PasswordHash.String); err != nil {
		if errors.Is(err, password.ErrInvalidPassword) {
			s.log.Warn().Str("email", input.Email).Msg("login failed: invalid password")
			return TokenPair{}, ErrInvalidCredentials
		}
		return TokenPair{}, logging.Unexpected(s.log, err, "login password verify failed")
	}

	user, err := s.repo.GetUserByID(ctx, identity.UserID)
	if err != nil {
		return TokenPair{}, logging.Unexpected(s.log, err, "login user lookup failed")
	}

	if !user.EmailVerified {
		s.log.Warn().Str("email", input.Email).Msg("login rejected: email not verified")
		return TokenPair{}, ErrEmailNotVerified
	}

	pair, err := s.createSession(ctx, user, input.UserAgent, input.IPAddress)
	if err != nil {
		return TokenPair{}, err
	}

	s.log.Info().
		Str("user_id", user.PublicID.String()).
		Msg("user logged in")

	return pair, nil
}

func (s *Service) Refresh(ctx context.Context, refreshTokenPlain string) (TokenPair, error) {
	hash := s.tokens.Hash(refreshTokenPlain)

	session, err := s.repo.GetSessionByRefreshHash(ctx, hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.log.Warn().Msg("refresh rejected: session not found")
			return TokenPair{}, ErrInvalidRefreshToken
		}
		return TokenPair{}, logging.Unexpected(s.log, err, "refresh session lookup failed")
	}

	if session.ExpiresAt.Valid && session.ExpiresAt.Time.Before(time.Now()) {
		s.log.Warn().
			Int64("session_id", session.ID).
			Msg("refresh rejected: session expired")
		return TokenPair{}, ErrSessionExpired
	}

	user, err := s.repo.GetUserByID(ctx, session.UserID)
	if err != nil {
		return TokenPair{}, logging.Unexpected(s.log, err, "refresh user lookup failed")
	}

	newRefresh, err := s.tokens.Generate()
	if err != nil {
		return TokenPair{}, logging.Unexpected(s.log, err, "refresh token generate failed")
	}

	accessToken, err := s.jwt.Generate(user.PublicID, session.PublicID)
	if err != nil {
		return TokenPair{}, logging.Unexpected(s.log, err, "refresh access token generate failed")
	}

	_, err = s.repo.RotateRefreshToken(ctx, sqlc.RotateRefreshTokenParams{
		ID:               session.ID,
		RefreshTokenHash: newRefresh.Hash,
		ExpiresAt:        pgtype.Timestamptz{Time: time.Now().Add(refreshTokenTTL), Valid: true},
	})
	if err != nil {
		return TokenPair{}, logging.Unexpected(s.log, err, "refresh rotate token failed")
	}

	s.log.Info().
		Str("user_id", user.PublicID.String()).
		Str("session_id", session.PublicID.String()).
		Msg("session refreshed")

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefresh.PlainText,
		Username:     user.Username,
		IsAdmin:      user.IsAdmin(),
		Subscription: string(user.Subscription),
	}, nil
}

func (s *Service) Logout(ctx context.Context, refreshTokenPlain string) error {
	hash := s.tokens.Hash(refreshTokenPlain)

	if err := s.repo.DeleteSession(ctx, hash); err != nil {
		return logging.Unexpected(s.log, err, "logout delete session failed")
	}

	s.log.Info().Msg("user logged out")

	return nil
}

func (s *Service) LogoutAll(ctx context.Context, userPublicID uuid.UUID) error {
	user, err := s.repo.GetUserByPublicID(ctx, userPublicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return logging.Reject(s.log, apperrors.ErrNotFound, "logout all rejected: user not found")
		}
		return logging.Unexpected(s.log, err, "logout all user lookup failed")
	}

	if err := s.repo.DeleteAllSessions(ctx, user.ID); err != nil {
		return logging.Unexpected(s.log, err, "logout all delete sessions failed")
	}

	s.log.Info().
		Str("user_id", user.PublicID.String()).
		Msg("user logged out from all devices")

	return nil
}

func (s *Service) createSession(
	ctx context.Context,
	user sqlc.User,
	userAgent *string,
	ipAddress *netip.Addr,
) (TokenPair, error) {
	newRefresh, err := s.tokens.Generate()
	if err != nil {
		return TokenPair{}, logging.Unexpected(s.log, err, "create session refresh token generate failed")
	}

	session, err := s.repo.CreateSession(ctx, sqlc.CreateSessionParams{
		UserID:           user.ID,
		RefreshTokenHash: newRefresh.Hash,
		UserAgent:        optionalText(userAgent),
		IpAddress:        ipAddress,
		ExpiresAt:        pgtype.Timestamptz{Time: time.Now().Add(refreshTokenTTL), Valid: true},
	})
	if err != nil {
		return TokenPair{}, logging.Unexpected(s.log, err, "create session failed")
	}

	accessToken, err := s.jwt.Generate(user.PublicID, session.PublicID)
	if err != nil {
		return TokenPair{}, logging.Unexpected(s.log, err, "create session access token generate failed")
	}

	return TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefresh.PlainText,
		Username:     user.Username,
		IsAdmin:      user.IsAdmin(),
		Subscription: string(user.Subscription),
	}, nil
}

func (s *Service) GetUserIsAdmin(ctx context.Context, publicID uuid.UUID) (bool, error) {
	return s.repo.GetUserIsAdminByPublicID(ctx, publicID)
}

func (s *Service) BootstrapAdmins(ctx context.Context, usernamesCSV string) error {
	parts := strings.Split(usernamesCSV, ",")
	usernames := make([]string, 0, len(parts))
	for _, p := range parts {
		u := strings.TrimSpace(p)
		if u != "" {
			usernames = append(usernames, u)
		}
	}
	if len(usernames) == 0 {
		return nil
	}
	if err := s.repo.SetUsersAdminByUsernames(ctx, usernames); err != nil {
		return logging.Unexpected(s.log, err, "bootstrap admins failed")
	}
	s.log.Info().Strs("usernames", usernames).Msg("admin usernames bootstrapped")
	return nil
}

func (s *Service) AdminMe(ctx context.Context, publicID uuid.UUID) (sqlc.User, error) {
	user, err := s.repo.GetUserByPublicID(ctx, publicID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.User{}, logging.Reject(s.log, apperrors.ErrNotFound, "admin me rejected: user not found")
		}
		return sqlc.User{}, logging.Unexpected(s.log, err, "admin me lookup failed")
	}
	if !user.IsAdmin() {
		return sqlc.User{}, logging.Reject(s.log, apperrors.ErrForbidden, "admin me rejected: not admin")
	}
	return user, nil
}

func optionalText(value *string) pgtype.Text {
	if value == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *value, Valid: true}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
