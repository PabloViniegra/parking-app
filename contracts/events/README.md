# Event contracts

- Every domain event uses the CloudEvents envelope in [`envelope.schema.json`](envelope.schema.json).
- One payload schema per event version: `<context>/<entity>.<past-tense>.v<n>.schema.json`, describing `data`.
- Versions are immutable once published. A breaking change is a new `v<n+1>` file; producers emit both until consumers migrate.
- The catalogue of events lives in `docs/domain/DOMAIN.md` §6.
