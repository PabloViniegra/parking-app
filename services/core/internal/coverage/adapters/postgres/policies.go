package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/PabloViniegra/parking-app/services/core/internal/coverage/domain"
	"github.com/PabloViniegra/parking-app/services/core/internal/coverage/ports"
)

type rowQuerier interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func (p *Policies) Locate(ctx context.Context, latitude, longitude float64) (domain.AnnouncementPolicy, error) {
	var policy domain.AnnouncementPolicy
	var ttlSeconds int64
	var base, reliability float64
	err := p.db.QueryRow(ctx, `select id::text, (policies->>'spot_ttl_seconds')::bigint,
	  (policies->>'confirmed_departure_base')::double precision,
	  (policies->>'new_releaser_reliability')::double precision
	  from public.coverage_areas where active
	  and extensions.st_covers(geom,extensions.st_setsrid(extensions.st_makepoint($1,$2),4326)::extensions.geography)
	  order by slug limit 1`, longitude, latitude).Scan(&policy.CoverageID, &ttlSeconds, &base, &reliability)
	if errors.Is(err, pgx.ErrNoRows) {
		return policy, ports.ErrOutsideCoverage
	}
	if err != nil {
		return policy, fmt.Errorf("locate Coverage: %w", err)
	}
	return domain.NewAnnouncementPolicy(policy.CoverageID, ttlSeconds, base, reliability)
}

type Policies struct {
	db rowQuerier
}

func NewPolicies(db rowQuerier) *Policies {
	return &Policies{db: db}
}

func (p *Policies) GetPolicy(ctx context.Context, coverageID string, latitude, longitude float64) (domain.Policy, error) {
	var ttlSeconds int64
	err := p.db.QueryRow(ctx, `
		select (policies->>'spot_ttl_seconds')::bigint
		from public.coverage_areas
		where id = $1 and active
		  and extensions.st_covers(geom, extensions.st_setsrid(extensions.st_makepoint($2, $3), 4326)::extensions.geography)
	`, coverageID, longitude, latitude).Scan(&ttlSeconds)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Policy{}, fmt.Errorf("coverage area %q is not active: %w", coverageID, err)
		}
		return domain.Policy{}, fmt.Errorf("read Coverage policy: %w", err)
	}

	policy, err := domain.NewPolicy(ttlSeconds)
	if err != nil {
		return domain.Policy{}, err
	}
	return policy, nil
}
