package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func Health(w http.ResponseWriter, r *http.Request) {
	corrID := r.Header.Get("X-Correlation-ID")
	if corrID == "" {
		corrID = "unknown"
	}
	slog.Info("Health check", "correlation_id", corrID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
