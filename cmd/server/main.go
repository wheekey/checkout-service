package main

import (
	"checkout-service/internal/db"
	"log"
	"log/slog"
	"os"

	"checkout-service/internal/config"
	"checkout-service/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("❌ Ошибка конфига: ", err)
	}

	slog.Info("Loading configuration", "port", cfg.Port, "log_level", cfg.LogLevel)

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
