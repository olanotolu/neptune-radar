// Instagram / Meta CDNs send Cross-Origin-Resource-Policy: same-origin, which
// makes browsers refuse to paint <img> from our origin. Route those URLs
// through our same-origin proxy.

const CDN_HINTS = ["cdninstagram.com", "fbcdn.net", "instagram.com", "scontent"];

function needsProxy(url: string): boolean {
  try {
    const u = new URL(url);
    const host = u.hostname.toLowerCase();
    return CDN_HINTS.some((h) => host === h || host.endsWith(`.${h}`) || host.includes(h));
  } catch {
    return false;
  }
}

/** Rewrite a remote media URL for safe display in the dashboard. */
export function mediaURL(url?: string | null): string {
  if (!url) return "";
  if (!needsProxy(url)) return url;
  // Relative so production (same origin) and dev (VITE_API_URL) both work.
  const base = (import.meta.env.VITE_API_URL as string | undefined) ?? "";
  return `${base}/api/media?url=${encodeURIComponent(url)}`;
}
