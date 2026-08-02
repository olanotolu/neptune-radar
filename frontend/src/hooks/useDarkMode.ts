import { useCallback, useEffect, useState } from "react";

const STORAGE_KEY = "neptune-dark-mode";

function readInitial(): boolean {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored !== null) return stored === "1";
  } catch {
    /* localStorage unavailable — fall through to system preference */
  }
  return (
    typeof window !== "undefined" &&
    window.matchMedia?.("(prefers-color-scheme: dark)").matches
  );
}

/** Opt-in dark mode: defaults to system preference, persists to localStorage,
 *  and toggles `data-theme="dark"` on <html>. */
export function useDarkMode(): [boolean, () => void] {
  const [dark, setDark] = useState(readInitial);

  useEffect(() => {
    const root = document.documentElement;
    if (dark) root.setAttribute("data-theme", "dark");
    else root.removeAttribute("data-theme");
    try {
      localStorage.setItem(STORAGE_KEY, dark ? "1" : "0");
    } catch {
      /* ignore write failures */
    }
  }, [dark]);

  const toggle = useCallback(() => setDark((d) => !d), []);
  return [dark, toggle];
}
