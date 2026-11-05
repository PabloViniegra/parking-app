import { act, fireEvent, render, screen } from '@testing-library/react-native';

import { t } from '../../../shared/i18n';
import { createCaptchaBridge } from './captchaBridge';
import { TurnstileChallenge } from './TurnstileChallenge';

jest.mock('react-native-webview', () => {
  const { View } = jest.requireActual<typeof import('react-native')>('react-native');
  return { WebView: View };
});

function setup() {
  const bridge = createCaptchaBridge();
  return {
    bridge,
    element: <TurnstileChallenge bridge={bridge} siteKey="site-key" baseUrl="https://localhost" />,
  };
}

function message(payload: object) {
  return { nativeEvent: { data: JSON.stringify(payload) } };
}

describe('TurnstileChallenge', () => {
  it('stays hidden until a CAPTCHA is requested', async () => {
    const { element } = setup();
    await render(element);

    expect(screen.queryByText(t('identity.captcha.title'))).toBeNull();
  });

  it('shows the security check while a CAPTCHA is being solved', async () => {
    const { bridge, element } = setup();
    await render(element);

    await act(async () => {
      void bridge.challenge.solve();
    });

    expect(screen.getByRole('header', { name: t('identity.captcha.title') })).toBeTruthy();
  });

  it('delivers the token to the pending solve and closes', async () => {
    const { bridge, element } = setup();
    await render(element);
    let solving!: Promise<string>;
    await act(async () => {
      solving = bridge.challenge.solve();
    });

    await fireEvent(
      screen.getByTestId('turnstile-webview'),
      'message',
      message({ type: 'token', token: 'turnstile-token' }),
    );

    await expect(solving).resolves.toBe('turnstile-token');
    expect(screen.queryByText(t('identity.captcha.title'))).toBeNull();
  });

  it('rejects the pending solve when the widget reports an error', async () => {
    const { bridge, element } = setup();
    await render(element);
    let solving!: Promise<string>;
    await act(async () => {
      solving = bridge.challenge.solve();
    });

    const rejection = expect(solving).rejects.toThrow('turnstile error 110200');
    await fireEvent(
      screen.getByTestId('turnstile-webview'),
      'message',
      message({ type: 'error', code: '110200' }),
    );

    await rejection;
  });

  it('rejects the pending solve when the user dismisses the dialog', async () => {
    const { bridge, element } = setup();
    await render(element);
    let solving!: Promise<string>;
    await act(async () => {
      solving = bridge.challenge.solve();
    });

    const rejection = expect(solving).rejects.toThrow('turnstile dismissed');
    await fireEvent(screen.getByTestId('turnstile-modal'), 'requestClose');

    await rejection;
    expect(screen.queryByText(t('identity.captcha.title'))).toBeNull();
  });

  it('keeps the WebView on the challenge page and refuses to navigate elsewhere', async () => {
    const { bridge, element } = setup();
    await render(element);
    await act(async () => {
      void bridge.challenge.solve();
    });
    const { onShouldStartLoadWithRequest } = screen.getByTestId('turnstile-webview').props;

    expect(onShouldStartLoadWithRequest({ url: 'https://localhost/' })).toBe(true);
    expect(onShouldStartLoadWithRequest({ url: 'https://evil.example.test/' })).toBe(false);
  });

  it('keeps waiting when the widget sends unrelated messages', async () => {
    const { bridge, element } = setup();
    await render(element);
    let settled = false;
    await act(async () => {
      bridge.challenge.solve().then(
        () => (settled = true),
        () => (settled = true),
      );
    });

    await fireEvent(screen.getByTestId('turnstile-webview'), 'message', message({ type: 'other' }));
    await fireEvent(screen.getByTestId('turnstile-webview'), 'message', {
      nativeEvent: { data: 'not json' },
    });

    expect(settled).toBe(false);
    expect(screen.getByTestId('turnstile-webview')).toBeTruthy();
  });
});
