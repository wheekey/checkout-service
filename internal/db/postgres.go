package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PoolConfig — конфигурация пула соединений
// Вынесли в отдельную структуру, чтобы можно было настраивать извне
type PoolConfig struct {
	URL             string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

// DefaultPoolConfig возвращает production-настройки пула
func DefaultPoolConfig(url string) PoolConfig {
	return PoolConfig{
		URL:             url,
		MaxConns:        20,              // Максимум одновременных соединений
		MinConns:        2,               // Минимум "тёплых" соединений
		MaxConnLifetime: 5 * time.Minute, // Пересоздаём соединения для балансировки
		MaxConnIdleTime: 1 * time.Minute, // Закрываем неиспользуемые
	}
}

// NewPool создаёт подключение к БД с production-настройками
func NewPool(ctx context.Context, cfg PoolConfig) (*pgxpool.Pool, error) {
	// 1. Парсим URL в конфиг
	config, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// 2. Применяем настройки ДО создания пула
	config.MaxConns = cfg.MaxConns
	config.MinConns = cfg.MinConns
	config.MaxConnLifetime = cfg.MaxConnLifetime
	config.MaxConnIdleTime = cfg.MaxConnIdleTime

	// 3. Создаём пул с настроенным конфигом
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	// 4. Проверяем подключение
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	slog.Info("Database connected",
		"max_conns", cfg.MaxConns,
		"min_conns", cfg.MinConns,
		"max_lifetime", cfg.MaxConnLifetime,
		"max_idle_time", cfg.MaxConnIdleTime,
	)

	return pool, nil
}
