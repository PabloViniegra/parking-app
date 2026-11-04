# Domain Definition

> Status: **Approved** (2026-11-04). Changes to this document require an ADR or explicit review.

## 1. What the business sells

We do not sell "parking" nor "cameras". We sell **perishable information**:

> *"There is a free spot here, right now, and it will probably still be free when you get there."*

The north-star metric is the **seeker's time-to-park**. The trust metric is the **share of offered spots that are confirmed free on arrival**.

Every mechanism in the product — driver collaboration, open data, cameras, push notifications — exists only to produce that information with the **highest confidence** and **lowest latency** possible.

**Primary source of truth:** collaboration between drivers.
**Secondary sources:** open data priors (OSM, SER, DGT NAP, traffic) and public cameras (zone-level occupancy).

---

## 2. Subdomains

| Type | Subdomain | Why |
|---|---|---|
| **Core** | **Availability** — spots, confidence, decay, TTL | The differentiating capability |
| **Core** | **Matching** — who gets notified, in which order | Defines the seeker experience and fairness |
| Supporting | Departure Detection — Bluetooth, activity recognition, geofences (on device) | Removes friction; relies on OS APIs |
| Supporting | Zones — parkable segments, priors, regulated zones, occupancy estimates | Solves cold start; zone-level, never spot-level |
| Supporting | Urban Vision — camera ingestion + inference | Feeds zone occupancy (feature-flagged in MVP) |
| Supporting | Community — reputation points | Sustains collaboration, filters noise |
| Supporting | Coverage — cities, enabled sources, per-city policies | Enables expansion to all of Spain by configuration |
| Generic | Identity, Push delivery, Maps, Geocoding | Bought / managed services |

---

## 3. Ubiquitous language

| Term | Definition |
|---|---|
| **Driver** | A user of the app. Plays one of two **roles** depending on the moment. |
| **Releaser** | Driver who is leaving a parking spot. |
| **Seeker** | Driver who is actively searching for a spot. |
| **Spot** | A *temporary claim* that a free street parking space exists at a location. **Not** a registered physical bay. |
| **Signal** | Any observation feeding a Spot or Zone: departure confirmation, verification answer, refutation, passive parking detection, camera estimate. |
| **Departure Intent** | Evidence that a driver is about to leave (car Bluetooth connected, vehicle motion starting). Not yet a Spot. |
| **Confidence** | Probability (0–1) that a Spot is still free. Decays over time; depends on source and reputation. |
| **TTL** | Moment after which a Spot is no longer offered. |
| **Search Session** | Period during which a Seeker is actively searching within a zone and radius. Only Seekers with an active session can be notified. |
| **Search Time** | Accumulated minutes a Seeker has been *effectively* searching (see BR-M4). Drives priority. |
| **Eligibility** | Whether a Seeker can be offered a Spot (reachable before TTL, within radius). |
| **Wave** | One round of offering a Spot to exactly **one** Seeker, with a short acceptance window. |
| **Claim** | A Seeker accepts a Spot; it is reserved for them until the claim expires. |
| **Verification Prompt** | Google-Maps-style binary question ("Is the spot still free?") shown to drivers passing by. |
| **Zone / Cell** | Geographic aggregation unit: H3 cell. |
| **Prior** | Pre-computed probability of finding parking in a cell at a given time (from open data + own history). |
| **Parkable Segment** | Street segment where street parking is allowed (from OSM). |
| **Regulated Zone** | Area with paid / time-limited parking (SER in Madrid city, ORA elsewhere). Out of MVP scope but **always loaded and labelled**. |
| **Reputation** | Points measuring a driver's reliability as a signal source. |

---

## 4. Aggregates

### 4.1 Spot (Availability) — central aggregate

```
             confirm (Releaser)
DepartureIntent ───────────────► ANNOUNCED
                                    │ eligible seeker found
                                    ▼
                    ┌──────────► OFFERED ───── wave window elapsed ──┐
                    │               │                               │ next eligible seeker
                    │               │ seeker accepts                │
                    │               ▼                               │
                    │           CLAIMED                             │
                    │   claim expired │ │ seeker parked (passive)   │
                    └─────────────────┘ ▼                           │
                                    OCCUPIED (terminal)             │
                                                                    │
  ANNOUNCED / OFFERED / CLAIMED ── TTL reached ──► EXPIRED (terminal)
  ANNOUNCED / OFFERED / CLAIMED ── refutation ───► REFUTED (terminal)
```

**Invariants**
- A Spot has at most **one active offer** and at most **one active claim** at any time.
- A Spot in a terminal state never changes again.
- A Spot is never created from a camera signal.
- A Spot always belongs to exactly one H3 cell and one Coverage area.

### 4.2 SearchSession (Matching)
- One active session per Driver.
- Holds: origin cell, radius, started at, **accumulated search time**, last known position (ephemeral).
- Ends when: driver parks (passive detection), cancels, or session times out.

### 4.3 Claim (Matching)
- Belongs to one Spot and one Seeker. Expires at `min(spot.ttl, now + eta + claim_grace)`.

### 4.4 ReputationAccount (Community)
- Points balance + reliability score per Driver.

### 4.5 Zone (Zones)
- Per H3 cell: parkable capacity, regulated flag, prior by hour-of-week, latest occupancy estimate.

---

## 5. Business rules

