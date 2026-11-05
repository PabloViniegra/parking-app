package http

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/PabloViniegra/parking-app/services/core/internal/availability/application"
	availability "github.com/PabloViniegra/parking-app/services/core/internal/availability/ports"
	coverage "github.com/PabloViniegra/parking-app/services/core/internal/coverage/ports"
	"github.com/PabloViniegra/parking-app/services/core/internal/identity/ports"
	"github.com/PabloViniegra/parking-app/services/core/internal/shared/ids"
)

type submitter interface {
	Submit(context.Context, string, string, application.Submission) (application.Announcement, error)
}

func NewAnnouncements(auth ports.Authenticator, service submitter) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 8*time.Second)
		defer cancel()
		header := strings.Fields(r.Header.Get("Authorization"))
		if len(header) != 2 || !strings.EqualFold(header[0], "Bearer") {
			w.Header().Set("WWW-Authenticate", "Bearer")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		driver, err := auth.Verify(ctx, header[1])
		if err != nil {
			if errors.Is(err, ports.ErrInvalidToken) {
				w.Header().Set("WWW-Authenticate", "Bearer")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
			} else {
				http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
			}
			return
		}
		key := strings.ToLower(r.Header.Get("Idempotency-Key"))
		if !ids.Valid(key) {
			http.Error(w, "Invalid idempotency key", http.StatusBadRequest)
			return
		}
		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if err != nil || mediaType != "application/json" {
			http.Error(w, "Expected application/json", http.StatusBadRequest)
			return
		}
		var body struct {
			Latitude  *float64 `json:"latitude"`
			Longitude *float64 `json:"longitude"`
			AccuracyM *float64 `json:"accuracyM"`
		}
		decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&body); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		if decoder.Decode(&struct{}{}) != io.EOF || body.Latitude == nil || body.Longitude == nil || body.AccuracyM == nil || *body.Latitude < -90 || *body.Latitude > 90 || *body.Longitude < -180 || *body.Longitude > 180 || *body.AccuracyM <= 0 {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		result, err := service.Submit(ctx, driver, key, application.Submission{Latitude: *body.Latitude, Longitude: *body.Longitude, AccuracyM: *body.AccuracyM})
		if err != nil {
			switch {
			case errors.Is(err, availability.ErrIdempotencyConflict):
				http.Error(w, "Idempotency key conflict", http.StatusConflict)
			case errors.Is(err, coverage.ErrOutsideCoverage):
				http.Error(w, "Outside active Coverage", http.StatusUnprocessableEntity)
			default:
				http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
			}
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(struct {
			ID        string `json:"id"`
			ExpiresAt string `json:"expiresAt"`
		}{result.ID, result.ExpiresAt.Format("2006-01-02T15:04:05.999999999Z07:00")})
	})
}
