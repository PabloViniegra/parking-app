# ADR-0009: Releaser explicit confirmation, opt-in auto-share

- **Status:** Accepted
- **Date:** 2026-11-04

## Context
Departure can be detected automatically (Bluetooth, activity recognition). Sharing by default would maximize volume, but the product prioritizes user control.

## Decision
- Detected departure shows a Google-Maps-style prompt "Are you leaving the spot?" with an 8 s countdown, while the car is still stationary.
- **No answer = not shared.** A Spot is created only on explicit confirmation or the manual "I'm leaving" button.
- Opt-in setting "Always share automatically" (default OFF).
- Extra reputation points incentivize sharing.

## Consequences
- Positive: user trust and control; fewer false spots.
- Negative: fewer announcements than auto-share; monitor departure prompt acceptance rate (target ≥ 30 %).
