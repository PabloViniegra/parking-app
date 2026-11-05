import { isChallengeNavigationAllowed } from './isChallengeNavigationAllowed';

const BASE_URL = 'https://parking.example.test';

describe('isChallengeNavigationAllowed', () => {
  it.each([
    {
      name: 'allows the initial load of the challenge page',
      request: { url: 'https://parking.example.test/' },
    },
    {
      name: 'allows the initial load without a trailing slash',
      request: { url: 'https://parking.example.test' },
    },
    { name: 'allows a blank page', request: { url: 'about:blank' } },
    {
      name: 'allows the Cloudflare challenge frame',
      request: { url: 'https://challenges.cloudflare.com/cdn-cgi/challenge', isTopFrame: false },
    },
  ])('$name', ({ request }) => {
    expect(isChallengeNavigationAllowed(request, BASE_URL)).toBe(true);
  });

  it.each([
    {
      name: 'blocks navigating the page to another site',
      request: { url: 'https://evil.example.test/phish', isTopFrame: true },
    },
    {
      name: 'blocks a look-alike host that merely starts with the base URL',
      request: { url: 'https://parking.example.test.evil.test/' },
    },
    {
      name: 'blocks navigating the page to Cloudflare',
      request: { url: 'https://challenges.cloudflare.com/', isTopFrame: true },
    },
    {
      name: 'blocks non-web schemes on the page',
      request: { url: 'intent://scan#Intent;end' },
    },
  ])('$name', ({ request }) => {
    expect(isChallengeNavigationAllowed(request, BASE_URL)).toBe(false);
  });

  it('accepts a base URL written with a trailing slash', () => {
    expect(
      isChallengeNavigationAllowed({ url: 'https://parking.example.test/' }, `${BASE_URL}/`),
    ).toBe(true);
  });
});
