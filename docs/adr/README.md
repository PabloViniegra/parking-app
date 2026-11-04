# Architecture Decision Records

Format: lightweight [MADR](https://adr.github.io/madr/). Copy [`template.md`](template.md) for new records. Status values: `Proposed` · `Accepted` · `Superseded by ADR-XXXX` · `Deprecated`.

| ADR | Title | Status |
|---|---|---|
| [0001](0001-record-architecture-decisions.md) | Record architecture decisions | Accepted |
| [0002](0002-event-driven-modular-monolith.md) | Event-driven modular monolith, hexagonal + screaming architecture | Accepted |
| [0003](0003-go-core-python-workers.md) | Go for the core, Python for ingestion and vision | Accepted |
| [0004](0004-vercel-and-supabase-free.md) | Vercel Functions + Supabase Free as platform | Accepted |
| [0005](0005-timers-in-postgres.md) | Timers in Postgres with pg_cron + pg_net | Accepted |
| [0006](0006-no-redis-h3-in-postgres.md) | No Redis in MVP; H3 indexing in Postgres | Accepted |
| [0007](0007-transactional-outbox.md) | Transactional outbox as event bus | Accepted |
| [0008](0008-matching-eligibility-priority-waves.md) | Matching: ETA eligibility, search-time priority, waves of one | Accepted |
| [0009](0009-releaser-explicit-confirmation.md) | Releaser explicit confirmation, opt-in auto-share | Accepted |
| [0010](0010-privacy-by-design.md) | Privacy by design: minimization and 24 h location retention | Accepted |
| [0011](0011-react-native-expo.md) | React Native + Expo (dev client) for mobile | Accepted |
| [0012](0012-expo-push.md) | Expo Push as notification provider | Accepted |
| [0013](0013-maplibre-openfreemap.md) | MapLibre + OpenFreeMap for maps | Accepted |
| [0014](0014-camera-pipeline-feature-flagged.md) | Camera pipeline built but feature-flagged off in MVP | Accepted |
| [0015](0015-regulated-zones-loaded-not-offered.md) | Regulated zones loaded and labelled, not offered in MVP | Accepted |
