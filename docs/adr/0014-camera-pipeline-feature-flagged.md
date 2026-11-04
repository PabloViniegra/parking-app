# ADR-0014: Camera pipeline built but feature-flagged off in MVP

- **Status:** Accepted
- **Date:** 2026-11-04

## Context
Public cameras (Madrid Informo, DGT NAP) provide still images every ~5 minutes, pointing at roadways. They cannot detect individual spots in time, but can estimate zone occupancy. The product wants a functional base in the MVP.

## Decision
- Build `services/vision`: camera catalog → poll every 5 min → in-memory vehicle-count inference → `zones.occupancy.estimated.v1`.
- Output feeds zone priors only; **never creates Spots**; not visible in UI (`FEATURE_VISION_ENABLED=false`).
- Scheduling: GitHub Actions every 5 min **only in a public repository** (private repo free minutes are insufficient for 288 runs/day). Alternative: a free container/Space host pinged by cron.

## Consequences
- Positive: the pipeline is ready for F2; data accumulates for evaluation.
- Negative: scheduling drift on GitHub Actions; public repo exposes the worker code (no secrets or images in it).
