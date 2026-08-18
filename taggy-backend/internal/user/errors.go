package user

import "errors"

var (
	ErrUsernameTaken            = errors.New("username is not available")
	ErrInvalidUsername          = errors.New("username format is invalid")
	ErrAvatarStorageUnavailable = errors.New("profile photo upload is unavailable")
	ErrInvalidAvatar            = errors.New("choose a JPG, PNG, WEBP, or GIF under 500 KB")
	ErrAvatarUploadFailed       = errors.New("failed to upload profile photo")
)
