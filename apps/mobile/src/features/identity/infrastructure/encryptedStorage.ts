import { gcm } from '@noble/ciphers/aes.js';
import { bytesToHex, bytesToUtf8, hexToBytes, utf8ToBytes } from '@noble/ciphers/utils.js';

export interface KeyValueStore {
  getItem(key: string): Promise<string | null>;
  setItem(key: string, value: string): Promise<void>;
  removeItem(key: string): Promise<void>;
}

interface Dependencies {
  keyStore: KeyValueStore;
  dataStore: KeyValueStore;
  randomBytes: (length: number) => Uint8Array;
}

const KEY_BYTES = 32;
const NONCE_BYTES = 12;
const KEY_PATTERN = /^[0-9a-f]{64}$/;

// SecureStore may reject values over ~2 KB on iOS and a Supabase session is larger, so only the
// AES key goes to SecureStore and the encrypted session goes to the data store.
export function createEncryptedStorage({
  keyStore,
  dataStore,
  randomBytes,
}: Dependencies): KeyValueStore {
  const loading = new Map<string, Promise<Uint8Array>>();

  async function loadOrCreateKey(key: string): Promise<Uint8Array> {
    const stored = await keyStore.getItem(key);
    if (stored && KEY_PATTERN.test(stored)) return hexToBytes(stored);
    const created = randomBytes(KEY_BYTES);
    await keyStore.setItem(key, bytesToHex(created));
    return created;
  }

  function keyFor(key: string): Promise<Uint8Array> {
    const inFlight = loading.get(key);
    if (inFlight) return inFlight;
    const pending = loadOrCreateKey(key).finally(() => loading.delete(key));
    loading.set(key, pending);
    return pending;
  }

  return {
    async getItem(key) {
      const [encrypted, storedKey] = await Promise.all([
        dataStore.getItem(key),
        keyStore.getItem(key),
      ]);
      if (!encrypted || !storedKey) return null;
      try {
        const payload = hexToBytes(encrypted);
        const nonce = payload.slice(0, NONCE_BYTES);
        const ciphertext = payload.slice(NONCE_BYTES);
        return bytesToUtf8(gcm(hexToBytes(storedKey), nonce).decrypt(ciphertext));
      } catch {
        return null;
      }
    },
    async setItem(key, value) {
      const nonce = randomBytes(NONCE_BYTES);
      const ciphertext = gcm(await keyFor(key), nonce).encrypt(utf8ToBytes(value));
      const payload = new Uint8Array(nonce.length + ciphertext.length);
      payload.set(nonce);
      payload.set(ciphertext, nonce.length);
      await dataStore.setItem(key, bytesToHex(payload));
    },
    async removeItem(key) {
      await Promise.all([dataStore.removeItem(key), keyStore.removeItem(key)]);
    },
  };
}
