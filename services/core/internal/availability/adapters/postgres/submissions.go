package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/PabloViniegra/parking-app/services/core/internal/availability/domain"
	"github.com/PabloViniegra/parking-app/services/core/internal/availability/ports"
)

type submissionDB interface {
	Begin(context.Context) (pgx.Tx, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type Submissions struct{ db submissionDB }

func NewSubmissions(db submissionDB) *Submissions { return &Submissions{db: db} }

func (s *Submissions) Find(ctx context.Context, driver, key string) (ports.Receipt, error) {
	var receipt ports.Receipt
	err := s.db.QueryRow(ctx, `select s.id::text,s.ttl_at,a.fingerprint from public.spot_announcements a join public.spots s on s.id=a.spot_id where a.releaser_id=$1 and a.idempotency_key=$2`, driver, key).Scan(&receipt.ID, &receipt.ExpiresAt, &receipt.Fingerprint)
	if errors.Is(err, pgx.ErrNoRows) {
		return receipt, ports.ErrAnnouncementNotFound
	}
	if err != nil {
		return receipt, fmt.Errorf("find announcement: %w", err)
	}
	return receipt, nil
}

func (s *Submissions) Save(ctx context.Context, key, fingerprint string, spot domain.Spot, event domain.Event) (ports.Receipt, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return ports.Receipt{}, fmt.Errorf("begin submission: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	inserted, err := tx.Exec(ctx, `insert into public.spot_announcements (releaser_id,idempotency_key,fingerprint,spot_id) values ($1,$2,$3,$4) on conflict (releaser_id,idempotency_key) do nothing`, spot.ReleaserID(), key, fingerprint, spot.ID())
	if err != nil {
		return ports.Receipt{}, fmt.Errorf("insert announcement receipt: %w", err)
	}
	if inserted.RowsAffected() == 0 {
		receipt, err := NewSubmissions(tx).Find(ctx, spot.ReleaserID(), key)
		if err != nil {
			return receipt, err
		}
		if receipt.Fingerprint != fingerprint {
			return ports.Receipt{}, ports.ErrIdempotencyConflict
		}
		return receipt, nil
	}
	if err := NewRepository(tx).Announce(ctx, spot, event); err != nil {
		return ports.Receipt{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ports.Receipt{}, fmt.Errorf("commit submission: %w", err)
	}
	return ports.Receipt{ID: spot.ID(), ExpiresAt: spot.TTLAt(), Fingerprint: fingerprint}, nil
}
