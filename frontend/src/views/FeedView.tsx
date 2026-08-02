import { useEffect, useMemo, useRef, useState } from "react";
import { useIngestStatus, useSignals } from "../api/hooks";
import type { Signal } from "../api/types";
import { EmptyState } from "../components/EmptyState";
import { LoadingState } from "../components/LoadingState";

type TypeFilter = "all" | "engagement" | "tags" | "mentions" | "vendor";

const TYPE_CHIPS: { key: TypeFilter; label: string }[] = [
  { key: "all", label: "All" },
  { key: "engagement", label: "Engagement" },
  { key: "tags", label: "Tags" },
  { key: "mentions", label: "Mentions" },
  { key: "vendor", label: "Vendor" },
];

// ponytail: observation_type is a free-form backend string; categorize by
// substring so the chips stay useful whatever the backend emits. Ceiling: a
// future typed enum would make this exact; upgrade path = match on the enum.
function categorize(s: Signal): TypeFilter {
  const t = s.observation_type.toLowerCase();
  if (t.includes("engag") || t.includes("post")) return "engagement";
  if (t.includes("tag")) return "tags";
  if (t.includes("mention")) return "mentions";
  if (t.includes("vendor") || s.monitor.startsWith("vendor:")) return "vendor";
  return "all";
}

function typeVisual(s: Signal): { icon: string; cls: string } {
  const c = categorize(s);
  switch (c) {
    case "engagement":
      return { icon: "💍", cls: "feed-card__icon--mesa" };
    case "tags":
      return { icon: "🏷", cls: "feed-card__icon--cove" };
    case "mentions":
      return { icon: "💬", cls: "feed-card__icon--cove-ink" };
    case "vendor":
      return { icon: "📸", cls: "feed-card__icon--ink-dim" };
    default:
      return { icon: "📡", cls: "feed-card__icon--ink-dim" };
  }
}

function relativeTime(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime();
  const s = Math.floor(diff / 1000);
  if (s < 60) return `${s}s ago`;
  const m = Math.floor(s / 60);
  if (m < 60) return `${m}m ago`;
  const h = Math.floor(m / 60);
  if (h < 24) return `${h}h ago`;
  const d = Math.floor(h / 24);
  return `${d}d ago`;
}

interface SignalGroup {
  key: string;
  handle: string;
  primary: Signal;
  rest: Signal[];
}

// Group signals from the same handle observed within 5 minutes of each other.
function groupSignals(signals: Signal[]): SignalGroup[] {
  const WINDOW = 5 * 60 * 1000;
  const groups: SignalGroup[] = [];
  for (const s of signals) {
    const last = groups[groups.length - 1];
    if (
      last &&
      last.handle === s.handle &&
      new Date(last.primary.observed_at).getTime() - new Date(s.observed_at).getTime() <= WINDOW
    ) {
      last.rest.push(s);
    } else {
      groups.push({ key: s.id, handle: s.handle, primary: s, rest: [] });
    }
  }
  return groups;
}

