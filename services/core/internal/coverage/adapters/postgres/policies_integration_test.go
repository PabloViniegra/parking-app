//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/PabloViniegra/parking-app/services/core/internal/coverage/adapters/postgres"
)

func TestPolicies_MadridCoverageMatchesKnownLocations(t *testing.T) {
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	for _, tt := range []struct {
		name     string
		lat, lng float64
		inMadrid bool
	}{
		{name: "Madrid", lat: 40.4168, lng: -3.7038, inMadrid: true},
		{name: "Alcalá de Henares", lat: 40.4821, lng: -3.3635, inMadrid: true},
		{name: "Toledo", lat: 39.8628, lng: -4.0273},
		{name: "Guadalajara", lat: 40.6325, lng: -3.1669},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := postgres.NewPolicies(pool).GetPolicy(context.Background(), "00000000-0000-4000-8000-000000000001", tt.lat, tt.lng)
			if tt.inMadrid && err != nil {
				t.Errorf("GetPolicy() rejected a location in Madrid: %v", err)
			}
			if !tt.inMadrid && !errors.Is(err, pgx.ErrNoRows) {
				t.Errorf("GetPolicy() error = %v, want no matching Coverage", err)
			}
		})
	}
}

func TestPolicies_GetPolicyReadsActiveMadridPolicy(t *testing.T) {
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	tx, err := pool.Begin(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	if _, err := tx.Exec(context.Background(), `update public.coverage_areas set active = true, geom = extensions.st_multi(extensions.st_geomfromtext('POLYGON((-4 40,-3 40,-3 41,-4 41,-4 40))', 4326))::extensions.geography where slug = 'madrid-region'`); err != nil {
		t.Fatal(err)
	}

	var coverageID string
	if err := tx.QueryRow(context.Background(), `select id from public.coverage_areas where slug = $1`, "madrid-region").Scan(&coverageID); err != nil {
		t.Fatal(err)
	}

	policy, err := postgres.NewPolicies(tx).GetPolicy(context.Background(), coverageID, 40.4168, -3.7038)
	if err != nil {
		t.Fatal(err)
	}
	if policy.SpotTTL() != 180*time.Second {
		t.Errorf("SpotTTL() = %s, want 3m", policy.SpotTTL())
	}
}

func TestPolicies_RejectsLocationOutsideCoverage(t *testing.T) {
	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	tx, err := pool.Begin(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	if _, err := tx.Exec(context.Background(), `update public.coverage_areas set active = true, geom = extensions.st_multi(extensions.st_geomfromtext('POLYGON((-4 40,-3 40,-3 41,-4 41,-4 40))', 4326))::extensions.geography where slug = 'madrid-region'`); err != nil {
		t.Fatal(err)
	}
	var coverageID string
	if err := tx.QueryRow(context.Background(), `select id from public.coverage_areas where slug = 'madrid-region'`).Scan(&coverageID); err != nil {
		t.Fatal(err)
	}
	if _, err := postgres.NewPolicies(tx).GetPolicy(context.Background(), coverageID, 48.8566, 2.3522); err == nil {
		t.Fatal("GetPolicy() accepted a location outside the Coverage area")
	}
}
