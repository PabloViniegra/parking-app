//go:build integration

package postgres_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/PabloViniegra/parking-app/services/core/internal/availability/adapters/postgres"
	"github.com/PabloViniegra/parking-app/services/core/internal/availability/domain"
)

func TestRepository_AnnouncementAndEventCommitTogether(t *testing.T) {
	ctx := context.Background()
	pool := newTestPool(t, ctx)
	spotID, eventID, releaserID, coverageID := newTestIDs(t, ctx, pool)
	createTestReleaser(t, ctx, pool, releaserID)
	cleanupTestData(t, ctx, pool, spotID, eventID, releaserID)

	now := time.Date(2026, time.November, 4, 12, 0, 0, 0, time.UTC)
	spot, err := domain.NewSpot(spotID, releaserID, coverageID, 635123456789012345, 40.4168, -3.7038, 0.8, now, 180*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	event := domain.NewSpotAnnouncedEvent(eventID, spot, now)
	if err := postgres.NewRepository(pool).Announce(ctx, spot, event); err != nil {
		t.Fatal(err)
	}

	var count int
	if err := pool.QueryRow(ctx, `select count(*) from public.spots where id = $1`, spotID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("stored Spot count = %d, want 1", count)
	}
	var payload []byte
	if err := pool.QueryRow(ctx, `select payload from public.outbox where id = $1`, eventID).Scan(&payload); err != nil {
		t.Fatal(err)
	}
	var envelope struct {
		ID              string    `json:"id"`
		SpecVersion     string    `json:"specversion"`
		Source          string    `json:"source"`
		Subject         string    `json:"subject"`
		Time            time.Time `json:"time"`
		DataContentType string    `json:"datacontenttype"`
		Type            string    `json:"type"`
		Data            struct {
			SpotID     string  `json:"spot_id"`
			CoverageID string  `json:"coverage_id"`
			CellR10    string  `json:"cell_r10"`
			Confidence float64 `json:"confidence"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Type != "availability.spot.announced.v1" || envelope.Data.CellR10 != "635123456789012345" {
		t.Errorf("persisted event = %+v, want announced event for the Spot cell", envelope)
	}
	if envelope.Data.Confidence != 0.8 {
		t.Errorf("event confidence = %v, want 0.8", envelope.Data.Confidence)
	}
	if envelope.ID != eventID || envelope.Subject != spotID || envelope.SpecVersion != "1.0" || envelope.Source != "/core/availability" || envelope.DataContentType != "application/json" || !envelope.Time.Equal(now) || envelope.Data.SpotID != spotID || envelope.Data.CoverageID != coverageID {
		t.Errorf("incomplete CloudEvent: %+v", envelope)
	}
}

func TestRepository_RollsBackSpotWhenOutboxInsertFails(t *testing.T) {
	ctx := context.Background()
	pool := newTestPool(t, ctx)
	spotID, eventID, releaserID, coverageID := newTestIDs(t, ctx, pool)
	createTestReleaser(t, ctx, pool, releaserID)
	cleanupTestData(t, ctx, pool, spotID, eventID, releaserID)
	if _, err := pool.Exec(ctx, `
		insert into public.outbox (id, type, aggregate_id, payload, occurred_at)
		values ($1, 'test.event.created.v1', $2, '{}', now())
	`, eventID, spotID); err != nil {
		t.Fatal(err)
	}

	now := time.Date(2026, time.November, 4, 12, 0, 0, 0, time.UTC)
	spot, err := domain.NewSpot(spotID, releaserID, coverageID, 635123456789012345, 40.4168, -3.7038, 0.8, now, 180*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	err = postgres.NewRepository(pool).Announce(ctx, spot, domain.NewSpotAnnouncedEvent(eventID, spot, now))
	if err == nil {
		t.Fatal("Announce() succeeded with a duplicate outbox event id")
	}
	var count int
	if err := pool.QueryRow(ctx, `select count(*) from public.spots where id = $1`, spotID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("Spot count after failed announcement = %d, want 0", count)
	}
}

func TestRetention_AggregatesOldSpotsAndRemovesIndividualLocations(t *testing.T) {
	ctx := context.Background()
	pool := newTestPool(t, ctx)
	spotID, eventID, releaserID, coverageID := newTestIDs(t, ctx, pool)
	createTestReleaser(t, ctx, pool, releaserID)
	cleanupTestData(t, ctx, pool, spotID, eventID, releaserID)
	now := time.Date(2026, time.November, 5, 12, 0, 0, 0, time.UTC)
	spot, err := domain.NewSpot(spotID, releaserID, coverageID, 622500375513628671, 40.4168, -3.7038, 0.8, now.Add(-24*time.Hour), 180*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if err := postgres.NewRepository(pool).Announce(ctx, spot, domain.NewSpotAnnouncedEvent(eventID, spot, spot.CreatedAt())); err != nil {
		t.Fatal(err)
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	var recentSpotID, recentEventID, expiredEventID string
	if err := tx.QueryRow(ctx, `select gen_random_uuid()::text, gen_random_uuid()::text, gen_random_uuid()::text`).Scan(&recentSpotID, &recentEventID, &expiredEventID); err != nil {
		t.Fatal(err)
	}
	recent, err := domain.NewSpot(recentSpotID, releaserID, coverageID, spot.CellR10(), 40.4168, -3.7038, 0.8, now.Add(-22*time.Hour), 180*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if err := postgres.NewRepository(tx).Announce(ctx, recent, domain.NewSpotAnnouncedEvent(recentEventID, recent, recent.CreatedAt())); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `insert into public.outbox (id, type, aggregate_id, payload, occurred_at) values ($1, 'availability.spot.announced.v1', $2, '{}', $3)`, expiredEventID, spotID, now.Add(-7*24*time.Hour)); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `select private.purge_spot_locations($1)`, now); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := tx.QueryRow(ctx, `select count(*) from public.spots where id = $1`, spotID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("retention kept an individual Spot older than 24 hours")
	}
	if err := tx.QueryRow(ctx, `select announcements from public.spot_daily_counts where cell_r10 = $1 and day = '2026-11-04'`, spot.CellR10()).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count < 1 {
		t.Fatal("retention lost the anonymous H3 aggregate")
	}
	before := count
	if _, err := tx.Exec(ctx, `select private.purge_spot_locations($1)`, now); err != nil {
		t.Fatal(err)
	}
	if err := tx.QueryRow(ctx, `select announcements from public.spot_daily_counts where cell_r10 = $1 and day = '2026-11-04'`, spot.CellR10()).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != before {
		t.Fatal("repeating retention counted the same Spot twice")
	}
	var hasReleaser bool
	if err := tx.QueryRow(ctx, `select payload->'data' ? 'releaser_id' from public.outbox where id = $1`, eventID).Scan(&hasReleaser); err != nil {
		t.Fatal(err)
	}
	if hasReleaser {
		t.Fatal("retained outbox event identifies the Releaser")
	}
	if err := tx.QueryRow(ctx, `select count(*) from public.spots where id = $1`, recentSpotID).Scan(&count); err != nil || count != 1 {
		t.Fatalf("recent Spot count = %d, error = %v, want 1", count, err)
	}
	if err := tx.QueryRow(ctx, `select count(*) from public.outbox where id = $1`, expiredEventID).Scan(&count); err != nil || count != 0 {
		t.Fatalf("expired outbox count = %d, error = %v, want 0", count, err)
	}
	if err := tx.QueryRow(ctx, `select count(*) from cron.job where jobname = 'purge-spot-locations' and schedule = '0 * * * *' and active`).Scan(&count); err != nil || count != 1 {
		t.Fatalf("hourly retention jobs = %d, error = %v, want 1", count, err)
	}
}

func newTestPool(t *testing.T, ctx context.Context) *pgxpool.Pool {
	t.Helper()
	config, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		t.Fatal(err)
	}
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pool.Close() })
	return pool
}

func newTestIDs(t *testing.T, ctx context.Context, pool *pgxpool.Pool) (string, string, string, string) {
	t.Helper()
	var spotID, eventID, releaserID, coverageID string
	if err := pool.QueryRow(ctx, `select gen_random_uuid()::text, gen_random_uuid()::text, gen_random_uuid()::text`).Scan(&spotID, &eventID, &releaserID); err != nil {
		t.Fatal(err)
	}
	if err := pool.QueryRow(ctx, `select id::text from public.coverage_areas where slug = 'madrid-region'`).Scan(&coverageID); err != nil {
		t.Fatal(err)
	}
	return spotID, eventID, releaserID, coverageID
}

func createTestReleaser(t *testing.T, ctx context.Context, pool *pgxpool.Pool, releaserID string) {
	t.Helper()
	if _, err := pool.Exec(ctx, `
		insert into auth.users (id, aud, role, email, raw_app_meta_data, raw_user_meta_data, is_anonymous, created_at, updated_at)
		values ($1, 'authenticated', 'authenticated', $2, '{}', '{}', true, now(), now())
	`, releaserID, releaserID+"@example.test"); err != nil {
		t.Fatal(err)
	}
}

func cleanupTestData(t *testing.T, ctx context.Context, pool *pgxpool.Pool, spotID, eventID, releaserID string) {
	t.Helper()
	t.Cleanup(func() {
		for _, cleanup := range []struct {
			query string
			id    string
		}{
			{query: `delete from public.spots where id = $1`, id: spotID},
			{query: `delete from public.outbox where id = $1`, id: eventID},
			{query: `delete from auth.users where id = $1`, id: releaserID},
		} {
			if _, err := pool.Exec(ctx, cleanup.query, cleanup.id); err != nil {
				t.Errorf("cleanup integration row: %v", err)
			}
		}
	})
}
