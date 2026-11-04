const esES = {
  'app.name': 'parking-app',
  'app.tagline': 'Encuentra aparcamiento libre gracias a otros conductores.',
};

export type MessageKey = keyof typeof esES;

export const messages = {
  'es-ES': esES,
  'en-GB': {
    'app.name': 'parking-app',
    'app.tagline': 'Find free street parking thanks to other drivers.',
  },
} satisfies Record<string, Record<MessageKey, string>>;
