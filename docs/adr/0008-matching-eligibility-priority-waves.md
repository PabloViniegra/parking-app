# ADR-0008: Matching — ETA eligibility, search-time priority, waves of one

- **Status:** Accepted
- **Date:** 2026-11-04

## Context
Notifying many Seekers about the same Spot causes frustration. The product wants to reward drivers who have been searching longer, without sending them to spots they cannot reach in time.

## Decision
1. **Eligibility:** active session ∧ Spot within session radius ∧ `ETA ≤ remaining TTL`.
2. **Priority:** `score = searchMinutes × w1 − etaMinutes × w2` (initial `w1 = 1.0`, `w2 = 0.3`, configurable per Coverage area).
3. **Waves of exactly one Seeker**, 12 s acceptance window; on timeout, next eligible Seeker.
4. **Anti-gaming:** search time accrues only while the device reports in-vehicle, inside the session zone, below 25 km/h.
5. Seeker ignoring 3 consecutive offers → session paused.
6. Live Spots are not shown on the public map in MVP (proposed default, see DOMAIN OQ-1).

## Consequences
- Positive: fairness, no stampedes to the same spot, rewards time invested.
- Negative: sequential waves reduce the chance someone accepts before TTL; mitigated by short windows and ETA filter. Weights must be tuned with data.
