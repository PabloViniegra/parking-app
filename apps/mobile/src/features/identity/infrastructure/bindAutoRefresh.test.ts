import { bindAutoRefresh, type AppStateSource, type AutoRefreshControl } from './bindAutoRefresh';

class FakeAppState implements AppStateSource {
  listener: ((state: string) => void) | null = null;
  removed = 0;

  constructor(public currentState: string) {}

  addEventListener(_type: 'change', listener: (state: string) => void) {
    this.listener = listener;
    return { remove: () => void (this.removed += 1) };
  }

  change(state: string) {
    this.currentState = state;
    this.listener?.(state);
  }
}

function fakeAuth() {
  const calls: string[] = [];
  const auth: AutoRefreshControl = {
    startAutoRefresh: async () => void calls.push('start'),
    stopAutoRefresh: async () => void calls.push('stop'),
  };
  return { auth, calls };
}

describe('bindAutoRefresh', () => {
  it('refreshes the session only while the app is in the foreground', () => {
    const { auth, calls } = fakeAuth();
    const appState = new FakeAppState('active');

    bindAutoRefresh(auth, appState);
    appState.change('background');
    appState.change('active');

    expect(calls).toEqual(['start', 'stop', 'start']);
  });

  it('does not refresh when the app starts in the background', () => {
    const { auth, calls } = fakeAuth();

    bindAutoRefresh(auth, new FakeAppState('background'));

    expect(calls).toEqual(['stop']);
  });

  it('stops listening once disposed', () => {
    const { auth } = fakeAuth();
    const appState = new FakeAppState('active');

    const dispose = bindAutoRefresh(auth, appState);
    dispose();

    expect(appState.removed).toBe(1);
  });
});
