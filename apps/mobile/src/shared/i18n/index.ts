import { getLocales } from 'expo-localization';

import { messages, type MessageKey } from './messages';
import { resolveLocale } from './resolveLocale';

const catalogue: Record<string, Record<MessageKey, string>> = messages;

const locale = resolveLocale(
  getLocales().map((l) => l.languageTag),
  Object.keys(messages),
);

export function t(key: MessageKey): string {
  return catalogue[locale]?.[key] ?? messages['es-ES'][key];
}
