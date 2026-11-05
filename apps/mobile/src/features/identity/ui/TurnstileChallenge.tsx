import { useEffect, useMemo, useState } from 'react';
import { Modal, StyleSheet, Text, View } from 'react-native';
import { WebView, type WebViewMessageEvent } from 'react-native-webview';

import { t } from '../../../shared/i18n';
import type { CaptchaBridge } from './captchaBridge';
import { isChallengeNavigationAllowed } from './isChallengeNavigationAllowed';
import { parseTurnstileMessage } from './parseTurnstileMessage';
import { turnstileHtml } from './turnstileHtml';

interface Props {
  bridge: CaptchaBridge;
  siteKey: string;
  baseUrl: string;
}

export function TurnstileChallenge({ bridge, siteKey, baseUrl }: Props) {
  const [visible, setVisible] = useState(false);
  const html = useMemo(() => turnstileHtml(siteKey), [siteKey]);

  useEffect(() => bridge.subscribe(setVisible), [bridge]);

  const onMessage = (event: WebViewMessageEvent) => {
    const message = parseTurnstileMessage(event.nativeEvent.data);
    if (message?.kind === 'token') bridge.complete(message.token);
    if (message?.kind === 'error') bridge.fail(new Error(`turnstile error ${message.code}`));
  };

  return (
    <Modal
      testID="turnstile-modal"
      visible={visible}
      transparent
      animationType="fade"
      onRequestClose={() => bridge.fail(new Error('turnstile dismissed'))}
    >
      <View style={styles.backdrop}>
        <View style={styles.card}>
          <Text accessibilityRole="header" style={styles.title}>
            {t('identity.captcha.title')}
          </Text>
          <WebView
            testID="turnstile-webview"
            originWhitelist={['*']}
            source={{ html, baseUrl }}
            onMessage={onMessage}
            onShouldStartLoadWithRequest={(request) =>
              isChallengeNavigationAllowed(request, baseUrl)
            }
            style={styles.webview}
          />
        </View>
      </View>
    </Modal>
  );
}

const styles = StyleSheet.create({
  backdrop: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    padding: 24,
    backgroundColor: 'rgba(0,0,0,0.5)',
  },
  card: { width: '100%', maxWidth: 360, padding: 16, borderRadius: 12, backgroundColor: '#fff' },
  title: { marginBottom: 12, fontSize: 18, fontWeight: '600' },
  webview: { height: 90, backgroundColor: 'transparent' },
});
