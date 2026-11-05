import { turnstileHtml } from './turnstileHtml';

describe('turnstileHtml', () => {
  it('loads the Cloudflare Turnstile script and renders the widget with the site key', () => {
    const html = turnstileHtml('0x4AAAAAAA-site-key');

    expect(html).toContain('https://challenges.cloudflare.com/turnstile/v0/api.js');
    expect(html).toContain('"0x4AAAAAAA-site-key"');
  });

  it('reports token and error to the app through the WebView bridge', () => {
    const html = turnstileHtml('key');

    expect(html).toContain("type: 'token'");
    expect(html).toContain("type: 'error'");
    expect(html).toContain('window.ReactNativeWebView.postMessage');
  });

  it('cannot be broken out of the script block by a hostile site key', () => {
    const html = turnstileHtml('</script><script>alert(1)</script>');

    expect(html).not.toContain('</script><script>alert(1)');
  });
});
