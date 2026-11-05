import { ensureAnonymousSession } from './ensureAnonymousSession';
import type { AnonymousAuth, CaptchaChallenge, Session } from './ports';

class FakeAuth implements AnonymousAuth {
  signInTokens: string[] = [];

  constructor(private stored: Session | null) {}

  async currentSession(): Promise<Session | null> {
    return this.stored;
  }

  async signInAnonymously(captchaToken: string): Promise<Session> {
    this.signInTokens.push(captchaToken);
    this.stored = { accessToken: 'new-access-token' };
    return this.stored;
  }
}

class FakeCaptcha implements CaptchaChallenge {
  solved = 0;

  constructor(private readonly outcome: string | Error) {}

  async solve(): Promise<string> {
    this.solved += 1;
    if (this.outcome instanceof Error) throw this.outcome;
    return this.outcome;
  }
}

describe('ensureAnonymousSession', () => {
  it('reuses the stored session without asking for a CAPTCHA', async () => {
    const auth = new FakeAuth({ accessToken: 'stored-access-token' });
    const captcha = new FakeCaptcha('unused');

    const session = await ensureAnonymousSession({ auth, captcha });

    expect(session).toEqual({ accessToken: 'stored-access-token' });
    expect(captcha.solved).toBe(0);
    expect(auth.signInTokens).toEqual([]);
  });

  it('signs in anonymously with the solved CAPTCHA token when there is no session', async () => {
    const auth = new FakeAuth(null);
    const captcha = new FakeCaptcha('turnstile-token');

    const session = await ensureAnonymousSession({ auth, captcha });

    expect(auth.signInTokens).toEqual(['turnstile-token']);
    expect(session).toEqual({ accessToken: 'new-access-token' });
  });

  it('does not sign in when the CAPTCHA fails', async () => {
    const auth = new FakeAuth(null);
    const captcha = new FakeCaptcha(new Error('challenge failed'));

    await expect(ensureAnonymousSession({ auth, captcha })).rejects.toThrow('challenge failed');
    expect(auth.signInTokens).toEqual([]);
  });
});
