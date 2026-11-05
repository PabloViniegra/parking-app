//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/PabloViniegra/parking-app/services/core/internal/availability/adapters/postgres"
	"github.com/PabloViniegra/parking-app/services/core/internal/availability/domain"
	"github.com/PabloViniegra/parking-app/services/core/internal/availability/ports"
)

func TestSubmissions_StoresAnnouncementAndReceipt(t *testing.T) {
	ctx := context.Background()
	pool := newTestPool(t, ctx)
	spotID, eventID, driver, coverage := newTestIDs(t, ctx, pool)
	createTestReleaser(t, ctx, pool, driver)
	cleanupTestData(t, ctx, pool, spotID, eventID, driver)
	now := time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC)
	spot, err := domain.NewSpot(spotID, driver, coverage, 622500375513628671, 40.4168, -3.7038, 0.64, now, 180*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	repository := postgres.NewSubmissions(pool)
	receipt, err := repository.Save(ctx, eventID, "request-fingerprint", spot, domain.NewSpotAnnouncedEvent(eventID, spot, now))
	if err != nil {
		t.Fatal(err)
	}
	found, err := repository.Find(ctx, driver, eventID)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.ID != found.ID || receipt.Fingerprint != found.Fingerprint || !receipt.ExpiresAt.Equal(found.ExpiresAt) || receipt.ID != spotID || receipt.Fingerprint != "request-fingerprint" || !receipt.ExpiresAt.Equal(time.Date(2026, 11, 5, 12, 3, 0, 0, time.UTC)) {
		t.Fatalf("receipt=%+v found=%+v", receipt, found)
	}
}

func TestSubmissions_RejectsChangedInputAndScopesKeysToReleaser(t *testing.T) {
	ctx := context.Background()
	pool := newTestPool(t, ctx)
	spotID, eventID, driver, coverage := newTestIDs(t, ctx, pool)
	createTestReleaser(t, ctx, pool, driver)
	cleanupTestData(t, ctx, pool, spotID, eventID, driver)
	now := time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC)
	spot, err := domain.NewSpot(spotID, driver, coverage, 622500375513628671, 40.4168, -3.7038, 0.64, now, 180*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	repository := postgres.NewSubmissions(pool)
	if _, err := repository.Save(ctx, eventID, "original", spot, domain.NewSpotAnnouncedEvent(eventID, spot, now)); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.Save(ctx, eventID, "different", spot, domain.NewSpotAnnouncedEvent(eventID, spot, now)); !errors.Is(err, ports.ErrIdempotencyConflict) {
		t.Fatalf("error=%v, want conflict", err)
	}
	_, _, otherDriver, _ := newTestIDs(t, ctx, pool)
	if _, err := repository.Find(ctx, otherDriver, eventID); !errors.Is(err, ports.ErrAnnouncementNotFound) {
		t.Fatalf("other Releaser lookup=%v", err)
	}
}

func TestSubmissions_ConcurrentRetriesCreateOneAnnouncement(t *testing.T) {
	ctx := context.Background()
	pool := newTestPool(t, ctx)
	spotID, eventID, driver, coverage := newTestIDs(t, ctx, pool)
	createTestReleaser(t, ctx, pool, driver)
	cleanupTestData(t, ctx, pool, spotID, eventID, driver)
	now := time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC)
	spot, err := domain.NewSpot(spotID, driver, coverage, 622500375513628671, 40.4168, -3.7038, 0.64, now, 180*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	repository := postgres.NewSubmissions(pool)
	var workers sync.WaitGroup
	for range 4 {
		workers.Go(func() {
			receipt, err := repository.Save(ctx, eventID, "same", spot, domain.NewSpotAnnouncedEvent(eventID, spot, now))
			if err != nil || receipt.ID != spotID {
				t.Errorf("receipt=%+v err=%v", receipt, err)
			}
		})
	}
	workers.Wait()
	var events int
	if err := pool.QueryRow(ctx, `select count(*) from public.outbox where aggregate_id=$1`, spotID).Scan(&events); err != nil {
		t.Fatal(err)
	}
	if events != 1 {
		t.Fatalf("events=%d, want 1", events)
	}
}

