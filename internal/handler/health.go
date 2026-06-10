package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// Pinger — минимальный интерфейс для проверки доступности базы данных.
// [Архитектура] Принцип инверсии зависимостей (Dependency Inversion).
// Вместо того чтобы зависеть от конкретной реализации (*pgxpool.Pool),
// мы зависим от абстракции. Это позволяет легко подменять БД на мок в тестах.
//
// *pgxpool.Pool из github.com/jackc/pgx/v5/pgxpool уже реализует этот интерфейс,
// так как у неё есть метод Ping(ctx context.Context) error.
type Pinger interface {
	Ping(ctx context.Context) error
}

// HealthHandler инкапсулирует зависимости для health endpoints.
type HealthHandler struct {
	db Pinger // Теперь здесь интерфейс, а не конкретный тип
}

// NewHealthHandler создаёт новый HealthHandler.
// Принимает любой объект, реализующий интерфейс Pinger.
func NewHealthHandler(db Pinger) *HealthHandler {
	return &HealthHandler{db: db}
}

// HandleLiveness — "Сервис жив?" (Kubernetes liveness probe).
// [Мониторинг] Если этот endpoint не отвечает, K8s считает, что под завис,
// и перезапускает его. Поэтому здесь НЕ должно быть тяжёых проверок (БД, внешние API).
//
// Почему НЕ проверяем БД в liveness:
//  1. Если БД медленная (но живая), liveness начнёт фейлиться.
//  2. K8s перезапустит под, хотя проблема не в нём.
//  3. Новый под снова попытается подключиться к медленной БД.
//  4. Получаем thundering herd и усугубление проблемы.
//
// Правила для liveness:
//   - Возвращай 200, если процесс жив и может обрабатывать запросы.
//   - НЕ проверяй БД, Redis, внешние сервисы.
//   - Должен отвечать за <100мс.
func (h *HealthHandler) HandleLiveness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

// HandleReadiness — "Сервис готов принимать трафик?" (Kubernetes readiness probe).
// [Мониторинг] Если этот endpoint возвращает не-200, K8s ВРЕМЕННО убирает под из Service
// (не отправляет трафик), но НЕ перезапускает его. Это позволяет переждать сбой БД.
//
// Правила для readiness:
//   - Проверяй критические зависимости (БД, Redis).
//   - Устанавливай таймаут, чтобы не блокировать probe надолго.
//   - Возвращай детальную информацию о том, что именно недоступно.
func (h *HealthHandler) HandleReadiness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// [Отказоустойчивость] Таймаут для проверки БД.
	// Без него readiness может зависнуть на 30с (дефолтный таймаут драйвера),
	// и K8s будет считать под "неготовым" слишком долго.
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	// [Архитектура] Вызываем Ping() через интерфейс Pinger.
	// В проде это будет *pgxpool.Pool, в тестах — mockPinger.
	if err := h.db.Ping(ctx); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		resp := map[string]string{
			"status": "error",
			"error":  "database unavailable: " + err.Error(),
		}
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	w.WriteHeader(http.StatusOK)
	resp := map[string]string{"status": "ready"}
	_ = json.NewEncoder(w).Encode(resp)
}
