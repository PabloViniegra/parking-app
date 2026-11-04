# ADR-0013: MapLibre + OpenFreeMap for maps

- **Status:** Accepted
- **Date:** 2026-11-04

## Context
The app shows a zone-probability map. Navigation is delegated to external apps.

## Decision
- **MapLibre React Native** with **OpenFreeMap** vector tiles (no API key, no quota).
- Navigation via deep links to Google Maps, Waze or Apple Maps; we never build turn-by-turn.

## Consequences
- Positive: free, no vendor lock-in, OSM-consistent with our data.
- Negative: public tile service without SLA; self-hosting tiles is the fallback.
