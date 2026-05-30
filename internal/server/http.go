package server

import (
	"checkout-service/internal/config"
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"checkout-service/internal/db"
	"checkout-service/internal/handler"
	"checkout-service/internal/repository"
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

	// 3. Роуты
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handler.Health)
	mux.HandleFunc("GET /api/v1/merchants/{id}", merchantHandler.GetMerchant)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	go func() {
		slog.Info("Starting server", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			slog.Error("Server failed", "err", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutdown signal received. Draining connections...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server forced shutdown", "err", err)
	}
	slog.Info("Server stopped gracefully")
	return nil
}
