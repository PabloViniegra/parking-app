package domain

import (
	"errors"
	"fmt"
	"math"
	"time"
)

var ErrInvalidSpot = errors.New("invalid spot")

type Spot struct {
	id         string
	releaserID string
	coverageID string
	cellR10    int64
	latitude   float64
	longitude  float64
	confidence float64
	createdAt  time.Time
	ttlAt      time.Time
}

func NewSpot(id, releaserID, coverageID string, cellR10 int64, latitude, longitude, confidence float64, createdAt time.Time, ttl time.Duration) (Spot, error) {
	if id == "" || releaserID == "" || coverageID == "" || createdAt.IsZero() || ttl <= 0 || !finite(latitude) || latitude < -90 || latitude > 90 || !finite(longitude) || longitude < -180 || longitude > 180 || !finite(confidence) || confidence < 0 || confidence > 1 {
		return Spot{}, fmt.Errorf("%w: required value out of range", ErrInvalidSpot)
	}

	return Spot{
		id:         id,
		releaserID: releaserID,
		coverageID: coverageID,
		cellR10:    cellR10,
		latitude:   latitude,
		longitude:  longitude,
		confidence: confidence,
		createdAt:  createdAt,
		ttlAt:      createdAt.Add(ttl),
	}, nil
}

func finite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func (s Spot) ID() string           { return s.id }
func (s Spot) ReleaserID() string   { return s.releaserID }
func (s Spot) CoverageID() string   { return s.coverageID }
func (s Spot) CellR10() int64       { return s.cellR10 }
func (s Spot) Latitude() float64    { return s.latitude }
func (s Spot) Longitude() float64   { return s.longitude }
func (s Spot) Confidence() float64  { return s.confidence }
func (s Spot) CreatedAt() time.Time { return s.createdAt }
func (s Spot) TTLAt() time.Time     { return s.ttlAt }
