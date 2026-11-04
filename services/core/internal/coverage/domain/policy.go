package domain

import (
	"errors"
	"fmt"
	"math"
	"time"
)

var ErrInvalidPolicy = errors.New("invalid Coverage policy")

type Policy struct {
	spotTTL time.Duration
}

func NewPolicy(spotTTLSeconds int64) (Policy, error) {
	if spotTTLSeconds <= 0 || spotTTLSeconds > math.MaxInt64/int64(time.Second) {
		return Policy{}, fmt.Errorf("%w: Spot TTL must be positive and representable", ErrInvalidPolicy)
	}
	return Policy{spotTTL: time.Duration(spotTTLSeconds) * time.Second}, nil
}

func (p Policy) SpotTTL() time.Duration {
	return p.spotTTL
}
