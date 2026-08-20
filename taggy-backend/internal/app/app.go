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
	"github.com/HR-Shekhar/taggy-backend/internal/aigen"
	"github.com/HR-Shekhar/taggy-backend/internal/audio"
	"github.com/HR-Shekhar/taggy-backend/internal/auth"
	"github.com/HR-Shekhar/taggy-backend/internal/billing"
	"github.com/HR-Shekhar/taggy-backend/internal/community"
	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/cloudinary"
	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/email"
	packageHttp "github.com/HR-Shekhar/taggy-backend/internal/infrastructure/http"
	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/livekit"
	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/openrouter"
	"github.com/HR-Shekhar/taggy-backend/internal/infrastructure/postgres"
	"github.com/HR-Shekhar/taggy-backend/internal/notification"
	"github.com/HR-Shekhar/taggy-backend/internal/pod"
	"github.com/HR-Shekhar/taggy-backend/internal/progress"
	"github.com/HR-Shekhar/taggy-backend/internal/quiz"
	"github.com/HR-Shekhar/taggy-backend/internal/report"
	"github.com/HR-Shekhar/taggy-backend/internal/roadmap"
	"github.com/HR-Shekhar/taggy-backend/internal/roadmaprequest"
	"github.com/HR-Shekhar/taggy-backend/internal/search"
	"github.com/HR-Shekhar/taggy-backend/internal/security/jwt"
	"github.com/HR-Shekhar/taggy-backend/internal/security/otp"
	"github.com/HR-Shekhar/taggy-backend/internal/security/password"
	"github.com/HR-Shekhar/taggy-backend/internal/security/token"
	"github.com/HR-Shekhar/taggy-backend/internal/shared/config"
	"github.com/HR-Shekhar/taggy-backend/internal/shared/logger"
	"github.com/HR-Shekhar/taggy-backend/internal/shared/validator"
	"github.com/HR-Shekhar/taggy-backend/internal/skill"
	"github.com/HR-Shekhar/taggy-backend/internal/skillrequest"
	"github.com/HR-Shekhar/taggy-backend/internal/user"

	jwtv5 "github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"
)

