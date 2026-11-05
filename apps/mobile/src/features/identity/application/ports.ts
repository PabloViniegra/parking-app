export interface Session {
  accessToken: string;
}

export interface AnonymousAuth {
  currentSession(): Promise<Session | null>;
  signInAnonymously(captchaToken: string): Promise<Session>;
}

export interface CaptchaChallenge {
  solve(): Promise<string>;
}
