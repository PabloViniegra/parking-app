# ADR-0007: Transactional outbox as event bus

- **Status:** Accepted
- **Date:** 2026-11-04

## Context
The architecture is event-driven, but there is no managed message broker on the chosen free platforms. Events must never be lost or published for rolled-back changes.

## Decision
- Every state change inserts its events into `outbox` in the **same transaction**.
- After commit, events are dispatched **in process** to subscribing modules.
- External consumers (analytics, workers) read the outbox by cursor.
- Events use the CloudEvents envelope and versioned names `<context>.<entity>.<past-tense>.v<n>`; schemas in `contracts/events/`.
- Retention: 7 days, then aggregated.

## Consequences
- Positive: exactly-once production semantics relative to state; no broker; replayable.
- Negative: subscribers must be idempotent; a broker adapter can be added later behind the `EventPublisher` port.
