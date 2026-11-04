# MVP Definition

> Status: **Approved** (2026-11-04)

## 1. Scope

- **Geography:** Comunidad de Madrid (all municipalities).
- **Parking type:** free street parking. Regulated zones (SER in Madrid city, ORA elsewhere) are **loaded and labelled but not offered**.
- **Primary mechanism:** driver collaboration. Cameras: pipeline built, **feature-flagged OFF**.

## 2. Functional requirements

| ID | Requirement | Acceptance criteria (summary) |
|---|---|---|
| FR-01 | Frictionless sign-up | Anonymous Supabase sign-in on first launch; no personal data required |
| FR-02 | Start / stop search session | One tap or voice; ends automatically when parking is detected; heartbeat every 20 s |
| FR-03 | Departure detection | Optional car Bluetooth pairing; activity-recognition fallback; prompt shown while stationary |
| FR-04 | Departure prompt (Releaser) | "Are you leaving the spot?" Yes / No, 8 s countdown; **no answer = not shared**; opt-in "always share" setting (default OFF); manual "I'm leaving" button |
| FR-05 | Matching | Eligibility by ETA ≤ remaining TTL and radius; priority by search time; waves of **one** Seeker; 12 s window |
| FR-06 | Spot offer push | "Spot 120 m away" with **Go** (deep link to Google Maps / Waze / Apple Maps) / **Ignore**; read aloud via TTS |
| FR-07 | Claim | Accepting reserves the Spot; claim expires at min(TTL, ETA + grace); on expiry the Spot returns to matching |
| FR-08 | Verification prompts | Drivers passing near an offered Spot get "Is the spot still free?" Yes / No, auto-dismiss 8 s |
| FR-09 | Passive parking detection | `IN_VEHICLE → WALKING` ends session, confirms claim (`spot.occupied`), saves "where's my car" locally |
| FR-10 | Reputation points | +10 confirmed, +1 consistent verification, −5 refuted; visible in profile |
| FR-11 | Zone probability map | Heatmap of priors per H3 cell from OSM + SER + own history; regulated zones labelled |
| FR-12 | Camera base pipeline | Informo catalog, 5-min polling, basic vehicle count, `zones.occupancy.estimated`; feeds priors only; flag OFF in UI |
| FR-13 | Multi-city configuration | Coverage areas with policies and enabled sources; only Comunidad de Madrid active |
| FR-14 | Privacy controls | Consent screens; delete account & data; 24 h location retention |

## 3. Notification UX (Google-Maps style)

| Moment | Prompt | Default if no answer |
|---|---|---|
| Releaser, car stationary, BT connected / motion imminent | "Are you leaving the spot?" | **Not shared** |
| Seeker, Spot offered | "Free spot 120 m away — Go / Ignore" (TTS) | Offer expires after 12 s |
| Passer-by near an offered Spot | "Is the spot still free?" Yes / No | No signal |
| Seeker arrives | *(no prompt)* — passive parking detection | Not parked after arrival = weak refutation |

Rules: large buttons, single tap, auto-dismiss, read aloud, max 1 verification prompt per driver per 5 min.

## 4. Out of MVP

CarPlay / Android Auto · cameras visible in UI · regulated-zone parking · cities outside Comunidad de Madrid · point redemption · predictive spot forecasting · live spots on public map.

## 5. Success metrics

| Metric | MVP target |
|---|---|
| Offered Spots confirmed free on arrival | ≥ 60 % |
| Median time-to-park for Seekers using offers | < 5 min |
| Departure prompt acceptance rate | ≥ 30 % |
| Offer acceptance rate (first wave) | ≥ 40 % |
| D7 retention | ≥ 25 % |
| Announce → push latency p95 | < 3 s |

## 6. Phases

| Phase | Content |
|---|---|
| **F0 — Foundations** | Monorepo, Supabase + Vercel + Expo setup, spikes (Go on Vercel, timers, departure detection, push latency), contracts, OSM + SER ingestion |
| **F1 — MVP** | FR-01 … FR-14 in Comunidad de Madrid |
| **F2** | Cameras visible, CarPlay / Android Auto (request Apple entitlement early), regulated zones |
| **F3** | Expansion to other cities via Coverage, predictive priors |
