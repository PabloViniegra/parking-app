# Architecture

> Status: **Approved** (2026-11-04). Decisions referenced as ADR-XXXX live in [`docs/adr/`](../adr/README.md).

## 1. Quality attributes (prioritized)

| # | Attribute | Measurable scenario |
|---|---|---|
| 1 | **Latency** | `spot.announced` → push delivered to first Seeker: **p95 < 3 s** |
| 2 | **Information trust** | ≥ **60 %** of offered Spots confirmed free on arrival (MVP target) |
| 3 | **Driving safety** | Zero mandatory interactions while moving; prompts are 1-tap/voice, auto-dismissed |
| 4 | **Privacy (GDPR)** | No plate/model; pseudonymous identity; precise locations deleted within **24 h**; camera images processed in memory, never stored |
| 5 | **Battery** | < **3 %/h** with active session; no continuous GPS outside a session |
| 6 | **Geographic extensibility** | New city = configuration + source adapters; **no core code changes** |
| 7 | **Scalability** | State partitioned by H3 cell; no global coordination on the hot path |
| 8 | **Cost** | Runs on free tiers during MVP (Supabase Free, Vercel Hobby, Expo Push, GitHub Actions) |
| 9 | **Observability** | End-to-end trace per Spot (announced → offered → outcome) |

Attributes 1–3 drive the architectural decisions.

---

## 2. Architectural style

**Event-driven modular monolith** for the core (ADR-0002), with **two separate Python workers** where runtime or scaling differs (ingestion, vision).
Each module follows **hexagonal architecture**; the repository follows **screaming architecture** (top-level folders are business capabilities, not technical layers).

---

## 3. System overview

```
┌──────────── Mobile (React Native + Expo, dev client) ───────────┐
│ features: search · departure · prompts · map · reputation        │
│ native modules: activity recognition · Bluetooth · geofences     │
└──────┬──────────────────────────────▲─────────────▲────────────┘
       │ HTTPS + Supabase JWT          │ Expo Push   │ Supabase Realtime (optional,
┌──────▼──── core (Go modular monolith on Vercel Functions) ──────┐   foreground map only)
│ availability · matching · community · zones · coverage · identity│
│ /v1/*  (app API)          /internal/timers/fire  (system)        │
└──────┬───────────────────────────────▲──────────────────────────┘
       │ Supavisor (transaction mode)   │ pg_net POST, only when timers are due
┌──────▼───────────── Supabase Free (Postgres + PostGIS) ──────────┐
│ spots · search_sessions · claims · timers · outbox · reputation   │
│ parkable_segments · zone_priors · regulated_zones · coverage      │
│ pg_cron 5 s  → fire due timers                                    │
│ pg_cron 1 h  → purge precise locations > 24 h; aggregate by H3    │
│ Auth: anonymous sign-ins                                          │
└──────▲───────────────────────────────▲──────────────────────────┘
       │                               │
┌──────┴── ingestion (Python, GH Actions, weekly) ──┐  ┌── vision (Python, flag OFF) ─┐
│ OSM street parking · SER Madrid · DGT NAP ·       │  │ Informo cameras every 5 min  │
│ traffic intensity → zones tables                  │  │ → zones.occupancy.estimated  │
└───────────────────────────────────────────────────┘  └──────────────────────────────┘
```

---

## 4. Platform & free-tier constraints

| Platform | Constraint | Design response |
|---|---|---|
| **Supabase Free** | 500 MB DB, 5 GB egress, 50k MAU, 2 free projects, **pauses after 7 days idle** | 24 h retention + H3 aggregates; keep-alive workflow during development; upgrade trigger defined below |
| **Vercel Hobby** | Cron limited to **once per day**; ~4 CPU-h/month active CPU; **non-commercial use only** | No Vercel Cron: timers live in Postgres (ADR-0005). Short, I/O-bound functions. Move to Pro or container host if commercialized (ADR-0004) |
| **Expo Push** | Free; best-effort delivery | Use push receipts; short `ttl`, `priority: high` |
| **GitHub Actions** | Private repos have limited free minutes; schedules ≥ 5 min and may drift | Ingestion weekly; vision only in a public repo or alternative host (ADR-0014) |

