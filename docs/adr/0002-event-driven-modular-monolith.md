# ADR-0002: Event-driven modular monolith, hexagonal + screaming architecture

- **Status:** Accepted
- **Date:** 2026-11-04

## Context
The domain has clear bounded contexts (Availability, Matching, Community, Zones, Coverage, Identity) that react to each other's events. The team is small, the MVP must be cheap, and the hot path needs low latency.

## Decision
- Build the core as **one deployable modular monolith** with one module per bounded context.
- Modules communicate **only via domain events** (in-process after commit, backed by the outbox — ADR-0007) or published port interfaces.
- Each module uses **hexagonal architecture** (`domain`, `application`, `ports`, `adapters`).
- Repository layout follows **screaming architecture**: top-level folders are business capabilities.
- Only ingestion and vision run as separate services (different runtime and scaling — ADR-0003).

## Consequences
- Positive: one deploy, no network hops on the hot path, easy refactoring of boundaries, can be split later along module lines.
- Negative: discipline required to keep boundaries; enforced with architecture tests in CI.

## Alternatives considered
- Microservices from day one — operational cost and latency without benefit at MVP scale.
- Layered monolith — blurs bounded contexts, harder to split later.
