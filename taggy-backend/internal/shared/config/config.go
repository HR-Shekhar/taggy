package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	App        AppConfig
	DB         DatabaseConfig `envPrefix:"DB_"`
	JWT        JWTConfig
	Auth       AuthConfig
	OAuth      OAuthConfig
	LiveKit    LiveKitConfig
	Email      EmailConfig
	AI         AIConfig
	Admin      AdminConfig
	Razorpay   RazorpayConfig
	Cloudinary CloudinaryConfig
}

// AIConfig is any OpenAI-compatible chat API (NVIDIA NIM preferred, OpenRouter fallback).
type AIConfig struct {
	// Preferred: NVIDIA NIM / generic
	APIKey  string `env:"AI_API_KEY"`
	NVIDIA  string `env:"NVIDIA_API_KEY"`
	Model   string `env:"AI_MODEL" envDefault:"z-ai/glm-5.2"`
	BaseURL string `env:"AI_BASE_URL" envDefault:"https://integrate.api.nvidia.com/v1"`

	// Legacy OpenRouter (still works if NVIDIA/AI key unset)
	OpenRouterAPIKey  string `env:"OPENROUTER_API_KEY"`
	OpenRouterModel   string `env:"OPENROUTER_MODEL"`
	OpenRouterBaseURL string `env:"OPENROUTER_BASE_URL"`
}

// Resolved returns apiKey, model, baseURL, useJSONMode for the active provider.
func (a AIConfig) Resolved() (apiKey, model, baseURL string, jsonMode bool) {
	apiKey = firstNonEmpty(a.APIKey, a.NVIDIA, a.OpenRouterAPIKey)
	usingNVIDIA := firstNonEmpty(a.APIKey, a.NVIDIA) != ""
	usingOpenRouterOnly := !usingNVIDIA && strings.TrimSpace(a.OpenRouterAPIKey) != ""

	if usingOpenRouterOnly {
		model = firstNonEmpty(a.OpenRouterModel, a.Model, "openai/gpt-4o-mini")
		baseURL = firstNonEmpty(a.OpenRouterBaseURL, "https://openrouter.ai/api/v1")
		jsonMode = true
		return apiKey, model, baseURL, jsonMode
	}

	model = firstNonEmpty(a.Model, "z-ai/glm-5.2")
	baseURL = firstNonEmpty(a.BaseURL, "https://integrate.api.nvidia.com/v1")
	// NIM GLM typically does not support response_format=json_object.
	return apiKey, model, baseURL, false
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

type AppConfig struct {
	Environment string `env:"APP_ENV,required,notEmpty"`
	Port        string `env:"PORT,required,notEmpty"`
	LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`
	FrontendURL string `env:"FRONTEND_URL" envDefault:"http://localhost:3000"`
}

type DatabaseConfig struct {
	Host     string `env:"HOST,required,notEmpty"`
	Port     string `env:"PORT,required,notEmpty"`
	User     string `env:"USER,required,notEmpty"`
	Password string `env:"PASSWORD,required,notEmpty"`
	Name     string `env:"NAME,required,notEmpty"`
	SSLMode  string `env:"SSLMODE,required,notEmpty"`
}

type JWTConfig struct {
	Secret         string        `env:"JWT_SECRET,required,notEmpty"`
	Issuer         string        `env:"JWT_ISSUER" envDefault:"taggy-api"`
	AccessTokenTTL time.Duration `env:"JWT_ACCESS_TOKEN_TTL" envDefault:"15m"`
}

type AuthConfig struct {
	PasswordBcryptCost int           `env:"PASSWORD_BCRYPT_COST" envDefault:"12"`
	EmailOTPTTL        time.Duration `env:"AUTH_EMAIL_OTP_TTL" envDefault:"10m"`
}

type OAuthConfig struct {
	GoogleClientID     string `env:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret string `env:"GOOGLE_CLIENT_SECRET"`
	GoogleRedirectURL  string `env:"GOOGLE_REDIRECT_URL" envDefault:"http://localhost:8080/auth/google/callback"`
}

type LiveKitConfig struct {
	URL       string `env:"LIVEKIT_URL"`
	APIKey    string `env:"LIVEKIT_API_KEY"`
	APISecret string `env:"LIVEKIT_API_SECRET"`
}

type EmailConfig struct {
	ResendAPIKey string `env:"RESEND_API_KEY"`
	ResendFrom   string `env:"RESEND_EMAIL_FROM"`
}

type AdminConfig struct {
	Usernames string `env:"ADMIN_USERNAMES"`
}

type RazorpayConfig struct {
	KeyID         string `env:"RAZORPAY_KEY_ID"`
	KeySecret     string `env:"RAZORPAY_KEY_SECRET"`
	WebhookSecret string `env:"RAZORPAY_WEBHOOK_SECRET"`
	AmountPaise   int32  `env:"RAZORPAY_PREMIUM_AMOUNT_PAISE" envDefault:"49900"`
	Currency      string `env:"RAZORPAY_CURRENCY" envDefault:"INR"`
}

type CloudinaryConfig struct {
	URL string `env:"CLOUDINARY_URL"`
}

func LoadConfig() (*Config, error) {
	var cfg Config

	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (db DatabaseConfig) URL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		db.User,
		db.Password,
		db.Host,
		db.Port,
		db.Name,
		db.SSLMode,
	)
}
