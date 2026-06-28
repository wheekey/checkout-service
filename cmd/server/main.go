package main

import (
	"checkout-service/internal/config"
	"checkout-service/internal/db"
	"checkout-service/internal/server"
	"checkout-service/internal/worker"
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"runtime" // 🆕 Добавь этот импорт
	"syscall"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("❌ Ошибка конфига: ", err)
	}

	slog.Info("Loading configuration", "port", cfg.Port, "log_level", cfg.LogLevel)

	// 🆕 [Go Internals] Логирование параметров рантайма при старте
	// Это помогает при диагностике проблем в продакшене:
	// - Если GOMAXPROCS != NumCPU, возможно, контейнер ограничил CPU
	// - Если NumGoroutine слишком большой на старте, есть утечка при инициализации
	// - GoVersion помогает понять, какие фичи рантайма доступны
	slog.Info("Runtime stats at startup",
		"GOMAXPROCS", runtime.GOMAXPROCS(0), // 0 возвращает текущее значение, не меняя его
		"NumCPU", runtime.NumCPU(),
		"NumGoroutine", runtime.NumGoroutine(),
		"GoVersion", runtime.Version(),
	)

	// Миграции
	if err := db.RunMigrations(cfg.DBURL); err != nil {
		slog.Error("Migration failed", "err", err)
		os.Exit(1)
	}
	slog.Info("Migrations applied successfully")

	// 🆕 Создаём контекст с обработкой сигналов
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 🆕 Запуск воркер-пула ВЫШЕ server.Run()
	pool := worker.NewPool(4, 100)
	pool.Start(ctx)
	defer pool.Stop() // Выполнится при выходе из main()

	// Пример: отправляем задачу
	go func() {
		pool.Submit(worker.Task{ID: "order_123", Amount: 1000})
	}()

	// Пример: читаем результаты
	go func() {
		for result := range pool.Results() {
			slog.Info("Task completed",
				"task_id", result.TaskID,
				"output", result.Output,
				"error", result.Err)
		}
	}()

	if err := server.Run(cfg); err != nil {
		slog.Error("Server exited with error", "err", err)
		os.Exit(1)
	}
}
