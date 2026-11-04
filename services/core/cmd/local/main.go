package main

import (
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/PabloViniegra/parking-app/services/core/internal/bootstrap"
)

func main() {
	addr := ":" + envOr("PORT", "8080")
	srv := &http.Server{
		Addr:              addr,
		Handler:           bootstrap.NewRouter(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	slog.Info("core listening", "addr", addr)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("core stopped", "error", err)
		os.Exit(1)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