export function FeedView() {
  const [monitorFilter, setMonitorFilter] = useState("");
  const [typeFilter, setTypeFilter] = useState<TypeFilter>("all");
  const [expanded, setExpanded] = useState<Set<string>>(new Set());
  const [newIds, setNewIds] = useState<Set<string>>(new Set());
  const [updatedAgo, setUpdatedAgo] = useState("just now");
  const [waiting, setWaiting] = useState(false);

  const { data: signals, error, isLoading, isFetching, dataUpdatedAt } = useSignals(monitorFilter || undefined);
  const { data: ingest, error: ingestError } = useIngestStatus();

  const prevIds = useRef<Set<string>>(new Set());
  const lastSignalAt = useRef<number>(Date.now());

  // Detect newly-arrived signals and highlight them for 3s.
  useEffect(() => {
    if (!signals || signals.length === 0) return;
    const ids = new Set(signals.map((s) => s.id));
    const fresh = signals.filter((s) => !prevIds.current.has(s.id)).map((s) => s.id);
    if (fresh.length && prevIds.current.size > 0) {
      setNewIds((prev) => new Set([...prev, ...fresh]));
      lastSignalAt.current = Date.now();
      const timers = fresh.map((id) =>
        setTimeout(() => setNewIds((prev) => {
          const next = new Set(prev);
          next.delete(id);
          return next;
        }), 3000),
      );
      return () => timers.forEach(clearTimeout);
    }
    prevIds.current = ids;
  }, [signals]);

  // "Updated Xs ago" ticker + "waiting for signals" after 30s idle.
  useEffect(() => {
    const tick = () => {
      if (dataUpdatedAt) {
        const s = Math.max(0, Math.floor((Date.now() - dataUpdatedAt) / 1000));
        setUpdatedAgo(s < 5 ? "just now" : `${s}s ago`);
      }
      setWaiting(Date.now() - lastSignalAt.current >= 30_000);
    };
    tick();
    const i = setInterval(tick, 1000);
    return () => clearInterval(i);
  }, [dataUpdatedAt]);

  const filtered = useMemo(() => {
    if (!signals) return [];
    return typeFilter === "all" ? signals : signals.filter((s) => categorize(s) === typeFilter);
  }, [signals, typeFilter]);

  const groups = useMemo(() => groupSignals(filtered), [filtered]);

  const liveState: { label: string; cls: string } = ingestError
    ? { label: "OFFLINE", cls: "feed-live--offline" }
    : ingest?.paused
      ? { label: "PAUSED", cls: "feed-live--paused" }
      : { label: "LIVE", cls: "feed-live--live" };

  const budgetPct =
    ingest && ingest.daily_budget ? Math.min(100, (ingest.results_used_today / ingest.daily_budget) * 100) : null;

  const toggleGroup = (key: string) =>
    setExpanded((prev) => {
      const next = new Set(prev);
      next.has(key) ? next.delete(key) : next.add(key);
      return next;
    });

  const renderCard = (s: Signal, isNew: boolean, isRefreshing: boolean) => {
    const vis = typeVisual(s);
    const img = (s as Signal & { image_url?: string }).image_url;
    return (
      <article
        key={s.id}
        className={`feed-card feed-card--${categorize(s)}${isNew ? " feed-card--new" : ""}${isRefreshing ? " feed-card--refreshing" : ""}`}
      >
        <div className={`feed-card__icon ${vis.cls}`} aria-hidden>{vis.icon}</div>
        <div className="feed-card__body">
          <div className="feed-card__handle">@{s.handle}</div>
          <div className="feed-card__summary">{s.summary}</div>
          <div className="feed-card__meta">
            <span className="feed-card__time">{relativeTime(s.observed_at)}</span>
            <span className="feed-card__monitor">{s.monitor}</span>
            <span className="feed-card__type">{s.observation_type}</span>
          </div>
        </div>
        <div className="feed-card__thumb" aria-hidden>
          {img ? <img src={img} alt="" loading="lazy" /> : <span className="feed-card__thumb-ph" />}
        </div>
      </article>
    );
  };

  return (
    <div className="view view--feed">
      <header className="feed-hero">
        <div className="feed-hero__text">
          <h2 className="feed-hero__title">Radar</h2>
          <p className="feed-hero__subtitle">Watching for love. Every signal is a couple beginning their story.</p>
        </div>
        <div className={`feed-live ${liveState.cls}`}>
          <span className="feed-live__dot" aria-hidden />
          {liveState.label}
        </div>
      </header>

      {budgetPct != null && (
        <div className="feed-budget" title={`Provider results used today: ${ingest!.results_used_today} / ${ingest!.daily_budget}`}>
          <div className="feed-budget__bar">
            <div className="feed-budget__fill" style={{ width: `${budgetPct}%` }} />
          </div>
          <span className="feed-budget__label">
            {ingest!.results_used_today}
            {ingest!.daily_budget != null ? ` / ${ingest!.daily_budget}` : ""} provider results today
          </span>
        </div>
      )}

      <div className="feed-filterbar">
        <div className="feed-chips" role="tablist" aria-label="Signal type filter">
          {TYPE_CHIPS.map((c) => (
            <button
              key={c.key}
              type="button"
              role="tab"
              aria-selected={typeFilter === c.key}
              className={`feed-chip${typeFilter === c.key ? " feed-chip--active" : ""}`}
              onClick={() => setTypeFilter(c.key)}
            >
              {c.label}
            </button>
          ))}
        </div>
        <div className="feed-search">
          <span className="feed-search__icon" aria-hidden>⌕</span>
          <input
            className="feed-search__input"
            placeholder="Filter by monitor (e.g. hashtag:justengaged, vendor:handle)…"
            value={monitorFilter}
            onChange={(e) => setMonitorFilter(e.target.value)}
          />
        </div>
      </div>

      {error && <EmptyState variant="warning" icon="⚠" title="Feed unavailable" message={(error as Error).message} />}

      <div className="feed-meta-row">
        {filtered.length > 0 && <span className="feed-meta-row__count">{filtered.length} signal{filtered.length === 1 ? "" : "s"}</span>}
        {filtered.length > 0 && <span className="feed-meta-row__updated">Updated {updatedAgo}</span>}
      </div>

      <div className="feed-stream">
        {isLoading ? (
          <LoadingState variant="skeleton" message="Loading signals…" />
        ) : !filtered || filtered.length === 0 ? (
          <EmptyState
            variant="empty"
            icon="📡"
            title="No signals yet"
            message="The radar is watching. Signals will appear here as couples are detected — check Sources to confirm vendors and budget are configured."
          />
        ) : (
          <>
            {waiting && !isFetching && (
              <div className="feed-waiting" aria-live="polite">
                <span className="feed-waiting__dot" aria-hidden />
                Waiting for signals…
              </div>
            )}
            {groups.map((g) => {
              const isRefreshing = isFetching && !newIds.has(g.primary.id);
              const isNew = newIds.has(g.primary.id);
              return (
                <div key={g.key} className="feed-group">
                  {renderCard(g.primary, isNew, isRefreshing)}
                  {g.rest.length > 0 && (
                    <>
                      <button
                        type="button"
                        className="feed-group__toggle"
                        onClick={() => toggleGroup(g.key)}
                        aria-expanded={expanded.has(g.key)}
                      >
                        {expanded.has(g.key) ? "Hide" : `+${g.rest.length} more`} from @{g.handle}
                      </button>
                      {expanded.has(g.key) && (
                        <div className="feed-group__rest">
                          {g.rest.map((s) => renderCard(s, newIds.has(s.id), isFetching && !newIds.has(s.id)))}
                        </div>
                      )}
                    </>
                  )}
                </div>
              );
            })}
          </>
        )}
      </div>
    </div>
  );
}
