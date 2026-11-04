package domain_test

import (
	"testing"
	"time"

	"github.com/PabloViniegra/parking-app/services/core/internal/coverage/domain"
)

func TestNewPolicy_UsesConfiguredSpotTTL(t *testing.T) {
	policy, err := domain.NewPolicy(180)
	if err != nil {
		t.Fatal(err)
	}
	if policy.SpotTTL() != 180*time.Second {
		t.Errorf("SpotTTL() = %s, want 3m", policy.SpotTTL())
	}
}
