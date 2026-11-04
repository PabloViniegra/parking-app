package application_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/PabloViniegra/parking-app/services/core/internal/availability/application"
	"github.com/PabloViniegra/parking-app/services/core/internal/availability/domain"
	coverageDomain "github.com/PabloViniegra/parking-app/services/core/internal/coverage/domain"
)

type coveragePolicies struct{}

func (coveragePolicies) GetPolicy(context.Context, string, float64, float64) (coverageDomain.Policy, error) {
	return coverageDomain.NewPolicy(180)
}

type announcedSpot struct {
	spot  domain.Spot
	event domain.Event
}

type spotRepository struct {
	announced *announcedSpot
}

func (r *spotRepository) Announce(ctx context.Context, spot domain.Spot, event domain.Event) error {
	r.announced = &announcedSpot{spot: spot, event: event}
	return nil
}

func TestAnnouncer_AnnouncesConfirmedSpotWithCoverageTTL(t *testing.T) {
	now := time.Date(2026, time.November, 4, 12, 0, 0, 0, time.UTC)
	repository := &spotRepository{}
	announcer := application.NewAnnouncer(repository, coveragePolicies{})

	spot, err := announcer.Announce(t.Context(), application.AnnounceCommand{
		ID:         "00000000-0000-4000-8000-000000000001",
		EventID:    "00000000-0000-4000-8000-000000000004",
		ReleaserID: "00000000-0000-4000-8000-000000000002",
		CoverageID: "00000000-0000-4000-8000-000000000003",
		CellR10:    622500375513628671,
		Latitude:   40.4168,
		Longitude:  -3.7038,
		Confidence: 0.8,
		Confirmed:  true,
		OccurredAt: now,
	})
	if err != nil {
		t.Fatalf("Announce() error = %v", err)
	}
	if spot.TTLAt() != now.Add(180*time.Second) {
		t.Errorf("TTLAt() = %v, want %v", spot.TTLAt(), now.Add(180*time.Second))
	}
	if repository.announced == nil {
		t.Fatal("repository did not receive the announced Spot and event")
	}
	if repository.announced.event.Type() != "availability.spot.announced.v1" {
		t.Errorf("event type = %q, want availability.spot.announced.v1", repository.announced.event.Type())
	}
}

func TestAnnouncer_DoesNotShareUnconfirmedDeparture(t *testing.T) {
	repository := &spotRepository{}
	announcer := application.NewAnnouncer(repository, coveragePolicies{})

	_, err := announcer.Announce(t.Context(), application.AnnounceCommand{
		ID:         "00000000-0000-4000-8000-000000000001",
		EventID:    "00000000-0000-4000-8000-000000000004",
		ReleaserID: "00000000-0000-4000-8000-000000000002",
		CoverageID: "00000000-0000-4000-8000-000000000003",
		Confirmed:  false,
		OccurredAt: time.Date(2026, time.November, 4, 12, 0, 0, 0, time.UTC),
	})
	if !errors.Is(err, application.ErrNotConfirmed) {
		t.Fatalf("Announce() error = %v, want ErrNotConfirmed", err)
	}
	if repository.announced != nil {
		t.Error("repository received an unconfirmed Spot")
	}
}

func TestAnnouncer_RejectsCellNotMatchingLocation(t *testing.T) {
	repository := &spotRepository{}
	announcer := application.NewAnnouncer(repository, coveragePolicies{})
	_, err := announcer.Announce(t.Context(), application.AnnounceCommand{
		ID: "00000000-0000-4000-8000-000000000001", EventID: "00000000-0000-4000-8000-000000000004",
		ReleaserID: "00000000-0000-4000-8000-000000000002", CoverageID: "00000000-0000-4000-8000-000000000003",
		CellR10: 0, Latitude: 40.4168, Longitude: -3.7038, Confidence: 0.8, Confirmed: true,
		OccurredAt: time.Date(2026, time.November, 4, 12, 0, 0, 0, time.UTC),
	})
	if err == nil {
		t.Fatal("Announce() accepted an invalid H3 cell")
	}
	if repository.announced != nil {
		t.Error("repository received a Spot with an invalid H3 cell")
	}
}
