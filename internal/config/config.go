package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env                string
	Host               string
	Port               string
	DBHost             string
	DBPort             string
	DBUser             string
	DBPassword         string
	DBName             string
	DBSSLMode          string
	JWTSecret          string
	JWTExpirationHours int
	LogLevel           string
}

func Load() (*Config, error) {
	// Load .env if present (ignore error in production)
	_ = godotenv.Load()

	cfg := &Config{
		Env:                getEnv("APP_ENV", "development"),
		Host:               getEnv("APP_HOST", "0.0.0.0"),
		Port:               getEnv("APP_PORT", "8080"),
		DBHost:             getEnv("DB_HOST", "localhost"),
		DBPort:             getEnv("DB_PORT", "5432"),
		DBUser:             getEnv("DB_USER", "notes"),
		DBPassword:         getEnv("DB_PASSWORD", "notes"),
		DBName:             getEnv("DB_NAME", "notes"),
		DBSSLMode:          getEnv("DB_SSLMODE", "disable"),
		JWTSecret:          getEnv("JWT_SECRET", ""),
		JWTExpirationHours: getEnvAsInt("JWT_EXPIRATION_HOURS", 24),
		LogLevel:           getEnv("LOG_LEVEL", "info"),
	}

	if cfg.JWTSecret == "" || len(cfg.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be set and at least 32 characters long")
	}

	return cfg, nil
}

func (c *Config) DatabaseURL() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode,
	)
}

func (c *Config) HTTPAddr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func (c *Config) JWTExpiration() time.Duration {
	return time.Duration(c.JWTExpirationHours) * time.Hour
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	if value, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return fallback
}
