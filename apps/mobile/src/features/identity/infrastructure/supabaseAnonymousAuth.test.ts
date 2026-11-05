import { createSupabaseAnonymousAuth } from './supabaseAnonymousAuth';
import type { KeyValueStore } from './encryptedStorage';

const URL = 'http://auth.example.test';
const PUBLISHABLE_KEY = 'sb_publishable_test';

function base64Url(value: object): string {
  return btoa(JSON.stringify(value)).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

function sessionBody(accessToken: string, secondsUntilExpiry = 3600) {
  const nowSeconds = Math.floor(Date.now() / 1000);
  return {
    access_token: accessToken,
    token_type: 'bearer',
    expires_in: secondsUntilExpiry,
    expires_at: nowSeconds + secondsUntilExpiry,
    refresh_token: 'refresh-token',
    user: { id: 'driver-1', aud: 'authenticated', role: 'authenticated', is_anonymous: true },
  };
}

function jwt(): string {
  const exp = Math.floor(Date.now() / 1000) + 3600;
  return [
    base64Url({ alg: 'HS256', typ: 'JWT' }),
    base64Url({ sub: 'driver-1', role: 'authenticated', exp }),
    'signature',
  ].join('.');
}

class MemoryStore implements KeyValueStore {
  readonly entries = new Map<string, string>();

  async getItem(key: string): Promise<string | null> {
    return this.entries.get(key) ?? null;
  }

  async setItem(key: string, value: string): Promise<void> {
    this.entries.set(key, value);
  }

  async removeItem(key: string): Promise<void> {
    this.entries.delete(key);
  }
}

interface RecordedRequest {
  url: string;
  headers: Headers;
  body: unknown;
}

function fakeGoTrue(respond: (url: string) => Response) {
  const requests: RecordedRequest[] = [];
  const fetchImpl: typeof fetch = async (input, init) => {
    const request = new Request(input, init);
    requests.push({
      url: request.url,
      headers: request.headers,
      body: await request.json().catch(() => null),
    });
    return respond(request.url);
  };
  return { requests, fetchImpl };
}

function jsonResponse(body: unknown, status = 200): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  });
}

describe('createSupabaseAnonymousAuth', () => {
  it('has no session before the first sign-in', async () => {
    const { fetchImpl } = fakeGoTrue(() => jsonResponse({}));
    const auth = createSupabaseAnonymousAuth({
      url: URL,
      publishableKey: PUBLISHABLE_KEY,
      storage: new MemoryStore(),
      fetch: fetchImpl,
    });

    await expect(auth.currentSession()).resolves.toBeNull();
  });

  it('signs in anonymously sending the CAPTCHA token and the publishable key', async () => {
    const token = jwt();
    const { requests, fetchImpl } = fakeGoTrue(() => jsonResponse(sessionBody(token)));
    const auth = createSupabaseAnonymousAuth({
      url: URL,
      publishableKey: PUBLISHABLE_KEY,
      storage: new MemoryStore(),
      fetch: fetchImpl,
    });

    const session = await auth.signInAnonymously('turnstile-token');

    expect(session).toEqual({ accessToken: token });
    expect(requests).toHaveLength(1);
    expect(requests[0]?.url).toBe(`${URL}/auth/v1/signup`);
    expect(requests[0]?.headers.get('apikey')).toBe(PUBLISHABLE_KEY);
    expect(requests[0]?.body).toMatchObject({
      gotrue_meta_security: { captcha_token: 'turnstile-token' },
    });
  });

  it('restores the signed-in session after the app restarts', async () => {
    const token = jwt();
    const storage = new MemoryStore();
    const { fetchImpl } = fakeGoTrue(() => jsonResponse(sessionBody(token)));
    const options = { url: URL, publishableKey: PUBLISHABLE_KEY, storage, fetch: fetchImpl };
    await createSupabaseAnonymousAuth(options).signInAnonymously('turnstile-token');

    const afterRestart = createSupabaseAnonymousAuth(options);

    await expect(afterRestart.currentSession()).resolves.toEqual({ accessToken: token });
  });

  it('fails and stores no session when Supabase rejects the CAPTCHA', async () => {
    const storage = new MemoryStore();
    const { fetchImpl } = fakeGoTrue(() =>
      jsonResponse(
        {
          code: 400,
          error_code: 'captcha_failed',
          msg: 'captcha protection: request disallowed (invalid-input-response)',
        },
        400,
      ),
    );
    const auth = createSupabaseAnonymousAuth({
      url: URL,
      publishableKey: PUBLISHABLE_KEY,
      storage,
      fetch: fetchImpl,
    });

    await expect(auth.signInAnonymously('bad-token')).rejects.toThrow('anonymous sign-in');
    await expect(auth.currentSession()).resolves.toBeNull();
    expect(storage.entries.size).toBe(0);
  });

  it('fails instead of reporting no session while the stored session cannot be refreshed offline', async () => {
    const storage = new MemoryStore();
    let online = true;
    const { fetchImpl } = fakeGoTrue((url) => {
      if (online || !url.includes('grant_type=refresh_token')) {
        return jsonResponse(sessionBody(jwt(), -10));
      }
      throw new TypeError('Network request failed');
    });
    const options = { url: URL, publishableKey: PUBLISHABLE_KEY, storage, fetch: fetchImpl };
    await createSupabaseAnonymousAuth(options).signInAnonymously('turnstile-token');
    online = false;
    jest.useFakeTimers();
    const errorLog = jest.spyOn(console, 'error').mockImplementation(() => {});

    try {
      const afterRestart = createSupabaseAnonymousAuth(options);
      const outcome = expect(afterRestart.currentSession()).rejects.toThrow('read stored session');
      await jest.advanceTimersByTimeAsync(120_000);
      await outcome;
    } finally {
      errorLog.mockRestore();
      jest.useRealTimers();
    }
  });
});
