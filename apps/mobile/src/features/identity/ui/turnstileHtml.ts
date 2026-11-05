const SCRIPT_URL = 'https://challenges.cloudflare.com/turnstile/v0/api.js';

function toScriptLiteral(value: string): string {
  return JSON.stringify(value).replace(/</g, '\\u003c');
}

export function turnstileHtml(siteKey: string): string {
  return `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<style>
  html, body { margin: 0; height: 100%; display: flex; align-items: center; justify-content: center; background: transparent; }
</style>
<script>
  function post(message) {
    window.ReactNativeWebView.postMessage(JSON.stringify(message));
  }
  function onTurnstileLoad() {
    turnstile.render('#widget', {
      sitekey: ${toScriptLiteral(siteKey)},
      callback: function (token) { post({ type: 'token', token: token }); },
      'error-callback': function (code) { post({ type: 'error', code: String(code) }); }
    });
  }
</script>
<script src="${SCRIPT_URL}?render=explicit&onload=onTurnstileLoad" async defer></script>
</head>
<body>
<div id="widget"></div>
</body>
</html>`;
}
