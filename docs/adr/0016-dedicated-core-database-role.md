# ADR-0016: Dedicated database role for the announcement API

- **Status:** Accepted
- **Date:** 2026-11-05
- **Deciders:** Pablo Viniegra

## Context

The announcement API requires database access, while client roles must have none. Using `postgres` or `service_role` as runtime credentials would bypass RLS and expose unrelated tables if the credential were compromised. The owner explicitly authorized a dedicated backend role and role-specific policies.

## Decision

- Create `parking_core` without inheritance, RLS bypass, role/database creation, superuser or replication privileges. The migration creates it without login; enable login and set a generated password outside version control during provisioning.
- Grant only the columns and operations used by the announcement API: active Coverage reads, Spot ID/expiry reads and announcement inserts, idempotency receipt reads/inserts, and announced-event inserts. No updates, deletes, maintenance, precise-location reads or event-history reads are needed.
- Add RLS policies exclusively for `parking_core` on these operations. Core is a trusted multi-driver service: it verifies Supabase Auth identity and scopes receipt lookup by Releaser and key. These policies do not constitute end-user row ownership isolation.
- Do not grant this role to `anon`, `authenticated`, or PostgREST's `authenticator`. Their default-deny posture remains unchanged. Tables unrelated to announcements remain inaccessible to the role.
- Store the password-bearing pooler connection string only as a sensitive Vercel production variable. Do not reuse administrator credentials or send this connection string to the mobile app.
- Require `sslmode=verify-full` with `DATABASE_CA_CERT` holding the Supabase root CA. A custom CA with unverified TLS or plaintext configuration is rejected. The public test fixture is the [Supabase Root 2021 CA](https://supabase-downloads.s3-ap-southeast-1.amazonaws.com/prod/ssl/prod-ca-2021.crt), SHA-256 certificate fingerprint `807025AD50D4ED219D2C9C7D299C004F824EB00CF7F65AFEF607D07B72E6CAFA`, valid until 2031-04-26; it contains no private key.

## Consequences

- Positive: the runtime cannot bypass RLS, read stored coordinates or alter existing rows; client access stays blocked.
- Negative / trade-offs: a compromised backend credential can still announce Spots and inspect ID/expiry and receipt metadata across Releasers; HTTP authorization remains a core responsibility.
- Follow-ups: new use cases require explicit grants and policies with tests, rather than expanding this role preemptively. Password rotation is a provisioning operation, not a migration containing a secret.

## Alternatives considered

- Administrator or `service_role` credentials: rejected because they bypass RLS and carry unrelated privileges.
- Client-owned table policies: rejected because the mobile app must use the core API.
