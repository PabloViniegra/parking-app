# ADR-0004: Vercel Functions + Supabase Free as platform

- **Status:** Accepted
- **Date:** 2026-11-04

## Context
The project must run at zero cost during the MVP. GCP is ruled out as a deployment platform.

## Decision
- **Supabase Free**: Postgres + PostGIS, Auth (anonymous sign-ins), `pg_cron`, `pg_net`, optional Realtime.
- **Vercel Hobby**: Go core deployed as Vercel Functions through a single entrypoint, region pinned next to Supabase.
- Connection through the Supabase pooler in **transaction mode**.

## Known constraints
- Supabase Free: 500 MB database, 5 GB egress, 50k MAU, **pause after 7 days of inactivity**.
- Vercel Hobby: cron limited to once per day, limited active CPU, **non-commercial use only**.

## Consequences
- Positive: zero cost; managed auth and database; quick start.
- Negative: no long-running processes (solved by ADR-0005); cold starts may affect latency SLO (validated in Spike 1); pause risk during development (keep-alive workflow).
- **Exit strategy:** the Go core is a plain HTTP server behind adapters; it can be containerized and moved to another host without touching domain code.
- **Upgrade triggers:** DB > 400 MB, real traffic where a pause is unacceptable, commercial use, latency SLO violations.

## Alternatives considered
- GCP (Cloud Run, Pub/Sub, Memorystore) — ruled out by requirement.
- Container hosts with free tiers — kept as exit strategy.
