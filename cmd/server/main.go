package main

import (
	"log/slog"
	"os"

	"github.com/wheekey/checkout-service/internal/config"
	"github.com/wheekey/checkout-service/internal/server"
)

func main() {
	cfg := config.Load()

	slog.Info("Loading configuration", "port", cfg.Port, "log_level", cfg.LogLevel)

	if err := server.Run(cfg); err != nil {
		slog.Error("Server exited with error", "err", err)
		os.Exit(1)
	}
}
