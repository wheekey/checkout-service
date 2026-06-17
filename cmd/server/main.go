package main

import (
	"checkout-service/internal/db"
	"log"
	"log/slog"
	"os"
	"runtime" // 🆕 Добавь этот импорт

	"checkout-service/internal/config"
	"checkout-service/internal/server"
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

	if err := server.Run(cfg); err != nil {
		slog.Error("Server exited with error", "err", err)
		os.Exit(1)
	}
}
