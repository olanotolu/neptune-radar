import { AuditTable } from "../components/AuditTable";
import { useToast } from "../components/Toast";
import { LoadingState } from "../components/LoadingState";
import {
  useAudit,
  useBackfillLocations,
  useEnrichMissing,
  useIngestStatus,
  useOpsSummary,
  usePauseIngest,
  useResumeIngest,
  useRunJanitor,
} from "../api/hooks";

export function OpsView() {
  const { data: status, isLoading } = useIngestStatus();
  const { data: ops } = useOpsSummary();
  const { data: audit } = useAudit();
  const toast = useToast();

  const enrich = useEnrichMissing();
  const backfill = useBackfillLocations();
  const janitor = useRunJanitor();
  const pause = usePauseIngest();
  const resume = useResumeIngest();

  if (isLoading) return <LoadingState variant="skeleton" message="Loading ops…" />;

  const running = status?.running ?? false;
  const paused = status?.paused ?? false;
  const providerAvailable = status?.provider_available ?? false;
  const pollInterval = status?.poll_interval ?? "—";

  const pendingActions = ops?.pending_actions ?? 0;
  const needsPics = ops?.needs_pics ?? 0;
  const needsLocation = ops?.needs_location ?? 0;

  const busy = enrich.isPending || backfill.isPending || janitor.isPending || pause.isPending || resume.isPending;

  function run(
    mut: ReturnType<typeof useEnrichMissing>,
    label: string,
  ) {
    mut.mutate(50, {
      onSuccess: (r) => toast.push(`${label}: ${JSON.stringify(r)}`, "ok"),
      onError: (e) => toast.push((e as Error).message, "err"),
    });
  }

  return (
    <div className="view view--ops">
      <header className="view__head">
        <h2>Ops health</h2>
        <p className="trust-panel__sub">The machinery behind the radar. Everything running, everything healthy.</p>
      </header>

      {/* System health overview */}
      <section className="trust-panel">
        <h3 className="trust-panel__title">System health</h3>
        <div className="funnel-kpis">
          <div className="funnel-kpi">
            <strong className={running ? "status-dot status-dot--live" : "status-dot status-dot--idle"}>
              {running ? "Running" : "Idle"}
            </strong>
            <span>Ingest worker</span>
          </div>
          <div className="funnel-kpi">
            <strong className={paused ? "status-dot status-dot--paused" : "status-dot status-dot--live"}>
              {paused ? "Paused" : "Active"}
            </strong>
            <span>Ingest state</span>
          </div>
          <div className="funnel-kpi">
            <strong className={providerAvailable ? "status-dot status-dot--live" : "status-dot status-dot--idle"}>
              {providerAvailable ? "Available" : "Unavailable"}
            </strong>
            <span>Provider</span>
          </div>
          <div className="funnel-kpi">
            <strong>{pollInterval}</strong>
            <span>Poll interval</span>
          </div>
          <div className="funnel-kpi">
            <strong>{pendingActions}</strong>
            <span>Pending actions</span>
          </div>
          <div className="funnel-kpi">
            <strong>{needsPics}</strong>
            <span>Needs pics</span>
          </div>
          <div className="funnel-kpi">
            <strong>{needsLocation}</strong>
            <span>Needs location</span>
          </div>
        </div>
      </section>

      {/* Quick actions */}
      <section className="trust-panel">
        <h3 className="trust-panel__title">Quick actions</h3>
        <div className="ops-actions">
          <button type="button" className="btn" disabled={busy} onClick={() => run(enrich, "Enrich")}>
            {enrich.isPending ? "Enriching…" : "Enrich Missing Profiles"}
          </button>
          <button type="button" className="btn" disabled={busy} onClick={() => run(backfill, "Backfill")}>
            {backfill.isPending ? "Backfilling…" : "Backfill Locations"}
          </button>
          <button type="button" className="btn" disabled={busy} onClick={() => run(janitor, "Janitor")}>
            {janitor.isPending ? "Running…" : "Run Janitor"}
          </button>
          {paused ? (
            <button
              type="button"
              className="btn btn--primary"
              disabled={busy}
              onClick={() =>
                resume.mutate(undefined, {
                  onSuccess: () => toast.push("Ingest resumed", "ok"),
                  onError: (e) => toast.push((e as Error).message, "err"),
                })
              }
            >
              {resume.isPending ? "Resuming…" : "Resume Ingest"}
            </button>
          ) : (
            <button
              type="button"
              className="btn"
              disabled={busy}
              onClick={() =>
                pause.mutate(undefined, {
                  onSuccess: () => toast.push("Ingest paused", "ok"),
                  onError: (e) => toast.push((e as Error).message, "err"),
                })
              }
            >
              {pause.isPending ? "Pausing…" : "Pause Ingest"}
            </button>
          )}
        </div>
      </section>

      {/* Ingest cursor status */}
      <section className="trust-panel">
        <h3 className="trust-panel__title">Ingest cursors</h3>
        {(status?.cursors ?? []).length === 0 ? (
          <p className="work-drawer__muted">No active monitors.</p>
        ) : (
          <table className="dossier-ledger">
            <thead>
              <tr>
                <th>Monitor</th>
                <th>Last seen</th>
                <th>Last run</th>
                <th>Updated</th>
              </tr>
            </thead>
            <tbody>
              {(status?.cursors ?? []).map((c) => (
                <tr key={c.monitor}>
                  <td><code>{c.monitor}</code></td>
                  <td>{c.last_seen_at?.slice(0, 19)?.replace("T", " ") ?? "—"}</td>
                  <td>{c.last_run_at?.slice(0, 19)?.replace("T", " ") ?? "—"}</td>
                  <td>{c.updated_at?.slice(0, 19)?.replace("T", " ") ?? "—"}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>

      {/* Recent audit events */}
      <section className="trust-panel">
        <h3 className="trust-panel__title">Recent audit events</h3>
        <AuditTable events={(audit ?? []).slice(0, 20)} />
      </section>
    </div>
  );
}
