export type TurnstileMessage =
  | { kind: 'token'; token: string }
  | { kind: 'error'; code: string };

export function parseTurnstileMessage(data: string): TurnstileMessage | null {
  let payload: unknown;
  try {
    payload = JSON.parse(data);
  } catch {
    return null;
  }
  if (typeof payload !== 'object' || payload === null) return null;

  const { type, token, code } = payload as Record<string, unknown>;
  if (type === 'token' && typeof token === 'string' && token !== '') {
    return { kind: 'token', token };
  }
  if (type === 'error') {
    return { kind: 'error', code: typeof code === 'string' ? code : 'unknown' };
  }
  return null;
}
