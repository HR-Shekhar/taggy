package notification

import "errors"

var (
	ErrNotificationNotFound = errors.New("notification not found")
	ErrAlreadyRead          = errors.New("notification already read")
)
