import { useProviderAccuracy } from "../api/hooks";
import type { ProviderAccuracyRow } from "../api/types";
import { EmptyState } from "../components/EmptyState";
import { LoadingState } from "../components/LoadingState";

function accuracyTone(acc: number): string {
  if (acc >= 0.75) return "status-dot--live";
  if (acc >= 0.5) return "status-dot--idle";
  return "status-dot--paused";
}

export function ProviderAccuracyView() {
  const { data: rows, error, isLoading } = useProviderAccuracy();

  // Pivot: provider → state → row, for a providers × states matrix.
  const providers = new Map<string, Map<string, ProviderAccuracyRow>>();
  const states = new Set<string>();
  for (const r of rows ?? []) {
    if (!providers.has(r.provider)) providers.set(r.provider, new Map());
    providers.get(r.provider)!.set(r.state, r);
    states.add(r.state);
  }
  const providerNames = [...providers.keys()].sort();
  const stateList = [...states].sort();

  return (
    <div className="view view--system">
      <h2 className="view__title">Provider accuracy</h2>
      <p className="view__subtitle">
        Bayesian fusion — each provider's historical hit rate per state. The system gets smarter every run.
      </p>

      <section className="trust-panel">
        <header className="trust-panel__header">
          <div>
            <h3 className="trust-panel__title">Accuracy by provider × state</h3>
            <p className="trust-panel__sub">
              Accuracy = successful / total attempts. Cold start uses 0.50 prior when no data exists.
            </p>
          </div>
        </header>

        {error ? (
          <EmptyState variant="warning" title="Provider accuracy unavailable" message={(error as Error).message} />
        ) : isLoading ? (
          <LoadingState variant="skeleton" message="Loading accuracy scores…" />
        ) : !rows || rows.length === 0 ? (
          <EmptyState variant="empty" title="No accuracy data yet" message="Run the detective with Lob verification to start collecting accuracy scores." />
        ) : (
          <div style={{ overflowX: "auto" }}>
            <table className="dossier-ledger">
              <thead>
                <tr>
                  <th style={{ textAlign: "left" }}>Provider</th>
                  {stateList.map((st) => (
                    <th key={st} style={{ textAlign: "right", fontFamily: "var(--font-mono, monospace)" }}>{st}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {providerNames.map((p) => {
                  const stateMap = providers.get(p)!;
                  const totalAttempts = [...stateMap.values()].reduce((s, r) => s + r.total_attempts, 0);
                  const totalSuccess = [...stateMap.values()].reduce((s, r) => s + r.successful, 0);
                  return (
                    <tr key={p}>
                      <td style={{ textAlign: "left" }}>
                        <strong>{p}</strong>
                        <span style={{ color: "var(--text-muted, #999)", marginLeft: "0.5rem", fontFamily: "var(--font-mono, monospace)" }}>
                          {totalSuccess}/{totalAttempts}
                        </span>
                      </td>
                      {stateList.map((st) => {
                        const r = stateMap.get(st);
                        if (!r) return <td key={st} style={{ textAlign: "right", color: "var(--text-muted, #999)" }}>—</td>;
                        const pct = Math.round(r.accuracy * 100);
                        return (
                          <td key={st} style={{ textAlign: "right", fontFamily: "var(--font-mono, monospace)" }}>
                            <span className={`status-dot ${accuracyTone(r.accuracy)}`} aria-hidden />
                            {pct}%
                          </td>
                        );
                      })}
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </div>
  );
}
