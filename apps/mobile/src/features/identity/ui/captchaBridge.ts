import type { CaptchaChallenge } from '../application/ports';

type VisibilityListener = (visible: boolean) => void;

interface Pending {
  promise: Promise<string>;
  resolve: (token: string) => void;
  reject: (error: Error) => void;
}

export interface CaptchaBridge {
  challenge: CaptchaChallenge;
  subscribe(listener: VisibilityListener): () => void;
  complete(token: string): void;
  fail(error: Error): void;
}

export function createCaptchaBridge(): CaptchaBridge {
  const listeners = new Set<VisibilityListener>();
  let pending: Pending | null = null;

  const notify = (visible: boolean) => listeners.forEach((listener) => listener(visible));

  const settle = (action: (current: Pending) => void) => {
    if (!pending) return;
    const current = pending;
    pending = null;
    action(current);
    notify(false);
  };

  return {
    challenge: {
      solve() {
        if (pending) return pending.promise;
        let resolve!: Pending['resolve'];
        let reject!: Pending['reject'];
        const promise = new Promise<string>((res, rej) => {
          resolve = res;
          reject = rej;
        });
        pending = { promise, resolve, reject };
        notify(true);
        return promise;
      },
    },
    subscribe(listener) {
      listeners.add(listener);
      return () => {
        listeners.delete(listener);
      };
    },
    complete: (token) => settle((current) => current.resolve(token)),
    fail: (error) => settle((current) => current.reject(error)),
  };
}
