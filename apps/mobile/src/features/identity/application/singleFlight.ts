export function singleFlight<T>(task: () => Promise<T>): () => Promise<T> {
  let inFlight: Promise<T> | null = null;

  return () => {
    if (inFlight) return inFlight;
    const current = task().finally(() => {
      inFlight = null;
    });
    inFlight = current;
    return current;
  };
}
