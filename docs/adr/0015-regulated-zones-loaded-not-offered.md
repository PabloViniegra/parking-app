# ADR-0015: Regulated zones loaded and labelled, not offered in MVP

- **Status:** Accepted
- **Date:** 2026-11-04

## Context
The MVP covers free street parking only. Much of central Madrid is SER; other municipalities have their own ORA zones without a common dataset.

## Decision
- Load Madrid SER datasets from day one into `regulated_zones`.
- For other municipalities, use OSM tags plus a crowd signal ("this street is regulated").
- Spots inside regulated zones are labelled and **not offered** in MVP.

## Consequences
- Positive: correct offers from day one; ready to enable regulated parking in F2.
- Negative: lower MVP value in central Madrid; focus early adoption outside the M-30 and in other municipalities.
