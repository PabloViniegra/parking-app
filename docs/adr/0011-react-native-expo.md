# ADR-0011: React Native + Expo (dev client) for mobile

- **Status:** Accepted
- **Date:** 2026-11-04

## Context
One codebase for Android and iOS, aligned with the team's TypeScript experience. Background departure detection needs native APIs.

## Decision
- React Native with **Expo**, using a **development client** (not Expo Go) to include custom native modules.
- Native modules (Expo Modules API) for activity recognition, Bluetooth connection events and geofences.
- Builds via EAS (free tier) or local builds.

## Consequences
- Positive: fast iteration, OTA updates, single codebase.
- Negative: native code still required for detection; background behaviour differs per OS (validated in Spike 3).
