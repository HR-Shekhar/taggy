package user

import (
	"context"
	"errors"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/cloudinary"
	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/postgres/sqlc"
	apperrors "github.com/HR-Shekhar/taggy-backend/internal/shared/errors"
	"github.com/HR-Shekhar/taggy-backend/internal/shared/logging"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rs/zerolog"
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9._]{3,30}$`)

type Service struct {
	repo     *Repository
	uploader *cloudinary.Client
	log      zerolog.Logger
}

func NewService(repo *Repository, uploader *cloudinary.Client, log zerolog.Logger) *Service {
	return &Service{repo: repo, uploader: uploader, log: log}
}

// GetProfileByUsername returns a profile. Private fields are included only when isOwner is true.
func (s *Service) GetProfileByUsername(
	ctx context.Context,
	username string,
	viewerPublicID *uuid.UUID,
) (sqlc.User, bool, error) {
	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			s.log.Debug().Str("username", username).Msg("profile not found")
			return sqlc.User{}, false, apperrors.ErrNotFound
		}
		return sqlc.User{}, false, logging.Unexpected(s.log, err, "get profile lookup failed")
	}

	isOwner := viewerPublicID != nil && user.PublicID == *viewerPublicID

	s.log.Debug().
		Str("username", user.Username).
		Bool("is_owner", isOwner).
		Msg("profile fetched")

	return user, isOwner, nil
}

func (s *Service) UpdateProfileByUsername(
	ctx context.Context,
	username string,
	viewerPublicID uuid.UUID,
	input UpdateProfileInput,
) (sqlc.User, error) {
	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.User{}, logging.Reject(s.log, apperrors.ErrNotFound, "update profile rejected: not found")
		}
		return sqlc.User{}, logging.Unexpected(s.log, err, "update profile lookup failed")
	}

	if user.PublicID != viewerPublicID {
		return sqlc.User{}, logging.Reject(s.log, apperrors.ErrForbidden, "update profile rejected: forbidden")
	}

	if input.Username != nil && *input.Username != user.Username {
		user, err = s.changeUsername(ctx, user, *input.Username)
		if err != nil {
			return sqlc.User{}, err
		}
	}

	updateParams := sqlc.UpdateUserProfileParams{ID: user.ID}

	if input.Name != nil {
		updateParams.Name = *input.Name
	} else {
		updateParams.Name = user.Name
	}

	if input.Bio != nil {
		updateParams.Bio = pgtype.Text{String: *input.Bio, Valid: true}
	} else {
		updateParams.Bio = user.Bio
	}

	if input.ProfilePictureURL != nil {
		updateParams.ProfilePictureUrl = pgtype.Text{String: *input.ProfilePictureURL, Valid: true}
	} else {
		updateParams.ProfilePictureUrl = user.ProfilePictureUrl
	}

	user, err = s.repo.UpdateProfile(ctx, updateParams)
	if err != nil {
		if isUniqueViolation(err) {
			return sqlc.User{}, logging.Reject(s.log, apperrors.ErrConflict, "update profile rejected: conflict")
		}
		return sqlc.User{}, logging.Unexpected(s.log, err, "update profile failed")
	}

	s.log.Info().
		Str("user_id", user.PublicID.String()).
		Msg("profile updated")

	return user, nil
}

func (s *Service) changeUsername(ctx context.Context, user sqlc.User, newUsername string) (sqlc.User, error) {
	if !usernamePattern.MatchString(newUsername) {
		s.log.Warn().
			Str("user_id", user.PublicID.String()).
			Msg("username change rejected: invalid format")
		return sqlc.User{}, ErrInvalidUsername
	}

	taken, err := s.repo.UsernameExists(ctx, newUsername)
	if err != nil {
		return sqlc.User{}, logging.Unexpected(s.log, err, "username change exists check failed")
	}
	if taken {
		s.log.Warn().
			Str("user_id", user.PublicID.String()).
			Msg("username change rejected: already taken")
		return sqlc.User{}, ErrUsernameTaken
	}

	reserved, err := s.repo.UsernameHistoryExists(ctx, newUsername)
	if err != nil {
		return sqlc.User{}, logging.Unexpected(s.log, err, "username change history check failed")
	}
	if reserved {
		s.log.Warn().
			Str("user_id", user.PublicID.String()).
			Msg("username change rejected: reserved in history")
		return sqlc.User{}, ErrUsernameTaken
	}

	updated, err := s.repo.ChangeUsername(ctx, user.ID, user.Username, newUsername)
	if err != nil {
		if isUniqueViolation(err) {
			return sqlc.User{}, logging.Reject(s.log, ErrUsernameTaken, "username change rejected: conflict")
		}
		return sqlc.User{}, logging.Unexpected(s.log, err, "username change failed")
	}

	s.log.Info().
		Str("user_id", user.PublicID.String()).
		Str("old_username", user.Username).
		Str("new_username", newUsername).
		Msg("username changed")

	return updated, nil
}

const maxAvatarBytes = 500 * 1024

func (s *Service) UploadAvatar(
	ctx context.Context,
	username string,
	viewerPublicID uuid.UUID,
	file io.Reader,
	filename string,
	size int64,
	contentType string,
) (sqlc.User, error) {
	if s.uploader == nil || !s.uploader.Available() {
		return sqlc.User{}, logging.Reject(s.log, ErrAvatarStorageUnavailable, "avatar upload rejected: cloudinary not configured")
	}
	if size <= 0 || size > maxAvatarBytes || !isAllowedAvatar(filename, contentType) {
		return sqlc.User{}, logging.Reject(s.log, ErrInvalidAvatar, "avatar upload rejected: invalid file")
	}

	user, err := s.repo.GetByUsername(ctx, username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sqlc.User{}, logging.Reject(s.log, apperrors.ErrNotFound, "avatar upload rejected: not found")
		}
		return sqlc.User{}, logging.Unexpected(s.log, err, "avatar upload lookup failed")
	}
	if user.PublicID != viewerPublicID {
		return sqlc.User{}, logging.Reject(s.log, apperrors.ErrForbidden, "avatar upload rejected: forbidden")
	}

	url, err := s.uploader.UploadAvatar(ctx, user.PublicID.String(), file, filename)
	if err != nil {
		if errors.Is(err, cloudinary.ErrNotConfigured) {
			return sqlc.User{}, logging.Reject(s.log, ErrAvatarStorageUnavailable, "avatar upload rejected: not configured")
		}
		return sqlc.User{}, logging.Reject(s.log, ErrAvatarUploadFailed, "avatar upload failed")
	}

	updated, err := s.repo.UpdateProfile(ctx, sqlc.UpdateUserProfileParams{
		ID:                user.ID,
		Name:              user.Name,
		Bio:               user.Bio,
		ProfilePictureUrl: pgtype.Text{String: url, Valid: true},
	})
	if err != nil {
		return sqlc.User{}, logging.Unexpected(s.log, err, "avatar upload: save url failed")
	}

	s.log.Info().
		Str("user_id", user.PublicID.String()).
		Msg("profile photo updated")

	return updated, nil
}

func isAllowedAvatar(filename, contentType string) bool {
	ct := strings.ToLower(strings.TrimSpace(contentType))
	if strings.Contains(ct, ";") {
		ct = strings.TrimSpace(strings.Split(ct, ";")[0])
	}
	switch ct {
	case "image/jpeg", "image/jpg", "image/png", "image/webp", "image/gif":
		return true
	}
	switch strings.ToLower(filepath.Ext(filename)) {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif":
		return true
	}
	return false
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func toProfileResponse(user sqlc.User, isOwner bool) profileResponse {
	resp := profileResponse{
		Username: user.Username,
		Name:     user.Name,
	}
	if user.Bio.Valid {
		resp.Bio = &user.Bio.String
	}
	if user.ProfilePictureUrl.Valid && strings.TrimSpace(user.ProfilePictureUrl.String) != "" {
		url := strings.TrimSpace(user.ProfilePictureUrl.String)
		resp.ProfilePictureURL = &url
	}
	if isOwner {
		publicID := user.PublicID.String()
		email := user.Email
		emailVerified := user.EmailVerified
		subscription := string(user.Subscription)
		isAdmin := user.IsAdmin()
		resp.PublicID = &publicID
		resp.Email = &email
		resp.EmailVerified = &emailVerified
		resp.Subscription = &subscription
		resp.IsAdmin = &isAdmin
	}
	return resp
}
