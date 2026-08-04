import { useState } from "react";
import { useRuns, useRun } from "../api/hooks";
import { EmptyState } from "../components/EmptyState";
import { LoadingState } from "../components/LoadingState";
import type { PipelineRun } from "../api/types";

const STOP_LABELS: Record<string, string> = {
  completed: "Completed",
  no_signal: "No signal",
  policy_no_action: "No action",
  mistaken_couple: "Mistaken",
  duplicate: "Duplicate",
  error: "Error",
};

const STOP_TONES: Record<string, string> = {
  completed: "run-stop--ok",
  no_signal: "run-stop--muted",
  policy_no_action: "run-stop--muted",
  mistaken_couple: "run-stop--warn",
  duplicate: "run-stop--muted",
  error: "run-stop--err",
};

function timeAgo(iso: string): string {
  const d = new Date(iso).getTime();
  const s = Math.floor((Date.now() - d) / 1000);
  if (s < 60) return `${s}s ago`;
  if (s < 3600) return `${Math.floor(s / 60)}m ago`;
  if (s < 86400) return `${Math.floor(s / 3600)}h ago`;
  return `${Math.floor(s / 86400)}d ago`;
}

function fmtTokens(n: number): string {
  if (n === 0) return "—";
  if (n >= 1000) return `${(n / 1000).toFixed(1)}k`;
  return String(n);
}

export function AgentRunsView() {
  const { data, isLoading, error } = useRuns();
  const [selectedId, setSelectedId] = useState<string | undefined>();
  const runs = data ?? [];

  return (
    <div className="view view--runs">
      <header className="ops-view__header">
        <div>
          <h2 className="ops-view__title">Agent runs</h2>
          <p className="ops-view__sub">
            Every pipeline execution: what it observed, what it concluded, what
            it cost, and why it stopped.
          </p>
        </div>
        <span className="ops-badge ops-badge--ok">{runs.length} run{runs.length === 1 ? "" : "s"}</span>
      </header>

      {isLoading && <LoadingState variant="skeleton" message="Loading runs…" />}
      {error && <EmptyState variant="warning" title="Runs unavailable" message={(error as Error).message} />}

      {!isLoading && !error && runs.length === 0 && (
        <EmptyState variant="info" title="No runs yet" message="Pipeline runs will appear here as signals are processed." />
      )}

      {!isLoading && !error && runs.length > 0 && (
        <div className="runs-layout">
          <table className="dossier-ledger runs-table">
            <thead>
              <tr>
                <th>Stopped</th>
                <th>Confidence</th>
                <th>Model</th>
                <th>Tokens</th>
                <th>Monitor</th>
                <th>When</th>
              </tr>
            </thead>
            <tbody>
              {runs.map((r: PipelineRun) => (
                <tr
                  key={r.id}
                  className={selectedId === r.id ? "runs-row--active" : ""}
                  onClick={() => setSelectedId(r.id)}
                  style={{ cursor: "pointer" }}
                >
                  <td>
                    <span className={`run-stop ${STOP_TONES[r.stop_reason] ?? ""}`}>
                      {STOP_LABELS[r.stop_reason] ?? r.stop_reason}
                    </span>
                  </td>
                  <td>
                    {r.confidence != null ? `${Math.round(r.confidence * 100)}%` : "—"}
                  </td>
                  <td><code className="run-model">{r.model || "—"}</code></td>
                  <td>
                    {r.prompt_tokens > 0 || r.completion_tokens > 0
                      ? `${fmtTokens(r.prompt_tokens)}→${fmtTokens(r.completion_tokens)}`
                      : "—"}
                  </td>
                  <td>{r.monitor || "—"}</td>
                  <td className="run-when">{timeAgo(r.created_at)}</td>
                </tr>
              ))}
            </tbody>
          </table>

          {selectedId && <RunDrawer id={selectedId} onClose={() => setSelectedId(undefined)} />}
        </div>
      )}
    </div>
  );
}

function RunDrawer({ id, onClose }: { id: string; onClose: () => void }) {
  const { data, isLoading } = useRun(id);

  return (
    <aside className="run-drawer">
      <header className="run-drawer__header">
        <h3 className="run-drawer__title">Run detail</h3>
        <button className="run-drawer__close" onClick={onClose} aria-label="Close">Close</button>
      </header>

      {isLoading && <LoadingState variant="dots" message="Loading run…" />}

      {data && (
        <div className="run-drawer__body">
          <dl className="run-meta">
            <dt>Run</dt><dd><code>{data.id.slice(0, 16)}…</code></dd>
            <dt>Observation</dt><dd><code>{data.observation_id.slice(0, 16)}…</code></dd>
            <dt>Agent</dt><dd>{data.agent_name}</dd>
            <dt>Model</dt><dd><code>{data.model || "—"}</code></dd>
            <dt>Confidence</dt><dd>{data.confidence != null ? `${Math.round(data.confidence * 100)}%` : "—"}</dd>
            <dt>Stop reason</dt><dd><span className={`run-stop ${STOP_TONES[data.stop_reason] ?? ""}`}>{STOP_LABELS[data.stop_reason] ?? data.stop_reason}</span></dd>
            <dt>Tokens</dt><dd>{data.prompt_tokens > 0 ? `${data.prompt_tokens} in → ${data.completion_tokens} out` : "—"}</dd>
            {data.hypothesis_id && <><dt>Hypothesis</dt><dd><code>{data.hypothesis_id.slice(0, 16)}…</code></dd></>}
            {data.action_id && <><dt>Action</dt><dd><code>{data.action_id.slice(0, 16)}…</code></dd></>}
            {data.couple_id && <><dt>Couple</dt><dd><code>{data.couple_id.slice(0, 16)}…</code></dd></>}
            {data.monitor && <><dt>Monitor</dt><dd>{data.monitor}</dd></>}
          </dl>

          {data.timings.length > 0 && (
            <section className="run-section">
              <h4 className="run-section__title">Stage timings</h4>
              <div className="run-timings">
                {data.timings.map((t, i) => (
                  <div key={i} className="run-timing">
                    <span className="run-timing__stage">{t.stage}</span>
                    <span className="run-timing__dur">{t.duration_ms}ms</span>
                  </div>
                ))}
              </div>
            </section>
          )}

          {data.events.length > 0 && (
            <section className="run-section">
              <h4 className="run-section__title">Audit trail</h4>
              <ol className="run-events">
                {data.events.map((e) => (
                  <li key={e.id} className="run-event">
                    <span className="run-event__type">{e.event}</span>
                    <span className="run-event__time">{timeAgo(e.created_at)}</span>
                  </li>
                ))}
              </ol>
            </section>
          )}
        </div>
      )}
    </aside>
  );
}