### Availability
- **BR-A1** A Spot is created **only** when the Releaser **explicitly confirms** the departure prompt. No answer = no Spot. (Opt-in setting "Always share automatically" exists, default OFF.) → ADR-0009
- **BR-A2** Default TTL: **180 s** (configurable per city and per zone density).
- **BR-A3** Confidence starts at `base(source) × reliability(releaser)` and decays linearly to 0 at TTL. Confirmations raise it; refutations end the Spot.
- **BR-A4** One **strong refutation** (seeker arrived, did not park and answered "No") or **two weak refutations** (passers-by) → `REFUTED`.
- **BR-A5** Spots in regulated zones are labelled; in MVP they are **not offered**.
- **BR-A6** We never ask for or store licence plate or car model.

### Matching
- **BR-M1 Eligibility:** a Seeker is eligible iff session active ∧ Spot within session radius ∧ `ETA ≤ remaining TTL`.
- **BR-M2 Priority:** eligible Seekers are ranked by `score = searchMinutes × w1 − etaMinutes × w2` (initial `w1 = 1.0`, `w2 = 0.3`). Rewards time invested without sending anyone to an unreachable spot.
- **BR-M3 Waves of one:** offer to **exactly one** Seeker per wave; wave window **12 s**. No acceptance → next Seeker. → ADR-0008
- **BR-M4 Anti-gaming:** search time only accrues while the device reports *in vehicle*, inside the session zone, at cruising speed (< 25 km/h).
- **BR-M5** A Seeker who ignores **3 consecutive offers** has their session paused (they are probably not searching anymore).
- **BR-M6** Live Spots are delivered **only through matching**; they are not shown publicly on the map (preserves fairness). The map shows zone priors. *(Proposed — see Open Questions.)*

### Verification
- **BR-V1** Verification prompts are binary, large buttons, auto-dismiss in **8 s**. No answer = no signal.
- **BR-V2** Prompted only to drivers passing within ~30 m of an OFFERED/CLAIMED Spot, max 1 prompt per driver per 5 min.

### Community
- **BR-C1** Initial points: Spot confirmed OCCUPIED by a Seeker → Releaser **+10**; verification answer consistent with outcome → **+1**; Spot REFUTED → Releaser **−5**.
- **BR-C2** Reliability score (0–1) derived from the ratio of confirmed vs refuted Spots, with a prior for new users.

### Safety
- **BR-S1** No mandatory interaction while the vehicle is moving. Everything asked while driving is one tap or voice, read aloud (TTS), and auto-dismissed.
- **BR-S2** The departure prompt is shown while the car is still stationary (Bluetooth connect / engine start before motion).

All numeric values above are **initial defaults**, configurable per Coverage area and to be tuned with real data.

---

## 6. Domain events

Naming: `<context>.<entity>.<past-tense>.v<version>`, CloudEvents envelope.

| Event | Producer | Main consumers |
|---|---|---|
| `departure.intent.detected.v1` | Mobile (Departure) | — (local prompt) |
| `availability.spot.announced.v1` | Availability | Matching, Analytics |
| `matching.spot.offered.v1` | Matching | Push, Analytics |
| `matching.offer.expired.v1` | Matching (timer) | Matching (next wave) |
| `matching.spot.claimed.v1` | Matching | Availability, Push |
| `matching.claim.expired.v1` | Matching (timer) | Availability, Matching |
| `availability.spot.occupied.v1` | Availability | Community, Matching (end session), Zones |
| `availability.spot.refuted.v1` | Availability | Community, Matching |
| `availability.spot.expired.v1` | Availability (timer) | Matching, Analytics |
| `availability.spot.verified.v1` | Availability | Community |
| `matching.search.started.v1` / `matching.search.ended.v1` | Matching | Analytics |
| `zones.occupancy.estimated.v1` | Urban Vision | Zones |
| `community.reputation.adjusted.v1` | Community | Availability (reliability) |

---

## 7. Bounded contexts & context map

```
                ┌──────────────┐  spot.announced   ┌──────────────┐
  Mobile ──────►│ Availability │──────────────────►│   Matching   │──► Push (Expo)
 (Departure)    │    (core)    │◄──────────────────│    (core)    │
                └──────┬───────┘ claimed/expired   └──────┬───────┘
                       │ occupied/refuted                 │ uses ETA, priors
                       ▼                                  ▼
                ┌──────────────┐                   ┌──────────────┐
                │  Community   │                   │    Zones     │◄── Ingestion (OSM, SER, NAP)
                └──────────────┘                   └──────────────┘◄── Urban Vision (cameras)
                                    ┌──────────────┐
                 all contexts ─────►│   Coverage   │ (cities, sources, policies)
                                    └──────────────┘
```

- Availability ↔ Matching: **partnership**, communicate via domain events only.
- Zones ← Ingestion / Urban Vision: **conformist + anti-corruption layer** (each source has its own adapter translating to Zones' model).
- Coverage: **published language** (policies read by all contexts).
- Identity: **open host service** wrapping Supabase Auth.

---

## 8. Open questions

| # | Question | Proposed default |
|---|---|---|
| OQ-1 | Should live Spots be visible on the map to all Seekers? | **No** in MVP (BR-M6) |
| OQ-2 | Exact radius of a Search Session | 600 m default, user-adjustable 300–1000 m |
| OQ-3 | Should points be redeemable in the future? | Not in MVP; keep ledger design compatible |
| OQ-4 | How do we label regulated zones outside Madrid city? | OSM tags + crowd signal "this street is regulated" |
