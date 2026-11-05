import { StatusBar } from 'expo-status-bar';
import { StyleSheet, Text, View } from 'react-native';

import { ensureAnonymousSession } from './features/identity/application/ensureAnonymousSession';
import { singleFlight } from './features/identity/application/singleFlight';
import { createCaptchaBridge } from './features/identity/ui/captchaBridge';
import { deviceSessionStorage } from './features/identity/infrastructure/deviceSessionStorage';
import { createSupabaseAnonymousAuth } from './features/identity/infrastructure/supabaseAnonymousAuth';
import { TurnstileChallenge } from './features/identity/ui/TurnstileChallenge';
import { SessionGate } from './features/identity/ui/SessionGate';
import { config } from './shared/config';
import { t } from './shared/i18n';

const captchaBridge = createCaptchaBridge();
const auth = createSupabaseAnonymousAuth({
  url: config.supabaseUrl,
  publishableKey: config.supabasePublishableKey,
  storage: deviceSessionStorage,
});
const ensureSession = singleFlight(() =>
  ensureAnonymousSession({ auth, captcha: captchaBridge.challenge }),
);

export default function App() {
  return (
    <SessionGate
      ensureSession={ensureSession}
      challenge={
        <TurnstileChallenge
          bridge={captchaBridge}
          siteKey={config.turnstileSiteKey}
          baseUrl={config.turnstileBaseUrl}
        />
      }
    >
      <View style={styles.container}>
        <Text accessibilityRole="header" style={styles.title}>
          {t('app.name')}
        </Text>
        <Text style={styles.tagline}>{t('app.tagline')}</Text>
        <StatusBar style="auto" />
      </View>
    </SessionGate>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    padding: 24,
    backgroundColor: '#fff',
  },
  title: { fontSize: 28, fontWeight: '700' },
  tagline: { marginTop: 8, fontSize: 16, textAlign: 'center' },
});
