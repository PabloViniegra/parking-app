# parking-app

[![CI](https://github.com/PabloViniegra/parking-app/actions/workflows/ci.yml/badge.svg)](https://github.com/PabloViniegra/parking-app/actions/workflows/ci.yml)
![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)
![Expo](https://img.shields.io/badge/Expo-SDK%2058-000020?logo=expo&logoColor=white)
![React Native](https://img.shields.io/badge/React%20Native-0.88-61DAFB?logo=react&logoColor=black)
![TypeScript](https://img.shields.io/badge/TypeScript-6.0-3178C6?logo=typescript&logoColor=white)
![Python](https://img.shields.io/badge/Python-3.13-3776AB?logo=python&logoColor=white)
![Supabase](https://img.shields.io/badge/Supabase-Postgres%20%2B%20PostGIS-3FCF8E?logo=supabase&logoColor=white)

> Working name. Collaborative mobile app that helps drivers find **free street parking** by letting drivers who are leaving a spot notify drivers who are searching nearby, enriched with open public data (OpenStreetMap, Madrid SER, DGT NAP, public traffic cameras).

**MVP scope:** Comunidad de Madrid · free street parking only · designed to expand to all of Spain by configuration.

**Status:** Phase 0 (foundations). The monorepo, toolchains, CI and architecture tests are in place; product features (FR-01 … FR-14) are not implemented yet.

## Contents

- [Documentation map](#documentation-map)
- [Tech stack](#tech-stack)
- [Getting started](#getting-started)
- [Repository layout](#repository-layout)
- [Initialization checklist (Phase 0)](#initialization-checklist-phase-0)
- [Environment variables (core)](#environment-variables-core)
- [Contributing](#contributing)
- [Language policy](#language-policy)

---

## Documentation map

| Document | What it covers |
|---|---|
| [`docs/domain/DOMAIN.md`](docs/domain/DOMAIN.md) | Core domain, subdomains, ubiquitous language, Spot lifecycle, business rules, domain events, bounded contexts |
| [`docs/architecture/ARCHITECTURE.md`](docs/architecture/ARCHITECTURE.md) | Quality attributes, system architecture, deployment on free tiers, timers, outbox, code structure, privacy by design |
| [`docs/product/MVP.md`](docs/product/MVP.md) | Functional requirements, notification UX, out of scope, success metrics, phases |
| [`docs/data-sources/DATA_SOURCES.md`](docs/data-sources/DATA_SOURCES.md) | Catalog of open data sources and how each one is used |
| [`docs/adr/`](docs/adr/README.md) | Architecture Decision Records |

---

## Tech stack

| Layer | Choice |
|---|---|
| Mobile | React Native + Expo (dev client, custom native modules) |
| Core backend | Go — modular monolith, hexagonal, deployed as Vercel Functions |
| Database / Auth / Cron | Supabase **Free** (Postgres + PostGIS, Auth anonymous sign-ins, `pg_cron`, `pg_net`) |
| Push notifications | Expo Push Service |
| Maps | MapLibre + OpenFreeMap tiles |
| Ingestion & vision workers | Python, scheduled via GitHub Actions |
| Geo indexing | Uber H3 (computed in application code, stored as `bigint`) |

---

## Getting started

Prerequisites: Go 1.26, Node 24 with pnpm 11, [uv](https://docs.astral.sh/uv/), Docker (only for local Supabase).

```sh
make setup        # pnpm install, uv sync, go mod download
make check        # all linters, type-checkers and tests
make core-run     # core API on http://localhost:8080 (GET /healthz)
make core-integration # PostgreSQL adapter tests (run make db-start first)
make mobile-start # Expo dev server (requires a development build on the device)
make db-start     # local Supabase via npx supabase
```

Every target is defined in the [`Makefile`](Makefile).

---

## Repository layout

```
parking-app/
├── apps/
│   └── mobile/                      # React Native + Expo (@parking-app/mobile)
│       └── src/
│           ├── features/{search,departure,prompts,map,reputation}/{ui,application,domain,infrastructure}
│           ├── native/              # Expo modules (activity recognition, Bluetooth)
│           └── shared/              # api client, i18n (es-ES default), design system
├── services/
│   ├── core/                        # Go modular monolith (Vercel Functions)
│   │   ├── api/                     # Vercel entrypoint (thin)
│   │   ├── cmd/local/               # local HTTP server
│   │   └── internal/
│   │       ├── availability/        # CORE: Spot aggregate, confidence, TTL
│   │       ├── matching/            # CORE: search sessions, eligibility, waves, claims
│   │       ├── community/           # reputation & points
│   │       ├── zones/               # parkable segments, priors, regulated zones, occupancy
│   │       ├── coverage/            # cities, enabled sources, per-city policies
│   │       ├── identity/            # users / devices (Supabase Auth bridge)
│   │       ├── shared/              # events, outbox, timers, geo, clock, push
│   │       ├── bootstrap/           # composition root and HTTP router
│   │       └── archtest/            # architecture tests enforcing module boundaries
│   ├── ingestion/                   # Python: OSM, SER, DGT NAP, traffic → zones tables
│   └── vision/                      # Python: camera polling + inference (feature-flagged OFF)
├── supabase/                        # config.toml, migrations/, cron/, seed/
├── contracts/
│   ├── events/                      # CloudEvents envelope + event JSON Schemas
│   └── api/                         # OpenAPI
├── docs/
├── .github/workflows/               # CI
├── Makefile                         # cross-toolchain entrypoint
├── package.json, pnpm-workspace.yaml
└── pyproject.toml                   # uv workspace
```

---

## Initialization checklist (Phase 0)

### Accounts & projects
- [x] Create Supabase project (Free) in an EU region; enable PostGIS, `pg_cron`, `pg_net`.
- [ ] Enable **anonymous sign-ins** in Supabase Auth.
- [x] Create Vercel project for `services/core`; deploy `/healthz` in Paris (`cdg1`), matching Supabase `eu-west-3`.
- [ ] Create Expo account + EAS project; configure FCM (Android) and APNs (iOS) credentials for Expo Push.
- [x] Create GitHub repository — public, so `services/vision` can be scheduled on GitHub Actions (see ADR-0014).

### Technical spikes (validate before building on them)
- [ ] **Spike 1 — Go on Vercel:** health endpoint deploys successfully; still measure cold start and p95 latency against the Supabase pooler.
- [ ] **Spike 2 — Timers:** local `pg_cron` accepted and ran a 5 s job; still validate `pg_net` POST to `/internal/timers/fire` only when timers are due, and measure drift.
- [ ] **Spike 3 — Departure detection:** Expo dev client + native modules (activity recognition, Bluetooth connect) on Android and iOS; measure battery.
- [ ] **Spike 4 — Expo Push latency:** server → device p95 with high priority and short TTL.

### Scaffolding
- [x] Monorepo skeleton as above; Go module with architecture tests (`internal/archtest`) enforcing module boundaries.
- [x] Initial Supabase migrations for Coverage, Spots, outbox and timers (see ARCHITECTURE.md §6).
- [x] Load the IGN Comunidad de Madrid boundary and activate its Coverage area through a migration. See [source and attribution](docs/data-sources/DATA_SOURCES.md#coverage-boundary) and apply the migration to each environment separately.
- [x] Contract: `availability.spot.announced.v1` event schema.
- [x] Hourly retention: aggregate Spots by H3 cell and UTC day, delete individual Spots after 23 hours, and retain outbox payloads without Releaser identifiers for seven days.
- [ ] OpenAPI: envelope schema, `/healthz` and `POST /v1/spots` are defined; remaining `/v1` routes are pending.
- [x] CI: lint + tests + architecture tests for Go; lint + type-check for mobile; local migration and PostgreSQL adapter checks.
- [ ] Ingestion: OSM street parking + Madrid SER loaded for Comunidad de Madrid.
- [ ] Keep-alive workflow for Supabase during development (free projects pause after 7 days idle).

---

## Environment variables (core)

| Variable | Purpose |
|---|---|
| `DATABASE_URL` | Supabase pooler connection string (transaction mode) |
| `SUPABASE_URL` | Supabase project URL |
| `SUPABASE_PUBLISHABLE_KEY` | Publishable API key for server-side access-token verification through Supabase Auth |
| `EXPO_ACCESS_TOKEN` | Authenticated access to Expo Push API |
| `INTERNAL_TIMERS_SECRET` | Shared secret for `pg_net` → `/internal/timers/fire` |
| `FEATURE_VISION_ENABLED` | `false` in MVP |
| `DEFAULT_CITY` | `madrid-region` |

Copy [`.env.example`](.env.example) to `.env`; never commit real values.

### Spot announcement API

`POST /v1/spots` accepts a confirmed departure with `latitude`, `longitude` and positive `accuracyM`. Send `Content-Type: application/json`, the driver's Supabase access token as `Authorization: Bearer …`, and a UUID `Idempotency-Key`. The server verifies the token through Supabase Auth `/auth/v1/user`, derives the Releaser identity, locates active Coverage, calculates H3 and confidence from Coverage policies, and saves the Spot and CloudEvent atomically. New Releasers use the approved provisional confidence defaults (`0.8 × 0.8`); reputation-derived reliability remains pending. A successful response is HTTP 201 with `id` and `expiresAt`; it does not yet trigger matching or push.

Configure `DATABASE_URL`, `SUPABASE_URL` and `SUPABASE_PUBLISHABLE_KEY` before enabling writes. The database URL must use an appropriately provisioned backend role, never a credential sent to the app; runtime role provisioning and hosted environment configuration are not performed by this implementation. Missing configuration fails closed while `/healthz` remains a liveness probe. The current verifier uses Auth rather than the scaffolded `SUPABASE_JWKS_URL`; this adds an Auth round trip whose latency still needs measuring. Transactions use no prepared statements for transaction-pooler compatibility.

The two new announcement migrations must be applied before using the API. Idempotency receipts inherit Spot retention through cascading deletion and contain no coordinates. Identical retries retain the original ID and expiry; different input with the same driver/key returns 409. Run `make core-integration` and `make db-test` against local Supabase. Neither API deployment nor applying these migrations to production is implied by committing the implementation.

---

## Contributing

Work test-first (TDD), use Conventional Commits with a module scope (`feat(matching): …`), keep `make check` green before opening a pull request, and record any significant decision as a new ADR in [`docs/adr/`](docs/adr/README.md).

---

## Language policy

- Code, docs, ADRs, commits: **English**.
- User-facing strings: i18n with **Spanish (`es-ES`) as default** locale.
