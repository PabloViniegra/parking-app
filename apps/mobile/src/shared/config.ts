function required(name: string, value: string | undefined): string {
  if (!value) throw new Error(`Missing environment variable ${name}`);
  return value;
}

export const config = {
  supabaseUrl: required('EXPO_PUBLIC_SUPABASE_URL', process.env.EXPO_PUBLIC_SUPABASE_URL),
  supabasePublishableKey: required(
    'EXPO_PUBLIC_SUPABASE_PUBLISHABLE_KEY',
    process.env.EXPO_PUBLIC_SUPABASE_PUBLISHABLE_KEY,
  ),
  turnstileSiteKey: required(
    'EXPO_PUBLIC_TURNSTILE_SITE_KEY',
    process.env.EXPO_PUBLIC_TURNSTILE_SITE_KEY,
  ),
  turnstileBaseUrl: required(
    'EXPO_PUBLIC_TURNSTILE_BASE_URL',
    process.env.EXPO_PUBLIC_TURNSTILE_BASE_URL,
  ),
};
