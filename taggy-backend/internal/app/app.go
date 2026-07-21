package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HR-Shekhar/taggy-backend/internal/api"
	packageHttp "github.com/HR-Shekhar/taggy-backend/internal/infrastructure/http"
	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/postgres"
	"github.com/HR-Shekhar/taggy-backend/internal/shared/config"
	"github.com/HR-Shekhar/taggy-backend/internal/shared/logger"
	"github.com/HR-Shekhar/taggy-backend/internal/shared/validator"

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

	log, err := logger.New(
		cfg.App.Environment,
		cfg.App.LogLevel,
	)
	if err != nil {
		return nil, err
	}

	db, err := postgres.New(cfg.DB, log)
	if err != nil {
		return nil, err
	}

	e := packageHttp.New()

	e.Validator = validator.New()
	e.HTTPErrorHandler = packageHttp.ErrorHandler(log)

	return &App{
		Config: cfg,
		Logger: log,
		DB:     db,
		Echo:   e,
	}, nil
}

func (a *App) Run() error {
	// Register all API routes before starting the server.
	api.RegisterRoutes(a.Echo)

	// Create a context that is cancelled when the OS sends SIGINT (Ctrl+C)
	// or SIGTERM (Docker/Kubernetes/systemd).
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop() // Release resources used by NotifyContext when Run exits.

	// Start the HTTP server in another goroutine because Start() blocks forever.
	go func() {
		a.Logger.Info().
			Str("port", a.Config.App.Port).
			Msg("Starting HTTP server")

		if err := a.Echo.Start(":" + a.Config.App.Port); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {

			a.Logger.Error().
				Err(err).
				Msg("HTTP server failed")
		}
	}()

	// Block here until ctx is cancelled (Ctrl+C or SIGTERM signal is received).
	<-ctx.Done()
	// After signal is recieved:-
	a.Logger.Info().
		Msg("Shutdown signal received")

	// Give existing requests up to 10 seconds to finish.
	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	// Gracefully stop the server and close other resources.
	if err := a.Shutdown(shutdownCtx); err != nil {
		a.Logger.Error().
			Err(err).
			Msg("Failed to shut down cleanly")

		return err
	}

	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	if err := a.Echo.Shutdown(ctx); err != nil {
		return err
	}

	a.DB.Close()

	a.Logger.Info().
		Msg("Application shutdown complete")

	return nil
}