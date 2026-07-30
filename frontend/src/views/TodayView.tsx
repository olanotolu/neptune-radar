import { useBackfillLocations, useEnrichMissing, useOpsSummary, useRunJanitor } from "../api/hooks";
import { useToast } from "../components/Toast";

export function TodayView({ onNavigate }: { onNavigate: (path: string) => void }) {
  const { data, isLoading, error } = useOpsSummary();
  const enrich = useEnrichMissing();
  const backfill = useBackfillLocations();
  const janitor = useRunJanitor();
  const toast = useToast();

  const budget = data?.daily_budget ?? 0;
  const used = data?.results_used_today ?? 0;
  const budgetPct = budget > 0 ? Math.min(100, Math.round((used / budget) * 100)) : 0;
  const budgetTone = budgetPct >= 90 ? "hot" : budgetPct >= 70 ? "warm" : "ok";

  return (
    <div className="view view--today">
      <header className="prospects-hero">
        <div>
          <h2 className="view__title">Today</h2>
          <p className="view__subtitle">
            Operator snapshot — pending work, data quality, and provider health. Jump into Work when something needs a human.
          </p>
        </div>
      </header>

      {error && <div className="empty-state">{(error as Error).message}</div>}
      {isLoading && <div className="empty-state">Loading ops…</div>}

      {data && (
        <>
          <div className="kpi-grid">
            <button type="button" className="kpi" onClick={() => onNavigate("/work?filter=action")}>
              <span className="kpi__value">{data.pending_actions}</span>
              <span className="kpi__label">Pending approvals</span>
            </button>
            <button type="button" className="kpi" onClick={() => onNavigate("/work")}>
              <span className="kpi__value">{data.couples_total}</span>
              <span className="kpi__label">Prospects</span>
              <span className="kpi__sub">+{data.couples_24h} in 24h</span>
            </button>
            <button type="button" className="kpi kpi--warn" onClick={() => onNavigate("/work?filter=pics")}>
              <span className="kpi__value">{data.needs_pics}</span>
              <span className="kpi__label">Missing profile pics</span>
            </button>
            <button type="button" className="kpi kpi--warn" onClick={() => onNavigate("/work?filter=loc")}>
              <span className="kpi__value">{data.needs_location}</span>
              <span className="kpi__label">Missing location</span>
            </button>
            <button type="button" className="kpi" onClick={() => onNavigate("/sources")}>
              <span className="kpi__value">{data.sources_total}</span>
              <span className="kpi__label">Sources</span>
              <span className="kpi__sub">{data.sources_stale} stale</span>
            </button>
            <button type="button" className="kpi" onClick={() => onNavigate("/map")}>
              <span className="kpi__value">{data.map_pins}</span>
              <span className="kpi__label">Map pins</span>
            </button>
          </div>

          <div className="today-panels">
            <section className="today-panel">
              <h3 className="today-panel__title">Provider & budget</h3>
              <div className={`budget-bar budget-bar--${budgetTone}`}>
                <div className="budget-bar__fill" style={{ width: `${budgetPct}%` }} />
              </div>
              <p className="today-panel__meta">
                {used}
                {budget ? ` / ${budget}` : ""} results today · {budgetPct}% of daily cap
              </p>
              <p className="today-panel__meta">
                Radar:{" "}
                <strong>
                  {data.paused ? "Paused" : data.running ? "Live" : data.provider_available === false ? "No provider" : "Idle"}
                </strong>
                {data.poll_interval ? ` · poll ${data.poll_interval}` : ""}
              </p>
              {budgetPct >= 90 && (
                <p className="today-panel__alert">Daily budget nearly exhausted — scans may fall back to stored posts only.</p>
              )}
            </section>

            <section className="today-panel">
              <h3 className="today-panel__title">Data quality actions</h3>
              <div className="today-panel__actions">
                <button
                  type="button"
                  className="btn btn--primary"
                  disabled={enrich.isPending}
                  onClick={() =>
                    enrich.mutate(15, {
                      onSuccess: (r) => toast.push(`Enriched ${r.succeeded}/${r.attempted} profiles`, r.succeeded ? "ok" : "err"),
                      onError: (e) => toast.push((e as Error).message, "err"),
                    })
                  }
                >
                  {enrich.isPending ? "Enriching…" : "Enrich missing pics"}
                </button>
                <button
                  type="button"
                  className="btn btn--ghost"
                  disabled={backfill.isPending}
                  onClick={() =>
                    backfill.mutate(100, {
                      onSuccess: (r) => toast.push(`Locations updated ${r.updated}/${r.checked}`, "ok"),
                      onError: (e) => toast.push((e as Error).message, "err"),
                    })
                  }
                >
                  {backfill.isPending ? "Backfilling…" : "Backfill map locations"}
                </button>
                <button type="button" className="btn btn--ghost" onClick={() => onNavigate("/work?filter=action")}>
                  Open work queue
                </button>
                <button type="button" className="btn btn--ghost" onClick={() => onNavigate("/congratulate")}>
                  Congratulate kits
                </button>
                <button
                  type="button"
                  className="btn btn--ghost"
                  disabled={janitor.isPending}
                  onClick={() =>
                    janitor.mutate(undefined, {
                      onSuccess: (r) =>
                        toast.push(
                          `Janitor: ${r.vendor_pairs_suppressed} vendor pairs · ${r.observation_facts_backfilled} posts structured`,
                          "ok",
                        ),
                      onError: (e) => toast.push((e as Error).message, "err"),
                    })
                  }
                >
                  {janitor.isPending ? "Cleaning…" : "Run janitor cleanup"}
                </button>
              </div>
            </section>
          </div>
        </>
      )}
    </div>
  );
}
