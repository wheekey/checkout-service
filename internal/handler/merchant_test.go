package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"checkout-service/internal/domain"
	"checkout-service/internal/handler"
	"checkout-service/internal/repository"

	"github.com/stretchr/testify/assert"
)

// mockRepo реализует интерфейс merchantRepo из хендлера
type mockRepo struct {
	getErr   error
	merchant *repository.Merchant
}

func (m *mockRepo) GetByID(_ context.Context, _ string) (*repository.Merchant, error) {
	return m.merchant, m.getErr
}

func TestMerchantHandler_GetMerchant(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		path           string
		mock           mockRepo
		expectedStatus int
	}{
		{
			name:           "success: 200 OK",
			path:           "/merchants/merch_1",
			mock:           mockRepo{merchant: &repository.Merchant{ID: "merch_1", Name: "Test", Balance: 100}},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "empty id: 404 (pattern not matched)",
			path:           "/merchants/",
			mock:           mockRepo{},
			expectedStatus: http.StatusNotFound, // ← Меняем с 400 на 404
		},
		{
			name:           "not found: 404",
			path:           "/merchants/not_exists",
			mock:           mockRepo{getErr: domain.ErrNotFound},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "db error: 500",
			path:           "/merchants/merch_1",
			mock:           mockRepo{getErr: assert.AnError},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// 1. Мокаем репозиторий
			repo := &mockRepo{
				getErr:   tt.mock.getErr,
				merchant: tt.mock.merchant,
			}

			// 2. Создаём хендлер
			h := handler.NewMerchantHandler(repo)

			// 3. Создаём ServeMux и регистрируем паттерн (Go 1.22+ синтаксис)
			// Это критично: без роутера r.PathValue("id") вернёт пустую строку
			mux := http.NewServeMux()
			mux.HandleFunc("/merchants/{id}", h.GetMerchant)

			// 4. Готовим запрос
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			req.Header.Set("X-Correlation-ID", "test-correlation-id")
			rr := httptest.NewRecorder()

			// 5. Прогоняем запрос ЧЕРЕЗ РОУТЕР (не напрямую в хендлер!)
			mux.ServeHTTP(rr, req)

			// 6. Ассерты
			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
