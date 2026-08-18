package billing

import "errors"

var (
	ErrBillingNotConfigured = errors.New("payment billing is not configured")
	ErrAlreadyPremium       = errors.New("already on premium")
	ErrInvalidSignature     = errors.New("invalid payment signature")
	ErrPaymentNotFound      = errors.New("payment not found")
	ErrPaymentForbidden     = errors.New("payment does not belong to this user")
	ErrPaymentNotPayable    = errors.New("payment cannot be completed")
	ErrOrderCreateFailed    = errors.New("failed to create payment order")
)
