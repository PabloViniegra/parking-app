package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/PabloViniegra/parking-app/services/core/internal/availability/domain"
	availabilityPorts "github.com/PabloViniegra/parking-app/services/core/internal/availability/ports"
	"github.com/PabloViniegra/parking-app/services/core/internal/coverage/ports"
	"github.com/PabloViniegra/parking-app/services/core/internal/shared/geo"
)

var ErrNotConfirmed = errors.New("departure was not confirmed")

type Announcer struct {
	repository availabilityPorts.SpotRepository
	policies   ports.PolicyReader
}

type AnnounceCommand struct {
	ID         string
	EventID    string
	ReleaserID string
	CoverageID string
	CellR10    int64
	Latitude   float64
	Longitude  float64
	Confidence float64
	Confirmed  bool
	OccurredAt time.Time
}

func NewAnnouncer(repository availabilityPorts.SpotRepository, policies ports.PolicyReader) *Announcer {
	return &Announcer{repository: repository, policies: policies}
}

func (a *Announcer) Announce(ctx context.Context, command AnnounceCommand) (domain.Spot, error) {
	if !command.Confirmed {
		return domain.Spot{}, ErrNotConfirmed
	}

	policy, err := a.policies.GetPolicy(ctx, command.CoverageID, command.Latitude, command.Longitude)
	if err != nil {
		return domain.Spot{}, fmt.Errorf("load Coverage policy: %w", err)
	}

	spot, err := domain.NewSpot(command.ID, command.ReleaserID, command.CoverageID, command.CellR10, command.Latitude, command.Longitude, command.Confidence, command.OccurredAt, policy.SpotTTL())
	if err != nil {
		return domain.Spot{}, err
	}
	cell, err := geo.CellR10(command.Latitude, command.Longitude)
	if err != nil {
		return domain.Spot{}, err
	}
	if cell != command.CellR10 {
		return domain.Spot{}, fmt.Errorf("%w: H3 cell does not match location", domain.ErrInvalidSpot)
	}

	event := domain.NewSpotAnnouncedEvent(command.EventID, spot, command.OccurredAt)
	if err := a.repository.Announce(ctx, spot, event); err != nil {
		return domain.Spot{}, fmt.Errorf("announce Spot: %w", err)
	}

	return spot, nil
}
