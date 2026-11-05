import { act, render, screen, userEvent } from '@testing-library/react-native';
import { Text } from 'react-native';

import { t } from '../../../shared/i18n';
import { SessionGate } from './SessionGate';

function deferred() {
  let resolve!: () => void;
  let reject!: (error: Error) => void;
  const promise = new Promise<void>((res, rej) => {
    resolve = res;
    reject = rej;
  });
  return { promise, resolve, reject };
}

describe('SessionGate', () => {
  it('shows progress while the session starts and the app once it is ready', async () => {
    const starting = deferred();
    await render(
      <SessionGate ensureSession={() => starting.promise}>
        <Text>app content</Text>
      </SessionGate>,
    );

    expect(screen.getByText(t('identity.starting'))).toBeTruthy();
    expect(screen.queryByText('app content')).toBeNull();

    starting.resolve();

    expect(await screen.findByText('app content')).toBeTruthy();
    expect(screen.queryByText(t('identity.starting'))).toBeNull();
  });

  it('offers a retry when the session cannot start and recovers on success', async () => {
    const attempts = [deferred(), deferred()];
    let call = 0;
    const ensureSession = jest.fn(() => attempts[call++]!.promise);
    await render(
      <SessionGate ensureSession={ensureSession}>
        <Text>app content</Text>
      </SessionGate>,
    );

    attempts[0]!.reject(new Error('captcha failed'));
    expect(await screen.findByText(t('identity.error'))).toBeTruthy();

    await userEvent.press(screen.getByRole('button', { name: t('identity.retry') }));
    attempts[1]!.resolve();

    expect(await screen.findByText('app content')).toBeTruthy();
    expect(ensureSession).toHaveBeenCalledTimes(2);
  });

  it('ignores the result of a superseded session start', async () => {
    const stale = deferred();
    const current = deferred();
    const view = (ensureSession: () => Promise<void>) => (
      <SessionGate ensureSession={ensureSession}>
        <Text>app content</Text>
      </SessionGate>
    );
    const { rerender } = await render(view(() => stale.promise));

    await rerender(view(() => current.promise));
    stale.resolve();
    await act(async () => {});

    expect(screen.queryByText('app content')).toBeNull();

    current.resolve();

    expect(await screen.findByText('app content')).toBeTruthy();
  });

  it('renders the challenge element alongside the progress state', async () => {
    const starting = deferred();
    await render(
      <SessionGate ensureSession={() => starting.promise} challenge={<Text>captcha widget</Text>}>
        <Text>app content</Text>
      </SessionGate>,
    );

    expect(screen.getByText('captcha widget')).toBeTruthy();
  });
});
