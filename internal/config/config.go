package config

import (
	"fmt"
	"log/slog"
	"os"
	"time"
)

type Config struct {
	Port         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	LogLevel     slog.Level
	DBURL        string
}

func Load() (*Config, error) {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	cfg := &Config{
		Port:         port,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
		LogLevel:     slog.LevelInfo,
	}

	// === ВАЛИДАЦИЯ ===
	if cfg.Port == "" {
		return nil, fmt.Errorf("порт не может быть пустым")
	}
	if cfg.ReadTimeout <= 0 || cfg.WriteTimeout <= 0 {
		return nil, fmt.Errorf("таймауты должны быть больше 0")
	}

	cfg.DBURL = os.Getenv("DATABASE_URL")
	if cfg.DBURL == "" {
		cfg.DBURL = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	}

	return cfg, nil
}
