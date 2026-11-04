export const DEFAULT_LOCALE = 'es-ES';

const language = (tag: string) => tag.split('-')[0]?.toLowerCase();

export function resolveLocale(
  deviceLocales: readonly string[],
  supported: readonly string[],
): string {
  for (const tag of deviceLocales) {
    const match =
      supported.find((s) => s.toLowerCase() === tag.toLowerCase()) ??
      supported.find((s) => language(s) === language(tag));
    if (match) return match;
  }
  return DEFAULT_LOCALE;
}