**Upgrade triggers** (to Supabase Pro / Vercel Pro): DB > 400 MB · real-user traffic where a pause is unacceptable · any commercial use · p95 latency SLO violated due to cold starts.

---

## 5. Key mechanisms

### 5.1 Hot path (synchronous)
1. Releaser confirms prompt → `POST /v1/spots` (lat, lng, accuracy, idempotency key).
2. Availability creates Spot (`ANNOUNCED`), writes event to **outbox** in the same transaction.
3. In-process dispatch → Matching selects the best eligible Seeker (BR-M1/M2) → Spot `OFFERED`, offer timer scheduled.
4. Push sent via Expo (`priority: high`, `ttl` = remaining wave window).
All in one invocation — no queue hop on the critical path.

### 5.2 Timers in Postgres (ADR-0005)
```sql
timers(id, kind, aggregate_id, due_at, payload jsonb, fired_at null, attempts int)
-- kinds: offer_window, claim_expiry, spot_ttl, session_timeout
```
- `pg_cron` every **5 s** runs a SQL function that checks `exists(select 1 from timers where due_at <= now() and fired_at is null)`; **only if true** it calls `pg_net.http_post('/internal/timers/fire')` with a shared secret.
- The core claims due timers with `FOR UPDATE SKIP LOCKED`, executes domain logic, marks them fired. Handlers are **idempotent**.
- Expected drift: ≤ 5 s + invocation latency. Acceptable for 12 s waves and 180 s TTLs.

### 5.3 Transactional outbox (ADR-0007)
- Every state change writes its domain event(s) to `outbox` in the same transaction.
- In-process subscribers handle intra-core reactions synchronously after commit.
- External consumers (analytics, workers) read `outbox` by cursor. Events retained 7 days, then aggregated.

### 5.4 Geo indexing (ADR-0006)
- H3 resolution **10** (~65 m edge) for Spots and sessions; resolution **8** for zone priors.
- H3 computed in Go (`uber/h3-go`), stored as `bigint`, B-tree indexed. Neighbour lookup = `k-ring` in code + `WHERE cell = ANY($1)`.
- PostGIS used for parkable segments, regulated zones and ingestion geometry.

### 5.5 Matching (ADR-0008)
Eligibility → score → wave of one → 12 s window → next. Search time accrues only under BR-M4 conditions, reported by the device in periodic heartbeats (every 20 s during a session).

### 5.6 Departure detection (on device)
- Primary: car Bluetooth connection (user pairs once).
- Secondary: activity recognition transition `STILL/WALKING → IN_VEHICLE` near the last parked location.
- Parking detection (passive): `IN_VEHICLE → WALKING` → stores "my car is here" locally + `spot.occupied` if a claim exists.

---

## 6. Initial data model (core tables)

| Table | Key columns | Retention |
|---|---|---|
| `drivers` | id (auth uid), created_at, settings jsonb | account lifetime |
| `spots` | id, cell_r10, location (geography), status, confidence, ttl_at, releaser_id, coverage_id, created_at | precise location 24 h, then cell only |
| `spot_announcements` | releaser_id, idempotency_key, fingerprint, spot_id | deleted with the Spot by cascade (≤ 24 h) |
| `search_sessions` | id, driver_id, origin_cell, radius_m, search_seconds, last_cell, status, started_at | 24 h |
| `offers` | id, spot_id, seeker_id, wave, offered_at, expires_at, outcome | 24 h → aggregates |
| `claims` | id, spot_id, seeker_id, expires_at, outcome | 24 h → aggregates |
| `timers` | see §5.2 | purged after fired + 1 h |
| `outbox` | id, type, aggregate_id, payload, occurred_at | 7 days |
| `reputation_ledger` | id, driver_id, delta, reason, ref_id, created_at | account lifetime |
| `parkable_segments` | id, geom, side, orientation, source, coverage_id | refreshed weekly |
| `regulated_zones` | id, geom, kind (SER/ORA), source | refreshed weekly |
| `zone_priors` | cell_r8, hour_of_week, p_free, source_mix | rolling |
| `zone_occupancy` | cell_r8, observed_at, estimate, source | 7 days |
| `coverage_areas` | id, name, geom, policies jsonb, enabled_sources | permanent |