func TestSubmissions_RollsBackReceiptWhenEventCannotBeStored(t *testing.T) {
	ctx := context.Background()
	pool := newTestPool(t, ctx)
	spotID, eventID, driver, coverage := newTestIDs(t, ctx, pool)
	createTestReleaser(t, ctx, pool, driver)
	cleanupTestData(t, ctx, pool, spotID, eventID, driver)
	if _, err := pool.Exec(ctx, `insert into public.outbox(id,type,aggregate_id,payload,occurred_at) values($1,'test.event.created.v1',$2,'{}',now())`, eventID, spotID); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC)
	spot, err := domain.NewSpot(spotID, driver, coverage, 622500375513628671, 40.4168, -3.7038, 0.64, now, 180*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	repository := postgres.NewSubmissions(pool)
	if _, err := repository.Save(ctx, eventID, "request", spot, domain.NewSpotAnnouncedEvent(eventID, spot, now)); err == nil {
		t.Fatal("Save succeeded despite duplicate event ID")
	}
	if _, err := repository.Find(ctx, driver, eventID); !errors.Is(err, ports.ErrAnnouncementNotFound) {
		t.Fatalf("receipt survived failed transaction: %v", err)
	}
	var spots int
	if err := pool.QueryRow(ctx, `select count(*) from public.spots where id=$1`, spotID).Scan(&spots); err != nil {
		t.Fatal(err)
	}
	if spots != 0 {
		t.Fatalf("spots=%d, want 0", spots)
	}
}

func TestSubmissions_PrivacyRetentionRemovesTheIdempotencyReceipt(t *testing.T) {
	ctx := context.Background()
	pool := newTestPool(t, ctx)
	spotID, eventID, driver, coverage := newTestIDs(t, ctx, pool)
	createTestReleaser(t, ctx, pool, driver)
	cleanupTestData(t, ctx, pool, spotID, eventID, driver)
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	now := time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC)
	spot, err := domain.NewSpot(spotID, driver, coverage, 622500375513628671, 40.4168, -3.7038, 0.64, now.Add(-24*time.Hour), 180*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	repository := postgres.NewSubmissions(tx)
	if _, err := repository.Save(ctx, eventID, "request", spot, domain.NewSpotAnnouncedEvent(eventID, spot, spot.CreatedAt())); err != nil {
		t.Fatal(err)
	}
	if _, err := tx.Exec(ctx, `select private.purge_spot_locations($1)`, now); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.Find(ctx, driver, eventID); !errors.Is(err, ports.ErrAnnouncementNotFound) {
		t.Fatalf("receipt survived privacy retention: %v", err)
	}
}

func TestSubmissions_RetryReturnsOriginalWithoutAnotherEvent(t *testing.T) {
	ctx := context.Background()
	pool := newTestPool(t, ctx)
	spotID, eventID, driver, coverage := newTestIDs(t, ctx, pool)
	createTestReleaser(t, ctx, pool, driver)
	cleanupTestData(t, ctx, pool, spotID, eventID, driver)
	now := time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC)
	spot, err := domain.NewSpot(spotID, driver, coverage, 622500375513628671, 40.4168, -3.7038, 0.64, now, 180*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	repository := postgres.NewSubmissions(pool)
	first, err := repository.Save(ctx, eventID, "same-request", spot, domain.NewSpotAnnouncedEvent(eventID, spot, now))
	if err != nil {
		t.Fatal(err)
	}
	second, err := repository.Save(ctx, eventID, "same-request", spot, domain.NewSpotAnnouncedEvent(eventID, spot, now))
	if err != nil {
		t.Fatal(err)
	}
	if first.ID != second.ID || !first.ExpiresAt.Equal(second.ExpiresAt) {
		t.Fatalf("retry changed result: %+v %+v", first, second)
	}
	var events int
	if err := pool.QueryRow(ctx, `select count(*) from public.outbox where aggregate_id=$1`, spotID).Scan(&events); err != nil {
		t.Fatal(err)
	}
	if events != 1 {
		t.Fatalf("events=%d, want 1", events)
	}
}
