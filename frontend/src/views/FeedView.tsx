import { useState } from "react";
import { useIngestStatus, useSignals } from "../api/hooks";

export function FeedView() {
  const [monitorFilter, setMonitorFilter] = useState("");
  const { data: signals, error, isLoading } = useSignals(monitorFilter || undefined);
  const { data: ingest } = useIngestStatus();

  return (
    <div className="view">
      <h2 className="view__title">Live signal feed</h2>
      <p className="view__subtitle">
        Raw observations from the watch loop — normalized, not yet interpreted. Newest first.
        {ingest && (
          <>
            {" "}
            Provider results used today: <strong>{ingest.results_used_today}</strong>
            {ingest.daily_budget != null ? ` / ${ingest.daily_budget}` : ""}.
            {ingest.paused ? (
              <>
                {" "}
                <strong className="feed-paused-hint">Radar paused</strong> — use Play in the header to resume.
              </>
            ) : null}
          </>
        )}
      </p>
      <div className="feed-toolbar">
        <input
          className="feed-filter"
          placeholder="Filter by monitor (e.g. hashtag:justengaged, vendor:handle)…"
          value={monitorFilter}
          onChange={(e) => setMonitorFilter(e.target.value)}
        />
      </div>
      {error && <div className="empty-state">Feed unavailable: {(error as Error).message}</div>}
      <div className="feed-list">
        {isLoading ? (
          <div className="empty-state">Loading signals…</div>
        ) : !signals || signals.length === 0 ? (
          <div className="empty-state">
            No signals yet. The watch loop ingests on its poll interval — check Sources to confirm vendors and
            budget are configured.
          </div>
        ) : (
          signals.map((s) => (
            <div key={s.id} className={`feed-item feed-item--${s.observation_type}`}>
              <div className="feed-item__handle">@{s.handle}</div>
              <div className="feed-item__summary">{s.summary}</div>
              <div className="feed-item__meta">
                <span className="feed-item__type">{s.observation_type}</span>
                <span className="feed-item__monitor">{s.monitor}</span>
                <span className="feed-item__time">{new Date(s.observed_at).toLocaleString()}</span>
              </div>
            </div>
          ))
        )}
      </div>
    </div>
  );
}
