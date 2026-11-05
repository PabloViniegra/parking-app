import { createCaptchaBridge } from './captchaBridge';

describe('captchaBridge', () => {
  it('shows the challenge while solving and resolves with the token once completed', async () => {
    const bridge = createCaptchaBridge();
    const visibility: boolean[] = [];
    bridge.subscribe((visible) => visibility.push(visible));

    const solving = bridge.challenge.solve();
    expect(visibility).toEqual([true]);

    bridge.complete('turnstile-token');

    await expect(solving).resolves.toBe('turnstile-token');
    expect(visibility).toEqual([true, false]);
  });

  it('rejects the pending solve and hides the challenge when the widget fails', async () => {
    const bridge = createCaptchaBridge();
    const visibility: boolean[] = [];
    bridge.subscribe((visible) => visibility.push(visible));

    const solving = bridge.challenge.solve();
    bridge.fail(new Error('widget error 110200'));

    await expect(solving).rejects.toThrow('widget error 110200');
    expect(visibility).toEqual([true, false]);
  });

  it('asks for a new challenge when solving again after a failure', async () => {
    const bridge = createCaptchaBridge();
    const visibility: boolean[] = [];
    bridge.subscribe((visible) => visibility.push(visible));
    const failed = bridge.challenge.solve();
    bridge.fail(new Error('dismissed'));
    await expect(failed).rejects.toThrow('dismissed');

    const retry = bridge.challenge.solve();
    bridge.complete('second-token');

    await expect(retry).resolves.toBe('second-token');
    expect(visibility).toEqual([true, false, true, false]);
  });

  it('shares one challenge between concurrent solve calls', async () => {
    const bridge = createCaptchaBridge();
    const visibility: boolean[] = [];
    bridge.subscribe((visible) => visibility.push(visible));

    const first = bridge.challenge.solve();
    const second = bridge.challenge.solve();
    bridge.complete('token');

    await expect(Promise.all([first, second])).resolves.toEqual(['token', 'token']);
    expect(visibility).toEqual([true, false]);
  });

  it('ignores a token delivered when no challenge is pending', () => {
    const bridge = createCaptchaBridge();
    const visibility: boolean[] = [];
    bridge.subscribe((visible) => visibility.push(visible));

    bridge.complete('late-token');

    expect(visibility).toEqual([]);
  });

  it('stops notifying a listener after it unsubscribes', () => {
    const bridge = createCaptchaBridge();
    const visibility: boolean[] = [];
    const unsubscribe = bridge.subscribe((visible) => visibility.push(visible));

    unsubscribe();
    void bridge.challenge.solve();

    expect(visibility).toEqual([]);
  });
});
