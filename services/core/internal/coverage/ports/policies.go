package ports

import (
	"context"

	"github.com/PabloViniegra/parking-app/services/core/internal/coverage/domain"
)

type PolicyReader interface {
	GetPolicy(context.Context, string, float64, float64) (domain.Policy, error)
}
