package repository

import (
	"context"
)

// MerchantRepository определяет контракт для работы с мерчантами.
// Хендлер будет зависеть от этого интерфейса, а не от конкретной реализации.
type MerchantRepository interface {
	GetByID(ctx context.Context, id string) (*Merchant, error)
}
