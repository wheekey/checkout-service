package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"checkout-service/internal/domain"
	"checkout-service/internal/repository"
)

// === ИНТЕРФЕЙС ЗДЕСЬ (потребитель) ===
// Он описывает, что хендлеру нужно от репозитория
type merchantRepo interface {
	GetByID(ctx context.Context, id string) (*repository.Merchant, error)
}

type MerchantHandler struct {
	repo merchantRepo
}

// Конструктор теперь принимает интерфейс репозитория
func NewMerchantHandler(repo merchantRepo) *MerchantHandler {
	return &MerchantHandler{repo: repo}
}

func (h *MerchantHandler) GetMerchant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	corrID := r.Header.Get("X-Correlation-ID")
	logger := slog.With("correlation_id", corrID)

	id := r.PathValue("id")
	if id == "" {
		h.writeError(w, http.StatusBadRequest, domain.ErrInvalidInput, logger)
		return
	}

	// 1. Получаем данные
	merch, err := h.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) || err.Error() == "merchant not found" {
			h.writeError(w, http.StatusNotFound, domain.ErrNotFound, logger)
		} else {
			logger.Error("Repository error", "err", err)
			h.writeError(w, http.StatusInternalServerError, domain.ErrInternalServer, logger)
		}
		return
	}

	// 2. Отдаём успешный ответ с данными
	// Переименовал переменную в 'merch', чтобы не путалась с пакетом 'merchant'
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(merch); err != nil {
		logger.Error("Failed to encode response", "err", err)
	}
}

func (h *MerchantHandler) writeError(w http.ResponseWriter, code int, err error, logger *slog.Logger) {
	logger.Warn("Handler error", "error", err.Error(), "status", code)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	// Можно вернуть тело ошибки: json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
}
