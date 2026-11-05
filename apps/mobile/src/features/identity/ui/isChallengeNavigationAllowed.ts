interface NavigationRequest {
  url: string;
  isTopFrame?: boolean;
}

// iOS reports sub-frame loads with isTopFrame=false; Android only reports top-level navigations.
export function isChallengeNavigationAllowed(request: NavigationRequest, baseUrl: string): boolean {
  if (request.isTopFrame === false) return true;
  if (request.url === 'about:blank') return true;

  const base = baseUrl.replace(/\/+$/, '');
  return request.url === base || request.url.startsWith(`${base}/`);
}
