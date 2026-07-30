// Empty fallback = same origin. Production serves the built dashboard and
// the API from one container, so relative paths are correct there; local dev
// sets VITE_API_URL=http://localhost:8080 in frontend/.env (not deployed).
const BASE_URL = import.meta.env.VITE_API_URL ?? "";

const TOKEN_KEY = "neptune_admin_token";

export function getToken(): string {
	// localStorage only. There used to be a VITE_NEPTUNE_TOKEN env fallback
	// here — any VITE_* var is statically inlined into the shipped bundle,
	// so setting it once would have baked the admin bearer token into every
	// visitor's browser.
	return localStorage.getItem(TOKEN_KEY) ?? "";
}

export function setToken(token: string) {
  localStorage.setItem(TOKEN_KEY, token);
}

export function clearToken() {
  localStorage.removeItem(TOKEN_KEY);
}

// App registers this once; any 401 from the API means the stored token is
// wrong or stale, so we drop it and bounce back to the token gate instead
// of dead-ending every view on "missing or invalid bearer token".
let onUnauthorized: (() => void) | null = null;
export function setUnauthorizedHandler(fn: (() => void) | null) {
  onUnauthorized = fn;
}

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${getToken()}`,
      ...(init?.headers ?? {}),
    },
  });
  if (res.status === 401) {
    clearToken();
    onUnauthorized?.();
    throw new Error("unauthorized");
  }
  if (!res.ok) {
    const body = await res.text();
    throw new Error(`${init?.method ?? "GET"} ${path} failed: ${res.status} ${body}`);
  }
  if (res.status === 204) return undefined as T;
  return res.json() as Promise<T>;
}

export const api = {
  get: <T>(path: string) => request<T>(path),
  post: <T>(path: string, body?: unknown) =>
    request<T>(path, { method: "POST", body: body ? JSON.stringify(body) : undefined }),
  patch: <T>(path: string, body?: unknown) =>
    request<T>(path, { method: "PATCH", body: body ? JSON.stringify(body) : undefined }),
  del: <T>(path: string) => request<T>(path, { method: "DELETE" }),
};
