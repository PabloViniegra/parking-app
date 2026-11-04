# ADR-0006: No Redis in MVP; H3 indexing in Postgres

- **Status:** Accepted
- **Date:** 2026-11-04

## Context
Hot state (active Spots, sessions) is small at MVP scale. An extra store adds cost, vendor and consistency issues (Redis key-expiry notifications are not reliable business events anyway).

## Decision
- Store hot state in Postgres with an H3 cell id (`bigint`, B-tree index) computed in Go (`uber/h3-go`): resolution 10 for Spots/sessions, 8 for zone priors.
- Neighbour queries: k-ring computed in code + `WHERE cell = ANY($1)`.
- PostGIS for segment and zone geometries.
- Keep a `SpotStore` / `SessionStore` port so a Redis adapter can be added later.

## Consequences
- Positive: one store, transactional, free.
- Negative: higher read latency than Redis under heavy load. Revisit when active sessions per cell grow or p95 degrades.