type App struct {
	Config       *config.Config
	Logger       zerolog.Logger
	DB           *pgxpool.Pool
	Echo         *echo.Echo
	audioService *audio.Service
	aiPool       *aigen.Pool
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
	packageHttp.RegisterMiddleware(e, log, cfg)

	authRepo := auth.NewRepository(db)
	jwtService := jwt.New(jwt.Config{
		SecretKey:     []byte(cfg.JWT.Secret),
		Issuer:        cfg.JWT.Issuer,
		TTL:           cfg.JWT.AccessTokenTTL,
		SigningMethod: jwtv5.SigningMethodHS256,
	})
	passwordService := password.New(password.Config{Cost: cfg.Auth.PasswordBcryptCost})
	tokenService := token.New()
	otpService := otp.New(tokenService.Hash)
	mailer, exposeDevOTP, err := email.NewSender(cfg.App.Environment, email.ResendConfig{
		APIKey: cfg.Email.ResendAPIKey,
		From:   cfg.Email.ResendFrom,
	}, log)
	if err != nil {
		return nil, err
	}
	googleOAuth := auth.NewGoogleOAuth(
		cfg.OAuth.GoogleClientID,
		cfg.OAuth.GoogleClientSecret,
		cfg.OAuth.GoogleRedirectURL,
		[]byte(cfg.JWT.Secret),
	)
	authService := auth.NewService(
		authRepo,
		passwordService,
		jwtService,
		tokenService,
		otpService,
		mailer,
		googleOAuth,
		log,
		cfg.Auth.EmailOTPTTL,
	)
	if err := authService.BootstrapAdmins(context.Background(), cfg.Admin.Usernames); err != nil {
		return nil, err
	}
	authHandler := auth.NewHandler(authService, log, exposeDevOTP, cfg.App.FrontendURL)

	userRepo := user.NewRepository(db)
	images := cloudinary.New(cfg.Cloudinary.URL, log)
	if !images.Available() {
		log.Warn().Msg("CLOUDINARY_URL not set; profile photo upload disabled")
	} else {
		log.Info().Msg("Cloudinary profile photos configured")
	}
	userService := user.NewService(userRepo, images, log)
	userHandler := user.NewHandler(userService, log)

	notificationRepo := notification.NewRepository(db)
	notificationService := notification.NewService(notificationRepo, log)
	notificationHandler := notification.NewHandler(notificationService, log)

	apiKey, model, baseURL, jsonMode := cfg.AI.Resolved()
	aiClient := openrouter.NewWithOptions(apiKey, model, baseURL, openrouter.Options{
		MaxTokens: 8192,
		JSONMode:  jsonMode,
		// Per-attempt HTTP timeout; pool job timeout (15m) bounds total retries.
		Timeout: 10 * time.Minute,
	}, log)
	if !aiClient.Available() {
		log.Warn().Msg("NVIDIA_API_KEY / AI_API_KEY not set; AI roadmap generation disabled")
	} else {
		log.Info().
			Str("model", model).
			Str("base_url", baseURL).
			Bool("json_mode", jsonMode).
			Msg("AI roadmap generator configured")
	}

	aiPool := aigen.NewPool(aigen.Config{
		Workers:    2,
		QueueSize:  64,
		JobTimeout: 15 * time.Minute,
	}, log)
	aiPool.Start()

	skillRequestRepo := skillrequest.NewRepository(db)
	skillRequestService := skillrequest.NewService(skillRequestRepo, aiClient, notificationService, aiPool, log)
	skillRequestHandler := skillrequest.NewHandler(skillRequestService, log)

	roadmapRequestRepo := roadmaprequest.NewRepository(db)
	roadmapRequestService := roadmaprequest.NewService(roadmapRequestRepo, aiClient, notificationService, aiPool, log)
	roadmapRequestHandler := roadmaprequest.NewHandler(roadmapRequestService, log)

	skillRepo := skill.NewRepository(db)
	skillService := skill.NewService(skillRepo, notificationService, log)
	skillHandler := skill.NewHandler(skillService, log)

	roadmapRepo := roadmap.NewRepository(db)
	roadmapService := roadmap.NewService(roadmapRepo, log)
	roadmapHandler := roadmap.NewHandler(roadmapService, log)

	progressRepo := progress.NewRepository(db)
	progressService := progress.NewService(progressRepo, log)
	progressHandler := progress.NewHandler(progressService, log)

	reportRepo := report.NewRepository(db)
	reportService := report.NewService(reportRepo, log)
	reportHandler := report.NewHandler(reportService, log)

	searchRepo := search.NewRepository(db)
	searchService := search.NewService(searchRepo, log)
	searchHandler := search.NewHandler(searchService, log)

	podRepo := pod.NewRepository(db)
	podService := pod.NewService(podRepo, notificationService, log)
	podHandler := pod.NewHandler(podService, log)

	communityHub := community.NewHub(log, []string{
		cfg.App.FrontendURL,
		"http://localhost:3000",
		"http://localhost:5173",
	})
	communityRepo := community.NewRepository(db)
	communityService := community.NewService(communityRepo, communityHub, log)
	communityHandler := community.NewHandler(communityService, communityHub, jwtService, log)

	livekitClient := livekit.NewTokenClient(livekit.Config{
		URL:       cfg.LiveKit.URL,
		APIKey:    cfg.LiveKit.APIKey,
		APISecret: cfg.LiveKit.APISecret,
	})
	audioRepo := audio.NewRepository(db)
	audioService := audio.NewService(audioRepo, livekitClient, log)
	audioHandler := audio.NewHandler(audioService, log)

	billingRepo := billing.NewRepository(db)
	billingService := billing.NewService(billingRepo, billing.Config{
		KeyID:         cfg.Razorpay.KeyID,
		KeySecret:     cfg.Razorpay.KeySecret,
		WebhookSecret: cfg.Razorpay.WebhookSecret,
		AmountPaise:   cfg.Razorpay.AmountPaise,
		Currency:      cfg.Razorpay.Currency,
	}, log)
	billingHandler := billing.NewHandler(billingService, log)
	if !billingService.Available() {
		log.Warn().Msg("RAZORPAY_KEY_ID / RAZORPAY_KEY_SECRET not set; billing disabled")
	} else {
		log.Info().
			Int32("premium_amount_paise", cfg.Razorpay.AmountPaise).
			Str("currency", cfg.Razorpay.Currency).
			Msg("Razorpay billing configured")
	}

	quizRepo := quiz.NewRepository(db)
	quizService := quiz.NewService(quizRepo, aiClient, notificationService, aiPool, log)
	quizHandler := quiz.NewHandler(quizService, log)

	skillRequestService.RequeueGenerating(context.Background())
	roadmapRequestService.RequeueGenerating(context.Background())
	quizService.RequeueGenerating(context.Background())

	jwtMiddleware := auth.JWT(jwtService, log)
	optionalJWTMiddleware := auth.OptionalJWT(jwtService, log)
	adminMiddleware := auth.RequireAdmin(authService.GetUserIsAdmin, log)

	api.RegisterRoutes(e, api.Routes{
		Auth: api.AuthRoutes{
			Handler:       authHandler,
			JWTMiddleware: jwtMiddleware,
		},
		User: api.UserRoutes{
			Handler:               userHandler,
			JWTMiddleware:         jwtMiddleware,
			OptionalJWTMiddleware: optionalJWTMiddleware,
		},
		Skill: api.SkillRoutes{
			Handler:       skillHandler,
			JWTMiddleware: jwtMiddleware,
		},
		SkillRequest: api.SkillRequestRoutes{
			Handler:       skillRequestHandler,
			JWTMiddleware: jwtMiddleware,
		},
		Roadmap: api.RoadmapRoutes{
			Handler:       roadmapHandler,
			JWTMiddleware: jwtMiddleware,
		},
		RoadmapRequest: api.RoadmapRequestRoutes{
			Handler:       roadmapRequestHandler,
			JWTMiddleware: jwtMiddleware,
		},
		Progress: api.ProgressRoutes{
			Handler:       progressHandler,
			JWTMiddleware: jwtMiddleware,
		},
		Pod: api.PodRoutes{
			Handler:       podHandler,
			JWTMiddleware: jwtMiddleware,
		},
		Community: api.CommunityRoutes{
			Handler:       communityHandler,
			JWTMiddleware: jwtMiddleware,
			Hub:           communityHub,
		},
		Audio: api.AudioRoutes{
			Handler:       audioHandler,
			JWTMiddleware: jwtMiddleware,
		},
		Notification: api.NotificationRoutes{
			Handler:       notificationHandler,
			JWTMiddleware: jwtMiddleware,
		},
		Report: api.ReportRoutes{
			Handler:       reportHandler,
			JWTMiddleware: jwtMiddleware,
		},
		Search: api.SearchRoutes{
			Handler:       searchHandler,
			JWTMiddleware: jwtMiddleware,
		},
		Admin: api.AdminRoutes{
			JWTMiddleware:   jwtMiddleware,
			AdminMiddleware: adminMiddleware,
			AuthHandler:     authHandler,
			SkillRequest:    skillRequestHandler,
			RoadmapRequest:  roadmapRequestHandler,
		},
		Billing: api.BillingRoutes{
			Handler:       billingHandler,
			JWTMiddleware: jwtMiddleware,
		},
		Quiz: api.QuizRoutes{
			Handler:       quizHandler,
			JWTMiddleware: jwtMiddleware,
		},
	})

	return &App{
		Config:       cfg,
		Logger:       log,
		DB:           db,
		Echo:         e,
		audioService: audioService,
		aiPool:       aiPool,
	}, nil
}

func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)
	defer stop()

	go a.audioService.RunEmptyRoomSweeper(ctx)

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

	<-ctx.Done()

	a.Logger.Info().
		Msg("Shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := a.Shutdown(shutdownCtx); err != nil {
		a.Logger.Error().
			Err(err).
			Msg("Failed to shut down cleanly")

		return err
	}

	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	if a.aiPool != nil {
		a.aiPool.Shutdown(ctx)
	}

	if err := a.Echo.Shutdown(ctx); err != nil {
		return err
	}

	if a.audioService != nil {
		if err := a.audioService.EndAllActiveRooms(ctx); err != nil {
			a.Logger.Error().
				Err(err).
				Msg("Failed to end active audio rooms during shutdown")
			// Continue shutdown — don't block process exit on this.
		}
	}

	a.DB.Close()

	a.Logger.Info().
		Msg("Application shutdown complete")

	return nil
}
