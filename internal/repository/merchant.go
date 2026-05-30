package repository

import (
	"checkout-service/internal/domain"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MerchantRepo struct {
	pool *pgxpool.Pool
}

func NewMerchantRepo(pool *pgxpool.Pool) *MerchantRepo {
	return &MerchantRepo{pool: pool}
}

func (r *MerchantRepo) GetByID(ctx context.Context, id string) (*Merchant, error) {
	var m Merchant
	err := r.pool.QueryRow(ctx, `
		SELECT id, name, balance FROM merchants WHERE id = $1
	`, id).Scan(&m.ID, &m.Name, &m.Balance)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("query failed: %w", err)
	}

	return &m, nil
}

type Merchant struct {
	ID      string
	Name    string
	Balance int
}
