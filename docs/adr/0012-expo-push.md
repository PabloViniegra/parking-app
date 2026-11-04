# ADR-0012: Expo Push as notification provider

- **Status:** Accepted
- **Date:** 2026-11-04

## Context
Push is the main channel of the product. It must be free and integrate with Expo.

## Decision
- Use the **Expo Push Service** (free) over FCM/APNs, behind a `PushSender` port.
- Offers sent with `priority: high` and a short `ttl` equal to the remaining wave window.
- Push receipts checked to detect invalid tokens.

## Consequences
- Positive: zero cost, one API for both platforms.
- Negative: extra third party in the delivery path; best-effort delivery. A direct FCM/APNs adapter can replace it behind the port.
