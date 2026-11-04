# ADR-0001: Record architecture decisions

- **Status:** Accepted
- **Date:** 2026-11-04

## Context
The project will evolve over phases and may be worked on by several people and coding agents. Decisions and their rationale must survive beyond conversations.

## Decision
Record every significant architectural decision as an ADR in `docs/adr/`, numbered sequentially, using `template.md`. Accepted ADRs are immutable; changes are made by superseding them.

## Consequences
- Positive: traceable rationale; agents can check decisions before changing code.
- Negative: small overhead per decision.
