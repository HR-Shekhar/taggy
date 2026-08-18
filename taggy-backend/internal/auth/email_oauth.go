package auth

import (
	"context"
	"errors"
	"net/netip"
	"strings"
	"time"

	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/postgres/sqlc"
	apperrors "github.com/HR-Shekhar/taggy-backend/internal/shared/errors"
	"github.com/HR-Shekhar/taggy-backend/internal/shared/logging"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Service) VerifyEmail(ctx context.Context, email, otp string) (sqlc.User, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.User{}, logging.Reject(s.log, apperrors.ErrNotFound, "verify email rejected: user not found")
		}
		return sqlc.User{}, logging.Unexpected(s.log, err, "verify email user lookup failed")
	}

	if user.EmailVerified {
		return sqlc.User{}, logging.Reject(s.log, ErrEmailAlreadyVerified, "verify email rejected: already verified")
	}

	otpRow, err := s.repo.GetActiveEmailVerificationOTPByUserID(ctx, user.ID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.User{}, logging.Reject(s.log, ErrOTPExpired, "verify email rejected: otp expired")
		}
		return sqlc.User{}, logging.Unexpected(s.log, err, "verify email otp lookup failed")
	}

	if s.otp.Hash(otp) != otpRow.OtpHash {
		s.log.Warn().Str("email", email).Msg("email verification rejected: invalid otp")
		return sqlc.User{}, ErrInvalidOTP
	}

	verified, err := s.repo.VerifyEmailWithOTP(ctx, user.ID, otpRow.ID)
	if err != nil {
		return sqlc.User{}, logging.Unexpected(s.log, err, "verify email commit failed")
	}

	s.log.Info().
		Str("user_id", verified.PublicID.String()).
		Msg("email verified")

	return verified, nil
}

func (s *Service) ResendVerification(ctx context.Context, email string) (string, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", logging.Reject(s.log, apperrors.ErrNotFound, "resend verification rejected: user not found")
		}
		return "", logging.Unexpected(s.log, err, "resend verification user lookup failed")
	}

	if user.EmailVerified {
		return "", logging.Reject(s.log, ErrEmailAlreadyVerified, "resend verification rejected: already verified")
	}

	return s.issueEmailVerificationOTP(ctx, user)
}

func (s *Service) issueEmailVerificationOTP(ctx context.Context, user sqlc.User) (string, error) {
	code, err := s.otp.Generate()
	if err != nil {
		return "", logging.Unexpected(s.log, err, "issue verification otp generate failed")
	}

	expiresAt := time.Now().Add(s.otpTTL)
	if err := s.repo.IssueEmailVerificationOTP(ctx, user.ID, code.Hash, expiresAt); err != nil {
		return "", logging.Unexpected(s.log, err, "issue verification otp store failed")
	}

	if err := s.mailer.SendVerificationOTP(ctx, user.Email, code.PlainText); err != nil {
		s.log.Error().
			Str("user_id", user.PublicID.String()).
			Err(err).
			Msg("verification email failed")
		return "", logging.Reject(s.log, ErrEmailDeliveryFailed, "verification email delivery failed")
	}

	s.log.Info().
		Str("user_id", user.PublicID.String()).
		Msg("verification otp issued")

	return code.PlainText, nil
}

func (s *Service) GoogleAuthURL() (string, error) {
	if s.google == nil || !s.google.Enabled() {
		return "", logging.Reject(s.log, ErrOAuthNotConfigured, "google auth url rejected: oauth not configured")
	}
	url, err := s.google.AuthURL()
	if err != nil {
		return "", logging.Unexpected(s.log, err, "google auth url failed")
	}
	return url, nil
}

