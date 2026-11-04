# ADR-0003: Go for the core, Python for ingestion and vision

- **Status:** Accepted
- **Date:** 2026-11-04

## Context
The core needs fast startup (serverless), concurrency and low latency. Ingestion and computer vision need the geospatial and ML ecosystems.

## Decision
- **Go** for `services/core`.
- **Python** for `services/ingestion` (OSM, SER, NAP, traffic) and `services/vision` (camera inference).
- Workers write only to Zones-owned tables or publish Zones events; they never touch Availability or Matching tables.

## Consequences
- Positive: right tool per workload; fits the team's existing skills.
- Negative: two toolchains in CI; shared contracts must be language-neutral (JSON Schema).
