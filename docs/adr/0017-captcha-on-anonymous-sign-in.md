# ADR-0017: CAPTCHA on anonymous sign-in

- **Status:** Accepted
- **Date:** 2026-11-05
- **Deciders:** Pablo Viniegra

## Context

FR-01 allows anonymous sign-in with no personal data. Without friction, scripts can mass-create anonymous users and abuse the announcement API; the per-IP rate limit alone is weak against distributed clients. The project runs on free tiers (ADR-0004), so the CAPTCHA must be free.

## Decision

- Enable Supabase Auth CAPTCHA with Cloudflare Turnstile, which is free. It applies to anonymous sign-in, so the mobile app must pass a `captchaToken` to `signInAnonymously`.
- The anonymous session is persisted encrypted (AES-GCM): the key lives in `expo-secure-store` and the ciphertext in AsyncStorage, because SecureStore may reject values over ~2 KB on iOS. If the key or data is lost or tampered with, the app treats it as no session and asks for a new CAPTCHA.
- The secret is read from `SUPABASE_AUTH_CAPTCHA_SECRET` and configured in the hosted project's Auth settings; it is never committed.

## Consequences

- Positive: automated mass registration becomes costly; no personal data is added (ADR-0010); no cost.
- Negative / trade-offs: Cloudflare receives the device IP and challenge signals; Turnstile has no official React Native SDK, so the widget runs in a WebView.
- Follow-ups: integrate the Turnstile widget in the mobile sign-in flow (`signInAnonymously({ options: { captchaToken } })`); enable it in the hosted project's Auth settings. Local tests insert `auth.users` directly and are unaffected.

## Alternatives considered

- hCaptcha: official React Native SDK, but useful features are paid.
- Rate limit only: insufficient against distributed abuse.
