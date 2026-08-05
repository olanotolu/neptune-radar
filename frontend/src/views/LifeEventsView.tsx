import { useFenrisEvents } from "../api/hooks";
import type { FenrisEvent } from "../api/types";
import { EmptyState } from "../components/EmptyState";
import { LoadingState } from "../components/LoadingState";

function relativeTime(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime();
  const d = Math.floor(diff / 86_400_000);
  if (d > 0) return `${d}d ago`;
  const h = Math.floor(diff / 3_600_000);
  if (h > 0) return `${h}h ago`;
  const m = Math.floor(diff / 60_000);
  return m > 0 ? `${m}m ago` : "just now";
}

export function LifeEventsView() {
  const { data, isLoading, error } = useFenrisEvents();
  const events = data ?? [];

  const validated = events.filter((e) => e.cross_validated).length;

  return (
    <div className="view view--life-events">
      <header className="ops-view__header">
        <div>
          <h2 className="ops-view__title">Life Events</h2>
          <p className="ops-view__sub">
            Fenris Digital cross-validation — two independent signals for every couple.
          </p>
        </div>
        <span className={`ops-badge ${events.length > 0 ? "ops-badge--warn" : "ops-badge--ok"}`}>
          {events.length} event{events.length === 1 ? "" : "s"}
          {validated > 0 && ` · ${validated} cross-validated`}
        </span>
      </header>

      {isLoading && <LoadingState variant="skeleton" message="Loading life events…" />}
      {error && (
        <EmptyState variant="warning" title="Life events unavailable" message={(error as Error).message} />
      )}

      {!isLoading && !error && events.length === 0 && (
        <EmptyState
          variant="info"
          title="No life events yet"
          message="Fenris Digital life events will appear here when FENRIS_API_KEY is configured."
        />
      )}

      {!isLoading && !error && events.length > 0 && (
        <table className="dossier-ledger">
          <thead>
            <tr>
              <th>Event Type</th>
              <th>Person</th>
              <th>Location</th>
              <th>Event Date</th>
              <th>Confidence</th>
              <th>Cross-Validated</th>
              <th>Ingested</th>
            </tr>
          </thead>
          <tbody>
            {events.map((e: FenrisEvent) => (
              <tr key={e.id}>
                <td>
                  <span className={`status-dot ${e.event_type === "Newly Engaged" ? "status-dot--live" : "status-dot--ok"}`} />
                  {e.event_type}
                </td>
                <td>{e.person_name}</td>
                <td>
                  {e.city ? `${e.city}, ` : ""}{e.state || "—"}
                </td>
                <td>{e.event_date?.slice(0, 10)}</td>
                <td>{(e.confidence * 100).toFixed(0)}%</td>
                <td>
                  {e.cross_validated ? (
                    <span className="status-dot status-dot--ok" title="Cross-validated with Instagram" />
                  ) : (
                    <span className="status-dot status-dot--idle" title="Fenris only" />
                  )}
                </td>
                <td>{relativeTime(e.ingested_at)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
