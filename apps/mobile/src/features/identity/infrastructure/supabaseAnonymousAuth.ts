import { createClient } from '@supabase/supabase-js';
import { AppState } from 'react-native';

import type { AnonymousAuth, Session } from '../application/ports';
import { bindAutoRefresh } from './bindAutoRefresh';
import type { KeyValueStore } from './encryptedStorage';

interface Options {
  url: string;
  publishableKey: string;
  storage: KeyValueStore;
  fetch?: typeof fetch;
}

export function createSupabaseAnonymousAuth({
  url,
  publishableKey,
  storage,
  fetch,
}: Options): AnonymousAuth {
  const client = createClient(url, publishableKey, {
    auth: {
      storage,
      autoRefreshToken: true,
      persistSession: true,
      detectSessionInUrl: false,
    },
    global: fetch ? { fetch } : {},
  });
  bindAutoRefresh(client.auth, AppState);

  return {
    async currentSession(): Promise<Session | null> {
      const { data, error } = await client.auth.getSession();
      if (error) throw new Error('read stored session', { cause: error });
      return data.session ? { accessToken: data.session.access_token } : null;
    },
    async signInAnonymously(captchaToken: string): Promise<Session> {
      const { data, error } = await client.auth.signInAnonymously({ options: { captchaToken } });
      if (error || !data.session) throw new Error('anonymous sign-in', { cause: error });
      return { accessToken: data.session.access_token };
    },
  };
}
