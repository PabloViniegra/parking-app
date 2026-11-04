package clock_test

import (
	"testing"
	"time"

	"github.com/PabloViniegra/parking-app/services/core/internal/shared/clock"
)

func TestFake(t *testing.T) {
	start := time.Date(2026, 11, 4, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		advance []time.Duration
		want    time.Time
	}{
		{name: "returns the start time when not advanced", want: start},
		{
			name:    "moves forward by every advance",
			advance: []time.Duration{12 * time.Second, 168 * time.Second},
			want:    time.Date(2026, 11, 4, 10, 3, 0, 0, time.UTC),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := clock.NewFake(start)
			for _, d := range tt.advance {
				c.Advance(d)
			}

			if got := c.Now(); !got.Equal(tt.want) {
				t.Errorf("Now() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSystemReturnsUTC(t *testing.T) {
	if loc := (clock.System{}).Now().Location(); loc != time.UTC {
		t.Errorf("Now().Location() = %v, want UTC", loc)
	}
}
