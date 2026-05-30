package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, url string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to create pool: %w", err)
	}

	// Проверка подключения (ping)
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Настройка пула под продакшен
	pool.Config().MaxConns = 20
	pool.Config().MinConns = 2
	pool.Config().MaxConnLifetime = time.Hour
	pool.Config().MaxConnIdleTime = 10 * time.Minute

	return pool, nil
}
