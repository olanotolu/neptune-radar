import { useScanJobs } from "../api/hooks";
import type { ScanJob } from "../api/types";
import { EmptyState } from "../components/EmptyState";
import { LoadingState } from "../components/LoadingState";

const STATUS_TONE: Record<string, string> = {
  running: "job-badge--running",
  done: "job-badge--done",
  failed: "job-badge--failed",
  queued: "job-badge--queued",
};

export function JobsView() {
  const { data, isLoading, error } = useScanJobs(50, "all");
  const jobs = data ?? [];

  return (
    <div className="view view--jobs">
      <header className="ops-view__header">
        <div>
          <h2 className="ops-view__title">Job queue</h2>
          <p className="ops-view__sub">
            Background scans running across all sources.
          </p>
        </div>
        <span className="ops-badge ops-badge--ok">{jobs.length} job{jobs.length === 1 ? "" : "s"}</span>
      </header>

      {isLoading && <LoadingState variant="skeleton" message="Loading jobs…" />}
      {error && <EmptyState variant="warning" icon="⚠" title="Jobs unavailable" message={(error as Error).message} />}

      {!isLoading && !error && jobs.length === 0 && (
        <EmptyState variant="empty" icon="🛠️" title="No scan jobs yet" message="Background scans will appear here as the radar queues them." />
      )}

      {!isLoading && !error && jobs.length > 0 && (
        <table className="dossier-ledger">
          <thead>
            <tr>
              <th>Job ID</th>
              <th>Handle</th>
              <th>Status</th>
              <th>Started</th>
              <th>Completed</th>
              <th>Items found</th>
              <th>Error</th>
            </tr>
          </thead>
          <tbody>
            {jobs.map((j: ScanJob) => (
              <tr key={j.id}>
                <td><code>{j.id.slice(0, 12)}…</code></td>
                <td>{j.handle ?? "—"}</td>
                <td>
                  <span className={`job-badge ${STATUS_TONE[j.status] ?? ""}`}>{j.status}</span>
                </td>
                <td>{j.created_at?.slice(0, 16)?.replace("T", " ")}</td>
                <td>{j.updated_at?.slice(0, 16)?.replace("T", " ")}</td>
                <td>{j.result?.actions_created ?? j.results?.reduce((n, r) => n + r.actions_created, 0) ?? 0}</td>
                <td className="ops-view__error">{j.error ?? "—"}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
