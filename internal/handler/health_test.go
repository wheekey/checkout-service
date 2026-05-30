package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealth(t *testing.T) {
	tests := []struct {
		name       string
		corrID     string
		wantStatus int
	}{
		{"With correlation ID", "req-123", http.StatusOK},
		{"Without correlation ID", "", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			if tt.corrID != "" {
				req.Header.Set("X-Correlation-ID", tt.corrID)
			}
			w := httptest.NewRecorder()

			Health(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("got %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}
