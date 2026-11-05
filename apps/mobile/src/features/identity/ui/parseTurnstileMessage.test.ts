import { parseTurnstileMessage } from './parseTurnstileMessage';

describe('parseTurnstileMessage', () => {
  it.each([
    {
      name: 'reads a solved challenge',
      data: JSON.stringify({ type: 'token', token: 'abc' }),
      expected: { kind: 'token', token: 'abc' },
    },
    {
      name: 'reads a widget error with its code',
      data: JSON.stringify({ type: 'error', code: '110200' }),
      expected: { kind: 'error', code: '110200' },
    },
    { name: 'ignores text that is not JSON', data: 'hello', expected: null },
    {
      name: 'ignores unknown message types',
      data: JSON.stringify({ type: 'other' }),
      expected: null,
    },
    {
      name: 'ignores a token message without a token',
      data: JSON.stringify({ type: 'token' }),
      expected: null,
    },
    {
      name: 'ignores an empty token',
      data: JSON.stringify({ type: 'token', token: '' }),
      expected: null,
    },
  ])('$name', ({ data, expected }) => {
    expect(parseTurnstileMessage(data)).toEqual(expected);
  });
});
