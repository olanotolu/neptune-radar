import { useEffect, useMemo, useRef, useState } from "react";
import { useLiveEvents } from "../api/hooks";
import type { LiveEvent, LiveEventType } from "../api/types";

const TONE: Record<LiveEventType, { color: string; icon: string; label: string }> = {
  couple_detected: { color: "green", icon: "♥", label: "Couple detected" },
  stage_transition: { color: "blue", icon: "→", label: "Stage transition" },
  action_created: { color: "yellow", icon: "!", label: "Action created" },
  alert: { color: "red", icon: "⚠", label: "Alert" },
};

function timeAgo(iso: string): string {
  const then = new Date(iso).getTime();
  if (Number.isNaN(then)) return "";
  const s = Math.max(0, Math.round((Date.now() - then) / 1000));
  if (s < 60) return `${s}s ago`;
  const m = Math.floor(s / 60);
  if (m < 60) return `${m}m ago`;
  const h = Math.floor(m / 60);
  if (h < 24) return `${h}h ago`;
  const d = Math.floor(h / 24);
  return `${d}d ago`;
}

function summary(evt: LiveEvent): string {
  const d = evt.data;
  switch (evt.type) {
    case "couple_detected":
      return `${d.person_a ?? d.handle_a ?? "pair"} & ${d.person_b ?? d.handle_b ?? "?"}`;
    case "stage_transition":
      return `${d.couple_id ?? "couple"}: ${d.from ?? "?"} → ${d.to ?? "?"}`;
    case "action_created":
      return `${d.action_type ?? "action"} for ${d.couple_id ?? d.hypothesis_id ?? "couple"}`;
    case "alert":
      return String(d.message ?? d.summary ?? "alert");
    default:
      return "";
  }
}

function routeFor(evt: LiveEvent): string | null {
  const d = evt.data;
  if (evt.type === "couple_detected" || evt.type === "stage_transition") {
    const id = d.couple_id as string | undefined;
    return id ? `/work?couple=${encodeURIComponent(id)}` : "/work?filter=action";
  }
  if (evt.type === "action_created") {
    const id = d.couple_id as string | undefined;
    return id ? `/work?couple=${encodeURIComponent(id)}` : "/work?filter=action";
  }
  return null;
}

export function NotificationCenter({ onNavigate }: { onNavigate: (path: string) => void }) {
  const { events } = useLiveEvents();
  const [open, setOpen] = useState(false);
  const [lastSeen, setLastSeen] = useState<number>(() => Date.now());
  const [now, setNow] = useState(Date.now());
  const ref = useRef<HTMLDivElement | null>(null);

  // Tick relative timestamps every 15s while open.
  useEffect(() => {
    if (!open) return;
    const id = window.setInterval(() => setNow(Date.now()), 15_000);
    return () => window.clearInterval(id);
  }, [open]);

  // Close on outside click.
  useEffect(() => {
    if (!open) return;
    const onDown = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false);
    };
    document.addEventListener("mousedown", onDown);
    return () => document.removeEventListener("mousedown", onDown);
  }, [open]);

  const recent = useMemo(() => [...events].reverse().slice(0, 50), [events]);
  const unread = useMemo(
    () => events.filter((e) => new Date(e.time).getTime() > lastSeen && e.type === "alert").length,
    [events, lastSeen],
  );

  const markRead = () => setLastSeen(Date.now());

  const onClickEvent = (evt: LiveEvent) => {
    const path = routeFor(evt);
    setOpen(false);
    if (path) onNavigate(path);
  };

  return (
    <div className="notif" ref={ref}>
      <button
        type="button"
        className="notif__bell"
        aria-label={`Notifications${unread ? ` (${unread} unread)` : ""}`}
        onClick={() => {
          setOpen((v) => !v);
          if (!open) markRead();
        }}
      >
        <span aria-hidden>🔔</span>
        {unread > 0 && <span className="notif__badge">{unread > 99 ? "99+" : unread}</span>}
      </button>
      {open && (
        <div className="notif__panel" role="menu">
          <header className="notif__panel-header">
            <span className="notif__panel-title">Recent events</span>
            <button type="button" className="btn btn--ghost btn--sm" onClick={markRead}>
              Mark all read
            </button>
          </header>
          <div className="notif__list">
            {recent.length === 0 && <div className="notif__empty">No events yet</div>}
            {recent.map((evt, i) => {
              const tone = TONE[evt.type];
              return (
                <button
                  key={`${evt.time}-${i}`}
                  type="button"
                  className={`notif__item notif__item--${tone.color}`}
                  onClick={() => onClickEvent(evt)}
                >
                  <span className="notif__item-icon" aria-hidden>{tone.icon}</span>
                  <span className="notif__item-body">
                    <span className="notif__item-type">{tone.label}</span>
                    <span className="notif__item-summary">{summary(evt)}</span>
                  </span>
                  <span className="notif__item-time">{timeAgo(evt.time)}</span>
                </button>
              );
            })}
          </div>
        </div>
      )}
      {/* now is referenced so relative timestamps re-render on tick */}
      <span hidden>{now}</span>
    </div>
  );
}
