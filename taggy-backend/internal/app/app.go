package app

import (
	"github.com/HR-Shekhar/taggy-backend/internal/api"
	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/http"
	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/postgres"
	"github.com/HR-Shekhar/taggy-backend/internal/shared/config"
	"github.com/HR-Shekhar/taggy-backend/internal/shared/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

type App struct {
	Config *config.Config
	Logger zerolog.Logger
	DB     *pgxpool.Pool
	Echo   *echo.Echo
}

func New() (*App, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return nil, err
	}

	logger, err := logger.New(
		cfg.App.Environment,
		cfg.App.LogLevel,
	)
	if err != nil {
		return nil, err
	}

	db, err := postgres.New(cfg.DB, logger)
	if err != nil {
		return nil, err
	}
	
	e := http.New()

	return &App{
		Config: cfg,
		Logger: logger,
		DB:     db,
		Echo:   e,
	}, nil
}

func (a *App) Run() error {
	defer a.DB.Close()
	api.RegisterRoutes(a.Echo)

	a.Logger.Info().
		Str("port", a.Config.App.Port).
		Msg("Starting HTTP server")

	return a.Echo.Start(":" + a.Config.App.Port)
}

// func (a *App) Shutdown(ctx context.Context) error {
// 	defer a.DB.Close()
// }