import { useCallback, useEffect, useState, type ReactNode } from 'react';
import { Pressable, StyleSheet, Text, View } from 'react-native';

import { t } from '../../../shared/i18n';

type Status = 'starting' | 'ready' | 'failed';

interface Props {
  ensureSession: () => Promise<unknown>;
  challenge?: ReactNode;
  children: ReactNode;
}

export function SessionGate({ ensureSession, challenge, children }: Props) {
  const [status, setStatus] = useState<Status>('starting');
  const [attempt, setAttempt] = useState(0);

  useEffect(() => {
    let active = true;
    ensureSession().then(
      () => active && setStatus('ready'),
      () => active && setStatus('failed'),
    );
    return () => {
      active = false;
    };
  }, [ensureSession, attempt]);

  const retry = useCallback(() => {
    setStatus('starting');
    setAttempt((current) => current + 1);
  }, []);

  return (
    <>
      {status === 'ready' ? (
        children
      ) : (
        <View style={styles.container}>
          {status === 'starting' ? (
            <Text accessibilityRole="progressbar">{t('identity.starting')}</Text>
          ) : (
            <>
              <Text accessibilityRole="alert" style={styles.message}>
                {t('identity.error')}
              </Text>
              <Pressable
                accessibilityRole="button"
                accessibilityLabel={t('identity.retry')}
                onPress={retry}
                style={styles.button}
              >
                <Text style={styles.buttonLabel}>{t('identity.retry')}</Text>
              </Pressable>
            </>
          )}
        </View>
      )}
      {challenge}
    </>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, alignItems: 'center', justifyContent: 'center', padding: 24 },
  message: { fontSize: 16, textAlign: 'center' },
  button: {
    marginTop: 16,
    minHeight: 48,
    minWidth: 160,
    alignItems: 'center',
    justifyContent: 'center',
    borderRadius: 8,
    backgroundColor: '#111',
  },
  buttonLabel: { color: '#fff', fontSize: 16, fontWeight: '600' },
});
