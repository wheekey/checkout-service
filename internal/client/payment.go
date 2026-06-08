package client

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"time"
)

// PaymentClient имитирует вызов внешнего платёжного шлюза (или любого внешнего HTTP-сервиса).
// [Отказоустойчивость + Интеграции] Реализует паттерны: Context Timeout, Exponential Backoff + Jitter.
type PaymentClient struct {
	httpClient *http.Client
	maxRetries int
	baseDelay  time.Duration
}

// NewPaymentClient создаёт новый клиент с безопасными настройками по умолчанию.
func NewPaymentClient() *PaymentClient {
	return &PaymentClient{
		httpClient: &http.Client{
			// [Highload / Масштабируемость] Глобальный таймаут защищает от "зависших" соединений
			// и предотвращает утечку файловых дескрипторов при сбоях внешней системы.
			// В реальном проекте здесь также настраивается http.Transport (MaxIdleConns, TLSHandshakeTimeout).
			Timeout: 5 * time.Second,
		},
		maxRetries: 3,
		baseDelay:  100 * time.Millisecond,
	}
}

// Charge пытается выполнить операцию (например, списать средства).
// [REST API / Go Internals] Контекст передаётся явно первым аргументом для управления
// временем жизни запроса и его корректной отмены (graceful degradation).
func (c *PaymentClient) Charge(ctx context.Context, orderID string, amount int) error {
	var lastErr error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		// [Отказоустойчивость] Дочерний контекст с таймаутом для ОДНОЙ попытки.
		// Если общий ctx отменится (например, клиент разорвал соединение), этот тоже отменится.
		reqCtx, cancel := context.WithTimeout(ctx, 2*time.Second)

		err := c.doCharge(reqCtx, orderID, amount)

		// Освобождаем ресурсы немедленно после попытки, чтобы не ждать истечения таймаута дочернего контекста.
		cancel()

		if err == nil {
			return nil // Успех, выходим из цикла
		}

		lastErr = err
		slog.Warn("Operation attempt failed", "order_id", orderID, "attempt", attempt+1, "err", err)

		// Не делаем retry после последней попытки.
		if attempt == c.maxRetries {
			break
		}

		// [Масштабируемость] Exponential Backoff + Jitter.
		// Без jitter все клиенты, получившие ошибку одновременно, отправят повторный запрос в одну миллисекунду,
		// создавая пиковую нагрузку и добивая сервис (эффект Thundering Herd).
		backoff := c.baseDelay * (1 << attempt)                  // 100ms, 200ms, 400ms...
		jitter := time.Duration(rand.Int63n(int64(c.baseDelay))) // 0..100ms
		sleepTime := backoff + jitter

		slog.Info("Retrying after backoff", "sleep", sleepTime)

		// [Go Internals] Прерываемый sleep.
		// Мы не используем time.Sleep(), потому что он блокирует горутину и игнорирует отмену контекста.
		select {
		case <-time.After(sleepTime):
			// Время вышло, продолжаем цикл retry
		case <-ctx.Done():
			// Общий контекст был отменён (например, graceful shutdown или таймаут пользователя)
			return fmt.Errorf("operation cancelled during backoff: %w", ctx.Err())
		}
	}

	return fmt.Errorf("operation failed after %d retries: %w", c.maxRetries, lastErr)
}

// doCharge имитирует сетевой вызов к внешнему сервису.
func (c *PaymentClient) doCharge(ctx context.Context, orderID string, amount int) error {
	// Имитация задержки сети (например, 500 мс)
	time.Sleep(500 * time.Millisecond)

	// Проверка, не был ли контекст отменён во время ожидания "сети"
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// Имитация случайной ошибки (например, 30% шанс получить 503 Service Unavailable)
	// Это позволяет протестировать работу retry-логики в тестах
	if rand.Intn(10) < 3 {
		return fmt.Errorf("503 Service Unavailable")
	}

	return nil
}
