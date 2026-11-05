import type { AnonymousAuth, CaptchaChallenge, Session } from './ports';

interface Dependencies {
  auth: AnonymousAuth;
  captcha: CaptchaChallenge;
}

export async function ensureAnonymousSession({ auth, captcha }: Dependencies): Promise<Session> {
  const existing = await auth.currentSession();
  if (existing) return existing;

  const captchaToken = await captcha.solve();
  return auth.signInAnonymously(captchaToken);
}
