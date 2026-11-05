package ports

import (
	"context"
	"errors"
	"time"

	"github.com/PabloViniegra/parking-app/services/core/internal/availability/domain"
)

var (
	ErrAnnouncementNotFound = errors.New("announcement not found")
	ErrIdempotencyConflict  = errors.New("idempotency key already used for different input")
)

type Receipt struct {
	ID          string
	ExpiresAt   time.Time
	Fingerprint string
}

type SubmissionRepository interface {
	Find(context.Context, string, string) (Receipt, error)
	Save(context.Context, string, string, domain.Spot, domain.Event) (Receipt, error)
}
