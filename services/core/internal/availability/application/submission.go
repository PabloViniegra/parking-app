package application

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/PabloViniegra/parking-app/services/core/internal/availability/domain"
	"github.com/PabloViniegra/parking-app/services/core/internal/availability/ports"
	coverage "github.com/PabloViniegra/parking-app/services/core/internal/coverage/ports"
	"github.com/PabloViniegra/parking-app/services/core/internal/shared/clock"
	"github.com/PabloViniegra/parking-app/services/core/internal/shared/geo"
)

type Submission struct {
	Latitude  float64
	Longitude float64
	AccuracyM float64
}

type Announcement struct {
	ID        string
	ExpiresAt time.Time
}

type Submitter struct {
	repository ports.SubmissionRepository
	policies   coverage.PolicyLocator
	clock      clock.Clock
	newID      func() string
}

func NewSubmitter(repository ports.SubmissionRepository, policies coverage.PolicyLocator, clock clock.Clock, newID func() string) *Submitter {
	return &Submitter{repository: repository, policies: policies, clock: clock, newID: newID}
}

func (s *Submitter) Submit(ctx context.Context, driver, key string, input Submission) (Announcement, error) {
	encoded, err := json.Marshal(input)
	if err != nil {
		return Announcement{}, fmt.Errorf("encode submission: %w", err)
	}
	fingerprint := fmt.Sprintf("%x", sha256.Sum256(encoded))
	previous, err := s.repository.Find(ctx, driver, key)
	if err == nil {
		if previous.Fingerprint != fingerprint {
			return Announcement{}, ports.ErrIdempotencyConflict
		}
		return Announcement{ID: previous.ID, ExpiresAt: previous.ExpiresAt}, nil
	}
	if !errors.Is(err, ports.ErrAnnouncementNotFound) {
		return Announcement{}, err
	}
	policy, err := s.policies.Locate(ctx, input.Latitude, input.Longitude)
	if err != nil {
		return Announcement{}, err
	}
	cell, err := geo.CellR10(input.Latitude, input.Longitude)
	if err != nil {
		return Announcement{}, err
	}
	now := s.clock.Now().UTC().Truncate(time.Microsecond)
	spot, err := domain.NewSpot(s.newID(), driver, policy.CoverageID, cell, input.Latitude, input.Longitude, policy.Confidence, now, policy.TTL)
	if err != nil {
		return Announcement{}, err
	}
	receipt, err := s.repository.Save(ctx, key, fingerprint, spot, domain.NewSpotAnnouncedEvent(s.newID(), spot, now))
	if err != nil {
		return Announcement{}, err
	}
	return Announcement{ID: receipt.ID, ExpiresAt: receipt.ExpiresAt}, nil
}
