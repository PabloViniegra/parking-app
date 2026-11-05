export interface AutoRefreshControl {
  startAutoRefresh(): Promise<void>;
  stopAutoRefresh(): Promise<void>;
}

export interface AppStateSource {
  currentState: string | null | undefined;
  addEventListener(type: 'change', listener: (state: string) => void): { remove(): void };
}

export function bindAutoRefresh(auth: AutoRefreshControl, appState: AppStateSource): () => void {
  const apply = (state: string | null | undefined) => {
    void (state === 'active' ? auth.startAutoRefresh() : auth.stopAutoRefresh());
  };

  apply(appState.currentState);
  const subscription = appState.addEventListener('change', apply);
  return () => subscription.remove();
}
