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
	httpClient    *http.Client
	maxRetries    int
	baseDelay     time.Duration
	testFailCount int // [Тестируемость] Позволяет детерминировано управлять числом сбоев в тестах (0 = успех сразу)
}

// NewPaymentClient создаёт новый клиент с безопасными настройками по умолчанию для продакшена.
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

// NewTestPaymentClient создаёт клиент специально для юнит-тестов.
// Позволяет точно задать, сколько раз метод должен вернуть ошибку перед успехом.
func NewTestPaymentClient(failCount int) *PaymentClient {
	c := NewPaymentClient()
	c.testFailCount = failCount
	c.baseDelay = 10 * time.Millisecond // Ускоряем тесты, чтобы не ждать реальные секунды
	return c
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

		// [Go Internals] Освобождаем ресурсы НЕМЕДЛЕННО после попытки.
		// Если не вызвать cancel(), внутри рантайма останется жить time.Timer и горутина до истечения 2 секунд.
		// При высокой нагрузке и циклах retry это приведёт к утечке памяти и горутин.
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
		// создавая пиковую нагрузку и добивая восстанавливающийся сервис (эффект Thundering Herd).
		backoff := c.baseDelay * (1 << attempt)                  // 10ms, 20ms, 40ms... (в тестах baseDelay уменьшен)
		jitter := time.Duration(rand.Int63n(int64(c.baseDelay))) // 0..baseDelay
		sleepTime := backoff + jitter

		slog.Info("Retrying after backoff", "sleep", sleepTime)

		// [Go Internals] Прерываемый sleep.
		// Мы НЕ используем time.Sleep(), потому что он блокирует горутину и игнорирует отмену контекста.
		// select позволяет мгновенно прервать ожидание, если внешний контекст был отменён.
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
	// Имитация задержки сети (в тестах ускорена до 10мс)
	time.Sleep(10 * time.Millisecond)

	// Проверка, не был ли контекст отменён во время ожидания "сети"
	if ctx.Err() != nil {
		return ctx.Err()
	}

	// [Тестируемость] Если это тест и мы ещё не исчерпали лимит фейков, возвращаем ошибку.
	// Это позволяет писать детерминированные тесты без reliance на rand.Intn.
	if c.testFailCount > 0 {
		c.testFailCount--
		return fmt.Errorf("503 Service Unavailable")
	}

	// [Отказоустойчивость] В продакшене (testFailCount == 0) работает эмуляция случайных сбоев,
	// чтобы проверить, что retry-логика действительно срабатывает в боевых условиях.
	if rand.Intn(10) < 3 {
		return fmt.Errorf("503 Service Unavailable")
	}

	return nil
}
