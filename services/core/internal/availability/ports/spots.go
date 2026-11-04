package ports

import (
	"context"

	"github.com/PabloViniegra/parking-app/services/core/internal/availability/domain"
)

type SpotRepository interface {
	Announce(context.Context, domain.Spot, domain.Event) error
}
