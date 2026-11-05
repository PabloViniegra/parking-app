package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/jackc/pgx/v5"

	"github.com/PabloViniegra/parking-app/services/core/internal/availability/domain"
	"github.com/PabloViniegra/parking-app/services/core/internal/shared/events"
)

type beginner interface {
	Begin(context.Context) (pgx.Tx, error)
}

type Repository struct {
	db beginner
}

type announcedSpotData struct {
	SpotID     string  `json:"spot_id"`
	CoverageID string  `json:"coverage_id"`
	CellR10    string  `json:"cell_r10"`
	Confidence float64 `json:"confidence"`
}

func NewRepository(db beginner) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Announce(ctx context.Context, spot domain.Spot, event domain.Event) (err error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin Spot announcement: %w", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			err = errors.Join(err, fmt.Errorf("rollback Spot announcement: %w", rollbackErr))
		}
	}()

	_, err = tx.Exec(ctx, `
		insert into public.spots (
			id, cell_r10, location, status, confidence, ttl_at, releaser_id, coverage_id, created_at
		) values (
			$1, $2, extensions.st_setsrid(extensions.st_makepoint($3, $4), 4326)::extensions.geography,
			'announced', $5, $6, $7, $8, $9
		)
	`, spot.ID(), spot.CellR10(), spot.Longitude(), spot.Latitude(), spot.Confidence(), spot.TTLAt(), spot.ReleaserID(), spot.CoverageID(), spot.CreatedAt())
	if err != nil {
		return fmt.Errorf("insert Spot: %w", err)
	}

	payload, err := json.Marshal(events.Envelope[announcedSpotData]{
		SpecVersion:     "1.0",
		ID:              event.ID(),
		Source:          "/core/availability",
		Type:            event.Type(),
		Time:            event.OccurredAt(),
		Subject:         event.SpotID(),
		DataContentType: "application/json",
		Data: announcedSpotData{
			SpotID:     event.SpotID(),
			CoverageID: event.CoverageID(),
			CellR10:    strconv.FormatInt(event.CellR10(), 10),
			Confidence: event.Confidence(),
		},
	})
	if err != nil {
		return fmt.Errorf("encode Spot event: %w", err)
	}

	_, err = tx.Exec(ctx, `
		insert into public.outbox (id, type, aggregate_id, payload, occurred_at)
		values ($1, $2, $3, $4, $5)
	`, event.ID(), event.Type(), event.SpotID(), string(payload), event.OccurredAt())
	if err != nil {
		return fmt.Errorf("insert Spot event: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit Spot announcement: %w", err)
	}
	return nil
}
