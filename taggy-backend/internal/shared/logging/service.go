package logging

import "github.com/rs/zerolog"

// Unexpected logs infra/unexpected errors at Error level and returns err unchanged.
// Use before returning database, network, or other unexpected failures from services.
func Unexpected(log zerolog.Logger, err error, msg string) error {
	if err == nil {
		return nil
	}
	log.Error().Err(err).Msg(msg)
	return err
}

// Reject logs expected domain/business failures at Warn level and returns err unchanged.
// Use for not-found, conflict, forbidden, validation-style service errors.
func Reject(log zerolog.Logger, err error, msg string) error {
	if err == nil {
		return nil
	}
	log.Warn().Err(err).Msg(msg)
	return err
}
