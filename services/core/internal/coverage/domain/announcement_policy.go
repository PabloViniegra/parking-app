package domain

import (
	"math"
	"time"
)

type AnnouncementPolicy struct {
	CoverageID string
	TTL        time.Duration
	Confidence float64
}

func NewAnnouncementPolicy(coverageID string, ttlSeconds int64, base, reliability float64) (AnnouncementPolicy, error) {
	policy, err := NewPolicy(ttlSeconds)
	if err != nil {
		return AnnouncementPolicy{}, err
	}
	if coverageID == "" || math.IsNaN(base) || math.IsNaN(reliability) || base < 0 || base > 1 || reliability < 0 || reliability > 1 {
		return AnnouncementPolicy{}, ErrInvalidPolicy
	}
	return AnnouncementPolicy{CoverageID: coverageID, TTL: policy.SpotTTL(), Confidence: base * reliability}, nil
}
