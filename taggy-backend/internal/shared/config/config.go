package config

import (
	"fmt"

	"github.com/caarlos0/env/v11"
	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	App AppConfig
	DB  DatabaseConfig `envPrefix:"DB_"`
}

type AppConfig struct {
    Environment string `env:"APP_ENV,required,notEmpty"`
    Port        string `env:"PORT,required,notEmpty"`
    LogLevel    string `env:"LOG_LEVEL" envDefault:"info"`
}

type DatabaseConfig struct {
	Host     string `env:"HOST,required,notEmpty"`
	Port     string `env:"PORT,required,notEmpty"`
	User     string `env:"USER,required,notEmpty"`
	Password string `env:"PASSWORD,required,notEmpty"`
	Name     string `env:"NAME,required,notEmpty"`
	SSLMode  string `env:"SSLMODE,required,notEmpty"`
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