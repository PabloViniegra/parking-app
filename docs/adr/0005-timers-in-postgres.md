# ADR-0005: Timers in Postgres with pg_cron + pg_net

- **Status:** Accepted
- **Date:** 2026-11-04

## Context
The domain needs short timers: wave windows (12 s), claim expiry, Spot TTL (180 s), session timeouts. Serverless functions cannot hold them; Vercel Hobby cron runs once per day; free external queues (e.g., QStash Free ~1,000 msgs/day) are too small for the whole Comunidad de Madrid.

## Decision
- A `timers` table (`kind`, `aggregate_id`, `due_at`, `payload`, `fired_at`, `attempts`).
- `pg_cron` every **5 s** runs a SQL function that calls `pg_net.http_post` to `/internal/timers/fire` **only if due timers exist**.
- The core claims due timers with `FOR UPDATE SKIP LOCKED`, runs the use case, marks them fired. Handlers are idempotent.
- Endpoint authenticated with a shared secret (`INTERNAL_TIMERS_SECRET`).

## Consequences
- Positive: free, transactional with domain state, no extra vendor, no invocations when idle.
- Negative: up to ~5 s drift (acceptable); depends on `pg_cron` sub-minute schedules and `pg_net` (validated in Spike 2).
- Fallback: an external free cron hitting the same endpoint if Spike 2 fails.
