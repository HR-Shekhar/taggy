package postgres

import (
	"context"
	"fmt"

	"github.com/HR-Shekhar/taggy-backend/internal/shared/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
)

func New(
	cfg config.DatabaseConfig,
	log zerolog.Logger,
) (*pgxpool.Pool, error) {

	log.Info().
		Msg("Connecting to PostgreSQL")

	poolConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse postgres config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	log.Info().
		Msg("Connected to PostgreSQL")

	return pool, nil
}
