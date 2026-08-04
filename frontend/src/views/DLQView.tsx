import { useDLQ, useReplayDLQ } from "../api/hooks";
import { useToast } from "../components/Toast";
import { EmptyState } from "../components/EmptyState";
import { LoadingState } from "../components/LoadingState";

export function DLQView() {
  const { data, isLoading, error } = useDLQ("pending", 50);
  const replay = useReplayDLQ();
  const toast = useToast();

  const items = data ?? [];

  return (
    <div className="view view--dlq">
      <header className="ops-view__header">
        <div>
          <h2 className="ops-view__title">Failed items</h2>
          <p className="ops-view__sub">
            Signals that didn't make it. Replay them to try again.
          </p>
        </div>
        <span className={`ops-badge ${items.length > 0 ? "ops-badge--warn" : "ops-badge--ok"}`}>
          {items.length} pending item{items.length === 1 ? "" : "s"}
        </span>
      </header>

      {isLoading && <LoadingState variant="skeleton" message="Loading DLQ…" />}
      {error && <EmptyState variant="warning" title="DLQ unavailable" message={(error as Error).message} />}

      {!isLoading && !error && items.length === 0 && (
        <EmptyState variant="success" title="All clear" message="No failed signals. Everything made it through." />
      )}

      {!isLoading && !error && items.length > 0 && (
        <table className="dossier-ledger">
          <thead>
            <tr>
              <th>ID</th>
              <th>Source</th>
              <th>Monitor</th>
              <th>Error</th>
              <th>Retries</th>
              <th>Created</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {items.map((it) => (
              <tr key={it.id}>
                <td><code>{it.id.slice(0, 12)}…</code></td>
                <td>{it.source}</td>
                <td>{it.monitor ?? "—"}</td>
                <td className="ops-view__error">{it.error_message}</td>
                <td>{it.retries}</td>
                <td>{it.created_at?.slice(0, 16)?.replace("T", " ")}</td>
                <td>
                  <button
                    type="button"
                    className="btn btn--primary"
                    disabled={replay.isPending}
                    onClick={() =>
                      replay.mutate(it.id, {
                        onSuccess: () => toast.push("Replay queued", "ok"),
                        onError: (e) => toast.push((e as Error).message, "err"),
                      })
                    }
                  >
                    Replay
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
