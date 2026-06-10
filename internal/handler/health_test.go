package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"checkout-service/internal/handler"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockPinger — заглушка для интерфейса handler.Pinger.
// [Тестируемость] Позволяет эмулировать любое поведение БД (успех, ошибка, таймаут)
// без реальных сетевых вызовов. Делает тесты мгновенными и детерминированными.
type mockPinger struct {
	pingErr error         // Какую ошибку вернуть при вызове Ping()
	delay   time.Duration // Имитация задержки ответа БД
}

// Ping реализует интерфейс handler.Pinger.
func (m *mockPinger) Ping(ctx context.Context) error {
	if m.delay > 0 {
		// Эмулируем медленную БД, но с уважением к контексту
		select {
		case <-time.After(m.delay):
			return m.pingErr
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return m.pingErr
}

// TestHealthHandler_HandleLiveness проверяет, что liveness probe всегда возвращает 200.
// [Мониторинг] Liveness probe не должен проверять БД или другие зависимости.
func TestHealthHandler_HandleLiveness(t *testing.T) {
	t.Parallel()

	// Передаём nil, так как liveness не должен использовать БД вообще
	h := handler.NewHealthHandler(nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	h.HandleLiveness(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
	assert.JSONEq(t, `{"status":"ok"}`, rec.Body.String())
}

// TestHealthHandler_HandleReadiness проверяет readiness probe в разных сценариях.
// [Тестируемость] Используем табличный подход и мок для проверки всех веток логики.
func TestHealthHandler_HandleReadiness(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		mock           *mockPinger
		expectedStatus int
		checkBody      func(t *testing.T, body []byte)
	}{
		{
			name:           "database available",
			mock:           &mockPinger{pingErr: nil}, // БД отвечает успешно
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body []byte) {
				var resp map[string]string
				require.NoError(t, json.Unmarshal(body, &resp))
				assert.Equal(t, "ready", resp["status"])
			},
		},
		{
			name:           "database unavailable (connection refused)",
			mock:           &mockPinger{pingErr: errors.New("dial tcp 127.0.0.1:5432: connect: connection refused")},
			expectedStatus: http.StatusServiceUnavailable,
			checkBody: func(t *testing.T, body []byte) {
				var resp map[string]string
				require.NoError(t, json.Unmarshal(body, &resp))

				// [Тестируемость] Проверяем семантику, а не строковое представление JSON.
				// Порядок ключей в map[string]string в Go случайный, поэтому Contains с "{" опасен.
				assert.Equal(t, "error", resp["status"])
				assert.Contains(t, resp["error"], "connection refused")
			},
		},
		{
			name: "database timeout",
			// Мокаем БД, которая "думает" 5 секунд, но таймаут в хендлере всего 2 секунды
			mock:           &mockPinger{delay: 5 * time.Second, pingErr: nil},
			expectedStatus: http.StatusServiceUnavailable,
			checkBody: func(t *testing.T, body []byte) {
				var resp map[string]string
				require.NoError(t, json.Unmarshal(body, &resp))

				assert.Equal(t, "error", resp["status"])
				// Проверяем, что сработал именно таймаут контекста, а не ошибка БД
				assert.Contains(t, resp["error"], "context deadline exceeded")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// [Тестируемость] Внедряем зависимость (мок) через конструктор
			h := handler.NewHealthHandler(tt.mock)

			req := httptest.NewRequest(http.MethodGet, "/ready", nil)
			rec := httptest.NewRecorder()

			h.HandleReadiness(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
			assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))
			tt.checkBody(t, rec.Body.Bytes())
		})
	}
}
