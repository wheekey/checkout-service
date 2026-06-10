package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// [Мониторинг] RED-метрики для HTTP-хендлеров (Rate, Errors, Duration).
// Это стандарт индустрии, описанный в книге "Site Reliability Engineering" (Google).
//
// Почему именно эти три метрики:
//   - Rate (RPS) — показывает нагрузку на сервис
//   - Errors (% ошибок) — показывает качество сервиса
//   - Duration (латентность) — показывает производительность
//
// Вместе они позволяют построить SLO (Service Level Objectives):
// например, "99% запросов должны выполняться быстрее 500мс".
// 📖 Изучить: https://www.weave.works/blog/the-red-method-key-metrics-for-microservices-architecture/

var (
	// httpRequestsTotal — счётчик всех HTTP-запросов с разбивкой по method/path/status.
	// [Мониторинг] Counter — только растёт (сбрасывается при перезапуске сервиса).
	// Используй rate() в PromQL для получения RPS.
	//
	// Labels (method, path, status) позволяют группировать данные:
	//   - rate(checkout_http_requests_total{status="500"}[5m]) — RPS ошибок
	//   - sum by (path) (rate(checkout_http_requests_total[5m])) — RPS по endpoint'ам
	//
	// ⚠️ Важно: не используй high-cardinality labels (например, user_id),
	// иначе Prometheus захлебнётся от количества time series.
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "checkout",       // Префикс для всех метрик сервиса
			Subsystem: "http",           // Подсистема
			Name:      "requests_total", // Название метрики
			Help:      "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"}, // Labels для группировки
	)

	// httpRequestDuration — гистограмма времени выполнения запросов.
	// [Мониторинг] Histogram — распределяет значения по бакетам.
	// Позволяет считать перцентили (p50, p95, p99) через histogram_quantile в PromQL.
	//
	// Пример PromQL для p95:
	//   histogram_quantile(0.95, rate(checkout_http_request_duration_seconds_bucket[5m]))
	//
	// Почему такие бакеты (DefBuckets = [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10]):
	//   - 5ms, 10ms — очень быстрые запросы (кэш, простые операции)
	//   - 25ms, 50ms, 100ms — обычные запросы
	//   - 250ms, 500ms, 1s — медленные запросы (нужна оптимизация)
	//   - 2.5s, 5s, 10s — очень медленные (почти всегда проблема)
	//
	// ⚠️ Важно: если твои запросы обычно длятся 100-500мс, добавь больше бакетов в этом диапазоне:
	//   Buckets: []float64{0.1, 0.2, 0.3, 0.4, 0.5, 1, 2, 5}
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "checkout",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "HTTP request duration in seconds",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	// httpRequestsInFlight — gauge количества запросов в обработке прямо сейчас.
	// [Мониторинг] Gauge — может расти и падать.
	// Полезно для понимания текущей нагрузки и поиска "узких горлышек".
	//
	// Если requests_in_flight постоянно растёт — значит, хендлеры не успевают обрабатывать запросы.
	// Это сигнал к масштабированию или оптимизации.
	httpRequestsInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "checkout",
			Subsystem: "http",
			Name:      "requests_in_flight",
			Help:      "Number of HTTP requests currently being processed",
		},
	)
)

// responseWriter — обёртка над http.ResponseWriter для перехвата status code.
// [Go Internals] Стандартный http.ResponseWriter не позволяет узнать, какой status code
// был записан. Мы оборачиваем его, чтобы сохранить код и передать в метрики.
//
// Почему это нужно:
//   - Prometheus-метрики должны включать status code (200, 404, 500 и т.д.)
//   - Без перехвата мы не сможем посчитать % ошибок
//   - Это стандартный паттерн в Go middleware
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

// WriteHeader перехватывает вызов, чтобы запомнить status code.
// [Go Internals] Важно: WriteHeader может быть вызван только ОДИН раз.
// Повторный вызов приведёт к предупреждению в логах и игнорированию.
// Флаг `written` защищает от этой проблемы.
func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
	}
	rw.ResponseWriter.WriteHeader(code)
}

// Write перехватывает запись тела, чтобы автоматически вызвать WriteHeader(200),
// если он ещё не был вызван явно (стандартное поведение http.ResponseWriter).
// [Go Internals] Если хендлер пишет в ResponseWriter без вызова WriteHeader,
// Go автоматически вызывает WriteHeader(200). Мы эмулируем это поведение.
func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.written {
		rw.statusCode = http.StatusOK
		rw.written = true
	}
	return rw.ResponseWriter.Write(b)
}

// Metrics — middleware, который записывает RED-метрики для каждого HTTP-запроса.
// [Мониторинг] Middleware — это "перехватчик", который выполняется до и после хендлера.
// Паттерн: сделать пролог → вызвать следующий хендлер → сделать эпilog.
//
// Как это работает:
//  1. Пролог: увеличиваем счётчик in-flight, запоминаем время начала
//  2. Вызываем следующий хендлер в цепочке (next.ServeHTTP)
//  3. Эпилог: записываем метрики (total, duration), уменьшаем in-flight
//
// Почему defer для Dec():
//   - Гарантирует, что счётчик уменьшится даже при панике в хендлере
//   - Иначе requests_in_flight будет расти бесконечно (утечка метрик)
func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// [Мониторинг] Увеличиваем счётчик "запросов в обработке"
		httpRequestsInFlight.Inc()
		defer httpRequestsInFlight.Dec() // Гарантируем уменьшение при выходе

		// Оборачиваем ResponseWriter для перехвата status code
		rw := newResponseWriter(w)

		// Вызываем следующий хендлер в цепочке
		next.ServeHTTP(rw, r)

		// [Мониторинг] После выполнения хендлера записываем метрики
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(rw.statusCode)
		path := r.URL.Path // ⚠️ В проде лучше нормализовать path (убрать ID)

		httpRequestsTotal.WithLabelValues(r.Method, path, status).Inc()
		httpRequestDuration.WithLabelValues(r.Method, path, status).Observe(duration)
	})
}
