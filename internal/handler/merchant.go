package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"checkout-service/internal/domain"
	"checkout-service/internal/repository"
)

type MerchantHandler struct {
	repo repository.MerchantRepository
}

// Конструктор теперь принимает интерфейс репозитория
func NewMerchantHandler(repo repository.MerchantRepository) *MerchantHandler {
	return &MerchantHandler{repo: repo}
}

func (h *MerchantHandler) GetMerchant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	corrID := r.Header.Get("X-Correlation-ID")
	logger := slog.With("correlation_id", corrID)

	// Получаем ID из URL (Go 1.22+)
	// Если у тебя 1.21, см. примечание ниже
	id := r.PathValue("id")
	if id == "" {
		h.writeError(w, http.StatusBadRequest, domain.ErrInvalidInput, logger)
		return
	}

	// Вызываем репозиторий напрямую
	merchant, err := h.repo.GetByID(ctx, id)
	if err != nil {
		// Маппинг ошибок БД на ошибки домена/HTTP
		// Если в репозитории вернули строку "merchant not found", можно проверять через strings.Contains,
		// но лучше, если репозиторий возвращает доменные ошибки (domain.ErrNotFound).
		if errors.Is(err, domain.ErrNotFound) || err.Error() == "merchant not found" {
			h.writeError(w, http.StatusNotFound, domain.ErrNotFound, logger)
		} else {
			// Любая другая ошибка (БД упала, сеть) -> 500
			logger.Error("Repository error", "err", err)
			h.writeError(w, http.StatusInternalServerError, domain.ErrInternalServer, logger)
		}
		return
	}

	// Успех: отдаём JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	// В реальном коде здесь: json.NewEncoder(w).Encode(merchant)
	// Для теста пока заглушка или раскомментируй, если нужно:
	// json.NewEncoder(w).Encode(merchant)
}

func (h *MerchantHandler) writeError(w http.ResponseWriter, code int, err error, logger *slog.Logger) {
	logger.Warn("Handler error", "error", err.Error(), "status", code)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	// Можно вернуть тело ошибки: json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}
