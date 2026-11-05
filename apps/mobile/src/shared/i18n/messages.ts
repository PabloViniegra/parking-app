const esES = {
  'app.name': 'parking-app',
  'app.tagline': 'Encuentra aparcamiento libre gracias a otros conductores.',
  'identity.starting': 'Preparando tu sesión…',
  'identity.error': 'No hemos podido iniciar tu sesión.',
  'identity.retry': 'Reintentar',
  'identity.captcha.title': 'Verificación de seguridad',
};

export type MessageKey = keyof typeof esES;

export const messages = {
  'es-ES': esES,
  'en-GB': {
    'app.name': 'parking-app',
    'app.tagline': 'Find free street parking thanks to other drivers.',
    'identity.starting': 'Setting up your session…',
    'identity.error': 'We could not start your session.',
    'identity.retry': 'Try again',
    'identity.captcha.title': 'Security check',
  },
} satisfies Record<string, Record<MessageKey, string>>;
