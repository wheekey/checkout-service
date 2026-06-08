package client_test

import (
	"context"
	"testing"
	"time"

	"checkout-service/internal/client"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPaymentClient_Charge проверяет логику retry, backoff и отмены контекста.
// [Тестируемость] Используем детерминированный testFailCount вместо rand, чтобы тесты не "мигали".
func TestPaymentClient_Charge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		failCount     int // Сколько раз симулировать ошибку перед успехом
		ctxTimeout    time.Duration
		expectedError string // Ожидаемая подстрока в ошибке (пустая = успех)
	}{
		{
			name:          "success on first attempt",
			failCount:     0,
			ctxTimeout:    1 * time.Second,
			expectedError: "",
		},
		{
			name:          "success after 2 retries",
			failCount:     2, // 2 ошибки, 3-я попытка успешна (maxRetries = 3)
			ctxTimeout:    1 * time.Second,
			expectedError: "",
		},
		{
			name:          "max retries exceeded",
			failCount:     10, // Больше, чем maxRetries (3), гарантируем исчерпание попыток
			ctxTimeout:    1 * time.Second,
			expectedError: "operation failed after 3 retries",
		},
		{
			// [Go Internals] Критически важный тест: проверяем, что select с ctx.Done() работает.
			// Если бы мы использовали time.Sleep() вместо select, этот тест завис бы навсегда.
			name:          "context canceled during backoff",
			failCount:     5,                     // Гарантируем, что уйдём в backoff
			ctxTimeout:    15 * time.Millisecond, // Таймаут меньше, чем время первого backoff (10ms + jitter)
			expectedError: "operation cancelled during backoff: context deadline exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// 1. Создаём тестовый клиент с контролируемым поведением
			c := client.NewTestPaymentClient(tt.failCount)

			// 2. Настраиваем контекст с нужным таймаутом
			ctx, cancel := context.WithTimeout(context.Background(), tt.ctxTimeout)
			defer cancel() // [Go Internals] Здесь defer уместен, так как контекст один на всю функцию теста

			// 3. Вызываем метод
			err := c.Charge(ctx, "order_123", 1000)

			// 4. Проверяем результат
			if tt.expectedError == "" {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			}
		})
	}
}
