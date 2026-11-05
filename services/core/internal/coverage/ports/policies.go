package ports

import (
	"context"
	"errors"

	"github.com/PabloViniegra/parking-app/services/core/internal/coverage/domain"
)

var ErrOutsideCoverage = errors.New("location outside active Coverage")

type PolicyLocator interface {
	Locate(context.Context, float64, float64) (domain.AnnouncementPolicy, error)
}

type PolicyReader interface {
	GetPolicy(context.Context, string, float64, float64) (domain.Policy, error)
}
