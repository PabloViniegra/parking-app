# ADR-0010: Privacy by design — minimization and 24 h location retention

- **Status:** Accepted
- **Date:** 2026-11-04

## Context
Location traces are personal data under GDPR. Camera images may contain plates and people. The free database is limited to 500 MB.

## Decision
- No plate, car model, name or phone. Anonymous sign-in by default.
- Precise coordinates retained **≤ 24 h**; an hourly `pg_cron` job purges them and keeps only H3 aggregates without driver id.
- Location is collected only during an active session or a departure/parking event.
- Camera images are processed in memory and discarded; only per-cell counts are stored.
- Account and data deletion available in-app.
- Legal/GDPR review before public launch.

## Consequences
- Positive: compliance posture, user trust, fits the storage limit.
- Negative: limited raw history for analytics; aggregates must be designed up front.
