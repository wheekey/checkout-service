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

// === НОВЫЙ МЕТОД: Списание средств с блокировкой ===
func (r *MerchantRepo) DeductBalance(ctx context.Context, id string, amount int) error {
	// 1. Начинаем транзакцию
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	// 2. Гарантируем откат, если что-то пойдет не так (или если Commit не вызовется)
	defer tx.Rollback(ctx)

	var currentBalance int

	// 3. SELECT FOR UPDATE — блокирует строку для других транзакций!
	// Другие запросы к этой строке будут ждать, пока мы не сделаем Commit или Rollback.
	err = tx.QueryRow(ctx, `
		SELECT balance FROM merchants WHERE id = $1 FOR UPDATE
	`, id).Scan(&currentBalance)

	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("merchant not found")
		}
		return fmt.Errorf("select for update failed: %w", err)
	}

	// 4. Проверка бизнес-логики
	if currentBalance < amount {
		return domain.ErrInsufficientBalance
	}

	// 5. Обновляем баланс
	_, err = tx.Exec(ctx, `
		UPDATE merchants SET balance = balance - $1, updated_at = NOW() WHERE id = $2
	`, amount, id)

	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	// 6. Коммитим изменения (снимаем блокировку)
	return tx.Commit(ctx)
}

type Merchant struct {
	ID      string
	Name    string
	Balance int
}
