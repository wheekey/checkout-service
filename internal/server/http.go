package server

import (
	"checkout-service/internal/config"
	"context"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"checkout-service/internal/db"
	"checkout-service/internal/handler"
	"checkout-service/internal/middleware"
	"checkout-service/internal/repository"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Run(cfg *config.Config) error {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))
	slog.SetDefault(logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Подключаем БД
	pool, err := db.NewPool(ctx, cfg.DBURL)
	if err != nil {
		slog.Error("DB connection failed", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	// 2. Инициализируем зависимости
	merchantRepo := repository.NewMerchantRepo(pool)
	merchantHandler := handler.NewMerchantHandler(merchantRepo)
	healthHandler := handler.NewHealthHandler(pool)

	// 3. Бизнес-роуты
	businessMux := http.NewServeMux()
	businessMux.HandleFunc("GET /api/v1/merchants/{id}", merchantHandler.GetMerchant)

	// Оборачиваем в middleware метрик
	// [Мониторинг] Middleware применяется только к бизнес-роутам,
	// чтобы служебные endpoints (health, metrics) не засоряли метрики.
	businessHandler := middleware.Metrics(businessMux)

	// 4. [Мониторинг] Admin-роуты (health, metrics, pprof)
	// [Безопасность] Эти endpoints не должны быть доступны из интернета.
	// В Kubernetes настраивается NetworkPolicy, чтобы доступ был только из:
	//   - /health, /ready — только от kubelet
	//   - /metrics — только от Prometheus
	//   - /debug/pprof — только от разработчиков (через port-forward)
	adminMux := http.NewServeMux()
	adminMux.HandleFunc("GET /health", healthHandler.HandleLiveness)
	adminMux.HandleFunc("GET /ready", healthHandler.HandleReadiness)

	// [Мониторинг] Prometheus endpoint.
	// promhttp.Handler() автоматически собирает все метрики из default registry.
	// Это включает:
	//   - Наши кастомные метрики (httpRequestsTotal, httpRequestDuration)
	//   - Встроенные метрики Go runtime (go_goroutines, go_gc_duration_seconds)
	//   - Метрики процесса (process_cpu_seconds_total, process_resident_memory_bytes)
	adminMux.Handle("GET /metrics", promhttp.Handler())

	// [Мониторинг] pprof endpoints для профилирования.
	// Это встроенный инструмент Go, который позволяет в рантайме посмотреть:
	//   - CPU profile (какие функции едят CPU)
	//   - Heap profile (сколько памяти аллоцировано)
	//   - Goroutine profile (сколько горутин, нет ли утечек)
	//   - Block/mutex profile (где блокировки)
	//
	// Использование в проде:
	//   go tool pprof http://localhost:8081/debug/pprof/heap
	//   go tool pprof http://localhost:8081/debug/pprof/profile?seconds=30
	//
	// ⚠️ Важно: pprof может быть security risk, если доступен извне.
	// Всегда закрывай доступ через NetworkPolicy или авторизацию.
	// 📖 Изучить: https://go.dev/doc/diagnostics#profiling
	adminMux.HandleFunc("GET /debug/pprof/", pprof.Index)
	adminMux.HandleFunc("GET /debug/pprof/cmdline", pprof.Cmdline)
	adminMux.HandleFunc("GET /debug/pprof/profile", pprof.Profile)
	adminMux.HandleFunc("GET /debug/pprof/symbol", pprof.Symbol)
	adminMux.HandleFunc("GET /debug/pprof/trace", pprof.Trace)

	// 5. Бизнес-сервер
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      businessHandler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	// 6. [Мониторинг] Admin-сервер (health, metrics, pprof)
	// [Мониторинг] Почему отдельный сервер на отдельном порту:
	//   1. Безопасность: служебные endpoints не должны быть доступны из интернета
	//   2. Производительность: метрики не проходят через бизнес-middleware (auth, logging)
	//   3. SLO: метрики должны работать, даже если бизнес-логика под нагрузкой
	//   4. Rate limiting: можно лимитировать бизнес-порт, не трогая служебный
	//   5. NetworkPolicy: в K8s можно закрыть доступ к admin-порту извне
	adminPort := "8083" // TODO: вынести в конфиг
	adminSrv := &http.Server{
		Addr:    ":" + adminPort,
		Handler: adminMux,
		// [Отказоустойчивость] Таймауты для admin-сервера тоже важны,
		// но обычно менее строгие, т.к. трафика меньше
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// 7. Запуск серверов
	go func() {
		slog.Info("Starting business server", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			slog.Error("Business server failed", "err", err)
			os.Exit(1)
		}
	}()

	go func() {
		slog.Info("Starting admin server", "addr", adminSrv.Addr)
		if err := adminSrv.ListenAndServe(); err != http.ErrServerClosed {
			slog.Error("Admin server failed", "err", err)
		}
	}()

	// 8. [Отказоустойчивость] Graceful shutdown
	// [Go Internals] signal.Notify регистрирует канал для получения сигналов ОС.
	// SIGINT — Ctrl+C, SIGTERM — команда на завершение (от K8s, systemd, docker stop).
	//
	// Почему buffer = 1:
	//   - Если буфер пуст, а сигнал пришёл, он теряется
	//   - Buffer = 1 гарантирует, что хотя бы один сигнал будет сохранён
	//   - 📖 Изучить: https://pkg.go.dev/os/signal#Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutdown signal received. Draining connections...")

	// [Отказоустойчивость] Graceful shutdown с таймаутом.
	// [Go Internals] srv.Shutdown():
	//   1. Закрывает listener → новые запросы не принимаются
	//   2. Ждёт завершения активных запросов (или истечения таймаута)
	//   3. Закрывает все idle-соединения
	//
	// Почему 10 секунд:
	//   - Это время, за которое активные запросы должны завершиться
	//   - Если запрос длится дольше — он будет оборван (клиент получит ошибку)
	//   - В платёжных системах может быть больше (30с), чтобы не оборвать транзакцию
	//
	// ⚠️ Важно: WriteTimeout сервера должен быть <= таймаута shutdown,
	// иначе активные запросы могут зависнуть.
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// [Go Internals] Вызовы Shutdown параллельны, можно не бояться порядка.
	// Но обычно сначала останавливают admin-сервер, чтобы K8s сразу увидел,
	// что под "не готов" и перестал слать трафик.
	if err := adminSrv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Admin server forced shutdown", "err", err)
	}
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Business server forced shutdown", "err", err)
	}

	slog.Info("Servers stopped gracefully")
	return nil
}
