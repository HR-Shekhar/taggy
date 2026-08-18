package errors

import "errors"

var (
	ErrInternal = errors.New("internal error")

	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")

	ErrNotFound = errors.New("resource not found")

	ErrConflict = errors.New("resource conflict")

	ErrBadRequest = errors.New("bad request")

	ErrTooManyRequests = errors.New("too many requests")

	ErrServiceUnavailable = errors.New("service unavailable")
)