// CompleteGoogleOAuth finishes the Google redirect.
// Existing users get tokens immediately. New Google users get a pending
// registration token and no DB row until they submit username/details.
func (s *Service) CompleteGoogleOAuth(
	ctx context.Context,
	code string,
	state string,
	userAgent *string,
	ipAddress *netip.Addr,
) (GoogleOAuthResult, error) {
	if s.google == nil || !s.google.Enabled() {
		return GoogleOAuthResult{}, logging.Reject(s.log, ErrOAuthNotConfigured, "google oauth rejected: oauth not configured")
	}

	googleUser, err := s.google.ExchangeUser(ctx, code, state)
	if err != nil {
		if errors.Is(err, ErrInvalidOAuthState) {
			return GoogleOAuthResult{}, logging.Reject(s.log, err, "google oauth rejected: invalid state")
		}
		return GoogleOAuthResult{}, logging.Unexpected(s.log, err, "google oauth exchange failed")
	}

	if !googleUser.EmailVerified {
		return GoogleOAuthResult{}, logging.Reject(s.log, ErrOAuthAccountInvalid, "google oauth rejected: email not verified")
	}

	identity, err := s.repo.GetIdentityByProviderUserID(ctx, sqlc.ProviderNameGoogle, googleUser.ID)
	if err == nil {
		user, err := s.repo.GetUserByID(ctx, identity.UserID)
		if err != nil {
			return GoogleOAuthResult{}, logging.Unexpected(s.log, err, "google oauth user lookup failed")
		}
		pair, err := s.createSession(ctx, user, userAgent, ipAddress)
		if err != nil {
			return GoogleOAuthResult{}, err
		}
		s.log.Info().
			Str("user_id", user.PublicID.String()).
			Msg("user logged in via google")
		return GoogleOAuthResult{Tokens: &pair}, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return GoogleOAuthResult{}, logging.Unexpected(s.log, err, "google oauth identity lookup failed")
	}

	existing, err := s.repo.GetUserByEmail(ctx, googleUser.Email)
	if err == nil {
		if _, err := s.repo.CreateGoogleIdentity(ctx, existing.ID, googleUser.ID); err != nil {
			if isUniqueViolation(err) {
				user, err := s.repo.GetUserByID(ctx, existing.ID)
				if err != nil {
					return GoogleOAuthResult{}, logging.Unexpected(s.log, err, "google oauth user lookup failed")
				}
				pair, err := s.createSession(ctx, user, userAgent, ipAddress)
				if err != nil {
					return GoogleOAuthResult{}, err
				}
				s.log.Info().
					Str("user_id", user.PublicID.String()).
					Msg("user logged in via google")
				return GoogleOAuthResult{Tokens: &pair}, nil
			}
			return GoogleOAuthResult{}, logging.Unexpected(s.log, err, "google oauth link identity failed")
		}
		if !existing.EmailVerified {
			if _, err := s.repo.VerifyUserEmail(ctx, existing.ID); err != nil {
				return GoogleOAuthResult{}, logging.Unexpected(s.log, err, "google oauth verify email failed")
			}
			existing.EmailVerified = true
		}
		pair, err := s.createSession(ctx, existing, userAgent, ipAddress)
		if err != nil {
			return GoogleOAuthResult{}, err
		}
		s.log.Info().
			Str("user_id", existing.PublicID.String()).
			Msg("user logged in via google")
		return GoogleOAuthResult{Tokens: &pair}, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return GoogleOAuthResult{}, logging.Unexpected(s.log, err, "google oauth email lookup failed")
	}

	regToken, err := s.google.IssuePendingRegistration(googleUser)
	if err != nil {
		return GoogleOAuthResult{}, logging.Unexpected(s.log, err, "google oauth pending token issue failed")
	}

	s.log.Info().
		Str("email", googleUser.Email).
		Msg("google oauth pending profile completion")

	return GoogleOAuthResult{
		Pending: &PendingGoogleRegistration{
			RegistrationToken: regToken,
			Email:             googleUser.Email,
			Name:              strings.TrimSpace(googleUser.Name),
			Picture:           strings.TrimSpace(googleUser.Picture),
		},
	}, nil
}

// CompleteGoogleRegistration creates the user after they chose a username.
// If they never call this, the pending token expires and no account is created.
func (s *Service) CompleteGoogleRegistration(
	ctx context.Context,
	registrationToken string,
	username string,
	name string,
	userAgent *string,
	ipAddress *netip.Addr,
) (TokenPair, error) {
	if s.google == nil || !s.google.Enabled() {
		return TokenPair{}, logging.Reject(s.log, ErrOAuthNotConfigured, "google registration rejected: oauth not configured")
	}

	googleUser, err := s.google.ParsePendingRegistration(registrationToken)
	if err != nil {
		return TokenPair{}, logging.Reject(s.log, err, "google registration rejected: invalid token")
	}

	username = strings.TrimSpace(username)
	if username == "" {
		return TokenPair{}, logging.Reject(s.log, ErrOAuthUsernameRequired, "google registration rejected: username required")
	}
	if err := validateUsername(username); err != nil {
		return TokenPair{}, logging.Reject(s.log, err, "google registration rejected: invalid username")
	}

	exists, err := s.repo.UsernameExists(ctx, username)
	if err != nil {
		return TokenPair{}, logging.Unexpected(s.log, err, "google registration username check failed")
	}
	if exists {
		return TokenPair{}, logging.Reject(s.log, ErrUsernameInUse, "google registration rejected: username in use")
	}

	// Race-safe: Google identity may have been created by a parallel complete.
	if identity, err := s.repo.GetIdentityByProviderUserID(ctx, sqlc.ProviderNameGoogle, googleUser.ID); err == nil {
		user, err := s.repo.GetUserByID(ctx, identity.UserID)
		if err != nil {
			return TokenPair{}, logging.Unexpected(s.log, err, "google registration user lookup failed")
		}
		return s.createSession(ctx, user, userAgent, ipAddress)
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return TokenPair{}, logging.Unexpected(s.log, err, "google registration identity lookup failed")
	}

	if existing, err := s.repo.GetUserByEmail(ctx, googleUser.Email); err == nil {
		_ = existing
		return TokenPair{}, logging.Reject(s.log, ErrEmailInUse, "google registration rejected: email in use")
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return TokenPair{}, logging.Unexpected(s.log, err, "google registration email lookup failed")
	}

	displayName := strings.TrimSpace(name)
	if displayName == "" {
		displayName = strings.TrimSpace(googleUser.Name)
	}
	if displayName == "" {
		displayName = username
	}

	user, err := s.repo.CreateUserWithGoogleIdentity(ctx, sqlc.CreateUserParams{
		Email:             googleUser.Email,
		Username:          username,
		Name:              displayName,
		Subscription:      sqlc.SubscriptionTierFREE,
		EmailVerified:     true,
		ProfilePictureUrl: optionalPictureURL(googleUser.Picture),
		Bio:               pgtype.Text{},
	}, googleUser.ID)
	if err != nil {
		if isUniqueViolation(err) {
			return TokenPair{}, logging.Reject(s.log, ErrUsernameInUse, "google registration rejected: username conflict")
		}
		return TokenPair{}, logging.Unexpected(s.log, err, "google registration create user failed")
	}

	s.log.Info().
		Str("user_id", user.PublicID.String()).
		Str("username", user.Username).
		Msg("google account created")

	return s.createSession(ctx, user, userAgent, ipAddress)
}

func optionalPictureURL(url string) pgtype.Text {
	url = strings.TrimSpace(url)
	if url == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: url, Valid: true}
}
