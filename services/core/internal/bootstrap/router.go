// Package bootstrap is the composition root: it wires adapters to use cases and exposes the HTTP router.
package bootstrap

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	availabilityHTTP "github.com/PabloViniegra/parking-app/services/core/internal/availability/adapters/http"
	availabilityDB "github.com/PabloViniegra/parking-app/services/core/internal/availability/adapters/postgres"
	"github.com/PabloViniegra/parking-app/services/core/internal/availability/application"
	coverageDB "github.com/PabloViniegra/parking-app/services/core/internal/coverage/adapters/postgres"
	"github.com/PabloViniegra/parking-app/services/core/internal/identity/adapters/supabase"
	"github.com/PabloViniegra/parking-app/services/core/internal/shared/clock"
	"github.com/PabloViniegra/parking-app/services/core/internal/shared/ids"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", health)
	client := &http.Client{Timeout: 3 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	auth := supabase.NewAuthenticator(os.Getenv("SUPABASE_URL"), os.Getenv("SUPABASE_PUBLISHABLE_KEY"), client)
	var submitter interface {
		Submit(context.Context, string, string, application.Submission) (application.Announcement, error)
	} = unavailableSubmissions{}
	if databaseURL := os.Getenv("DATABASE_URL"); databaseURL != "" {
		config, err := pgxpool.ParseConfig(databaseURL)
		if err == nil {
			config.MaxConns = 3
			config.MaxConnIdleTime = time.Minute
			config.ConnConfig.ConnectTimeout = 3 * time.Second
			config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec
			pool, err := pgxpool.NewWithConfig(context.Background(), config)
			if err == nil {
				submitter = application.NewSubmitter(availabilityDB.NewSubmissions(pool), coverageDB.NewPolicies(pool), clock.System{}, ids.New)
			}
		}
	}
	mux.Handle("POST /v1/spots", availabilityHTTP.NewAnnouncements(auth, submitter))
	return mux
}

type unavailableSubmissions struct{}

func (unavailableSubmissions) Submit(context.Context, string, string, application.Submission) (application.Announcement, error) {
	return application.Announcement{}, errors.New("database is not configured")
}

func health(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
