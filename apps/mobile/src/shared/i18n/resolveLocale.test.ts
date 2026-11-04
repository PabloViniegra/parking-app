import { resolveLocale } from './resolveLocale';

describe('resolveLocale', () => {
  it.each([
    {
      name: 'falls back to es-ES when the device reports no locale',
      device: [],
      expected: 'es-ES',
    },
    { name: 'falls back to es-ES for unsupported languages', device: ['fr-FR'], expected: 'es-ES' },
    { name: 'keeps a supported device locale', device: ['en-GB'], expected: 'en-GB' },
    {
      name: 'matches a supported language with another region',
      device: ['en-US'],
      expected: 'en-GB',
    },
    {
      name: 'uses the first supported preference',
      device: ['fr-FR', 'en-GB', 'es-ES'],
      expected: 'en-GB',
    },
  ])('$name', ({ device, expected }) => {
    expect(resolveLocale(device, ['es-ES', 'en-GB'])).toBe(expected);
  });
});
