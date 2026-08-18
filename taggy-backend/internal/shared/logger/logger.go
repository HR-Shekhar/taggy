package logger

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

func New(environment, logLevel string) (zerolog.Logger, error) {
	level, err := zerolog.ParseLevel(strings.ToLower(logLevel))
	if err != nil {
		return zerolog.Logger{}, fmt.Errorf("invalid log level %q: %w", logLevel, err)
	}

	var logger zerolog.Logger

	if strings.EqualFold(environment, "development") {
		output := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}

		logger = zerolog.New(output)
	} else {
		logger = zerolog.New(os.Stdout)
	}

	logger = logger.
		Level(level).
		With().
		Timestamp().
		Logger()

	return logger, nil
}
