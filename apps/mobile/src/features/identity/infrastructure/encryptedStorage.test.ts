import { createEncryptedStorage, type KeyValueStore } from './encryptedStorage';

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

function sequentialRandomBytes() {
  let counter = 0;
  return (length: number) => {
    counter += 1;
    return new Uint8Array(length).fill(counter);
  };
}

function setup() {
  const keyStore = new MemoryStore();
  const dataStore = new MemoryStore();
  const storage = createEncryptedStorage({
    keyStore,
    dataStore,
    randomBytes: sequentialRandomBytes(),
  });
  return { keyStore, dataStore, storage };
}

describe('createEncryptedStorage', () => {
  it('returns null for a key that was never stored', async () => {
    const { storage } = setup();

    await expect(storage.getItem('session')).resolves.toBeNull();
  });

  it('round-trips a value larger than the 2048 bytes SecureStore may reject', async () => {
    const { storage } = setup();
    const session = JSON.stringify({ access_token: 'a'.repeat(4000), user: { id: 'ñ-✓' } });

    await storage.setItem('session', session);

    await expect(storage.getItem('session')).resolves.toBe(session);
  });

  it('never writes the plain value to the data store and keeps only the key in the key store', async () => {
    const { keyStore, dataStore, storage } = setup();
    const secret = 'super-secret-refresh-token';

    await storage.setItem('session', secret);

    expect(dataStore.entries.get('session')).not.toContain(secret);
    const storedKey = keyStore.entries.get('session');
    expect(storedKey).toBeDefined();
    expect(storedKey).toHaveLength(64);
    expect(keyStore.entries.size).toBe(1);
  });

  it('encrypts the same value differently on every write', async () => {
    const { dataStore, storage } = setup();

    await storage.setItem('session', 'same value');
    const first = dataStore.entries.get('session');
    await storage.setItem('session', 'same value');
    const second = dataStore.entries.get('session');

    expect(first).not.toBe(second);
    await expect(storage.getItem('session')).resolves.toBe('same value');
  });

  it('reuses the stored key when an entry is written again', async () => {
    const { keyStore, storage } = setup();
    await storage.setItem('session', 'first');
    const firstKey = keyStore.entries.get('session');

    await storage.setItem('session', 'second');

    expect(keyStore.entries.get('session')).toBe(firstKey);
    await expect(storage.getItem('session')).resolves.toBe('second');
  });

  it('creates a single key when the same entry is written concurrently', async () => {
    const { keyStore, storage } = setup();
    const created: string[] = [];
    const originalSetItem = keyStore.setItem.bind(keyStore);
    keyStore.setItem = async (key, value) => {
      created.push(value);
      await originalSetItem(key, value);
    };

    await Promise.all([storage.setItem('session', 'first'), storage.setItem('session', 'second')]);

    expect(created).toHaveLength(1);
    await expect(storage.getItem('session')).resolves.not.toBeNull();
  });

  it('replaces a corrupt stored key instead of failing every write', async () => {
    const { keyStore, storage } = setup();
    keyStore.entries.set('session', 'not-a-valid-key');

    await storage.setItem('session', 'value');

    expect(keyStore.entries.get('session')).toHaveLength(64);
    await expect(storage.getItem('session')).resolves.toBe('value');
  });

  it('keeps independent keys for independent entries', async () => {
    const { keyStore, storage } = setup();

    await storage.setItem('one', 'first');
    await storage.setItem('two', 'second');

    expect(keyStore.entries.get('one')).not.toBe(keyStore.entries.get('two'));
    await expect(storage.getItem('one')).resolves.toBe('first');
    await expect(storage.getItem('two')).resolves.toBe('second');
  });

  it('treats tampered data as missing', async () => {
    const { dataStore, storage } = setup();
    await storage.setItem('session', 'value');
    const stored = dataStore.entries.get('session') ?? '';
    const flipped = (stored.startsWith('0') ? '1' : '0') + stored.slice(1);
    dataStore.entries.set('session', flipped);

    await expect(storage.getItem('session')).resolves.toBeNull();
  });

  it('treats data as missing when its key is gone, for example after a reinstall', async () => {
    const { keyStore, storage } = setup();
    await storage.setItem('session', 'value');
    keyStore.entries.clear();

    await expect(storage.getItem('session')).resolves.toBeNull();
  });

  it('removes both the data and its key', async () => {
    const { keyStore, dataStore, storage } = setup();
    await storage.setItem('session', 'value');

    await storage.removeItem('session');

    expect(dataStore.entries.size).toBe(0);
    expect(keyStore.entries.size).toBe(0);
    await expect(storage.getItem('session')).resolves.toBeNull();
  });
});
