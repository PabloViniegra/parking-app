import { singleFlight } from './singleFlight';

function deferred<T>() {
  let resolve!: (value: T) => void;
  let reject!: (error: Error) => void;
  const promise = new Promise<T>((res, rej) => {
    resolve = res;
    reject = rej;
  });
  return { promise, resolve, reject };
}

describe('singleFlight', () => {
  it('runs the task once for concurrent calls and shares its result', async () => {
    const pending = deferred<string>();
    const task = jest.fn(() => pending.promise);
    const run = singleFlight(task);

    const first = run();
    const second = run();
    pending.resolve('session');

    await expect(Promise.all([first, second])).resolves.toEqual(['session', 'session']);
    expect(task).toHaveBeenCalledTimes(1);
  });

  it('runs the task again once the previous call has succeeded', async () => {
    const task = jest.fn(async () => 'session');
    const run = singleFlight(task);

    await run();
    await run();

    expect(task).toHaveBeenCalledTimes(2);
  });

  it('runs the task again after a failure so a retry is possible', async () => {
    const task = jest
      .fn<Promise<string>, []>()
      .mockRejectedValueOnce(new Error('captcha failed'))
      .mockResolvedValueOnce('session');
    const run = singleFlight(task);

    await expect(run()).rejects.toThrow('captcha failed');

    await expect(run()).resolves.toBe('session');
    expect(task).toHaveBeenCalledTimes(2);
  });
});
