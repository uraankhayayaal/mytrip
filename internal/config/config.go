package config

import (
	"errors"
	"fmt"
	"os"
	"time"
)

// Config — конфигурация сервера mytrip (заполняется из env).
type Config struct {
	Port          string
	DatabaseURL   string
	JWTSecret     string
	JWTAccessTTL  time.Duration // default 15m
	JWTRefreshTTL time.Duration // default 30d
	S3Endpoint    string
	S3Bucket      string
	S3AccessKey   string
	S3SecretKey   string
}

// Load читает конфигурацию из переменных окружения.
// Ошибка, если DATABASE_URL или JWT_SECRET пустые.
func Load() (*Config, error) {
	accessTTL, err := durationEnv("JWT_ACCESS_TTL", 15*time.Minute)
	if err != nil {
		return nil, err
	}
	refreshTTL, err := durationEnv("JWT_REFRESH_TTL", 720*time.Hour)
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Port:          envOr("PORT", "8080"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		JWTAccessTTL:  accessTTL,
		JWTRefreshTTL: refreshTTL,
		S3Endpoint:    os.Getenv("S3_ENDPOINT"),
		S3Bucket:      envOr("S3_BUCKET", "mytrip"),
		S3AccessKey:   os.Getenv("S3_ACCESS_KEY"),
		S3SecretKey:   os.Getenv("S3_SECRET_KEY"),
	}

	var missing []string
	if cfg.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if cfg.JWTSecret == "" {
		missing = append(missing, "JWT_SECRET")
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("config: missing required env: %v", missing)
	}
	return cfg, nil
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func durationEnv(key string, def time.Duration) (time.Duration, error) {
	v := os.Getenv(key)
	if v == "" {
		return def, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("config: %s: %w", key, err)
	}
	return d, nil
}

var _ = errors.New // placeholder, чтобы не было пустого импорта при рефакторинге
