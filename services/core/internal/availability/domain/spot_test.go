package domain_test

import (
	"errors"
	"math"
	"testing"
	"time"

	"github.com/PabloViniegra/parking-app/services/core/internal/availability/domain"
)

func TestNewSpotRejectsNonFiniteCoordinatesAndConfidence(t *testing.T) {
	tests := []struct {
		name       string
		latitude   float64
		longitude  float64
		confidence float64
	}{
		{name: "NaN latitude", latitude: math.NaN(), longitude: -3.7, confidence: 0.8},
		{name: "infinite longitude", latitude: 40.4, longitude: math.Inf(1), confidence: 0.8},
		{name: "NaN confidence", latitude: 40.4, longitude: -3.7, confidence: math.NaN()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewSpot("spot", "releaser", "coverage", 1, tt.latitude, tt.longitude, tt.confidence, time.Date(2026, time.November, 4, 12, 0, 0, 0, time.UTC), time.Minute)
			if !errors.Is(err, domain.ErrInvalidSpot) {
				t.Errorf("NewSpot() error = %v, want ErrInvalidSpot", err)
			}
		})
	}
}