**Row Level Security:** enabled on every table, **default deny**. The mobile app never queries tables directly (except Realtime channels, if enabled); all reads/writes go through the core API using a dedicated database role.

The `anon` and `authenticated` roles have no direct privileges on application tables in `public`. New tables, sequences and functions created by `postgres` in `public` do not automatically grant these roles access. The global `PUBLIC` execute default is also removed for future functions created by `postgres`; intentional callers must receive explicit grants. Existing `service_role` grants are preserved. Objects created by other owners, including `supabase_admin`, require a separate privilege review. Run `make db-test` against local Supabase to verify the privilege boundary.

---

## 7. Code structure

### 7.1 Core (Go)
```
services/core/
├── api/index.go                         # Vercel entrypoint → router (thin)
├── cmd/local/main.go                    # local HTTP server for development
└── internal/
    ├── availability/
    │   ├── domain/                      # Spot aggregate, policies, events (pure Go)
    │   ├── application/                 # use cases: AnnounceSpot, RefuteSpot, ExpireSpot…
    │   ├── ports/                       # SpotRepository, Clock, EventPublisher
    │   └── adapters/{http,postgres}/
    ├── matching/      (same layout)     # SearchSession, Offer, Claim, EligibilityPolicy
    ├── community/     (same layout)
    ├── zones/         (same layout)
    ├── coverage/      (same layout)
    ├── identity/      (same layout)
    └── shared/{events,outbox,timers,geo,clock,push,observability}
```
**Dependency rule:** `domain` ← `application` ← `adapters`. `domain` imports nothing outside stdlib + `shared/events` types. Modules never import each other's internals; they communicate via events or published interfaces in `ports`. Enforced in CI (`depguard` / `arch-go`).

### 7.2 Mobile (React Native + Expo)
```
apps/mobile/src/
├── features/
│   ├── search/        # start/stop session, heartbeat
│   ├── departure/     # Bluetooth + activity detection, departure prompt
│   ├── prompts/       # Google-Maps-style binary prompts, TTS
│   ├── map/           # MapLibre, zone priors layer
│   └── reputation/
│   each: ui/ · application/ · domain/ · infrastructure/
├── native/            # Expo modules (activity recognition, BT)
└── shared/            # api client, i18n (es-ES default), design system
```

### 7.3 Workers (Python)
```
services/ingestion/sources/{osm,madrid_ser,dgt_nap,madrid_traffic}/   # one adapter per source
services/vision/{cameras,inference,publisher}/
```
Each source implements a `PriorSource` / `SignalSource` port and is enabled per Coverage area.

---

## 8. Privacy by design (ADR-0010)

- **Data minimization:** no plate, no model, no name required (anonymous sign-in).
- **Ephemeral location:** precise coordinates kept ≤ 24 h; afterwards only H3 aggregates without driver id.
- **Purpose limitation:** location used only during an active session or a departure/parking event.
- **Cameras:** images fetched to memory, inferred, discarded; only counts per cell are stored.
- **Transparency:** clear consent screens for location, motion and Bluetooth; settings to delete account and data.
- A legal/GDPR review is required before public launch.

---

## 9. Observability

- Structured logs with `spot_id` as correlation id across events.
- Metrics: announce→push latency, offer acceptance rate, confirmed-free rate, refutation rate, timer drift.
- Error tracking: free tier of a hosted tool (e.g., Sentry) for core and mobile.
