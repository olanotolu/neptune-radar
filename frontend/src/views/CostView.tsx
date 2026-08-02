import { useIngestStatus, useOpsSummary } from "../api/hooks";
import { LoadingState } from "../components/LoadingState";

function fmt(n: number): string {
  return n.toLocaleString();
}

export function CostView() {
  const { data: status, isLoading } = useIngestStatus();
  const { data: ops } = useOpsSummary();

  if (isLoading) return <LoadingState variant="skeleton" message="Loading budget…" />;

  const provider = status?.provider ?? "—";
  const budget = status?.daily_budget ?? ops?.daily_budget ?? 0;
  const used = status?.results_used_today ?? ops?.results_used_today ?? 0;
  const remaining = Math.max(budget - used, 0);
  const usedPct = budget > 0 ? Math.min((used / budget) * 100, 100) : 0;
  const overBudget = budget > 0 && used > budget;

  const couplesTotal = ops?.couples_total ?? 0;
  const couples24h = ops?.couples_24h ?? 0;
  // ponytail: cost-per-couple uses budget (daily cap) as the cost proxy since
  // the API exposes no $ amount, only result-count budget. Ceiling: accurate
  // only while spend tracks results 1:1 with budget units.
  const costPerCouple = couplesTotal > 0 && used > 0 ? used / couplesTotal : 0;

  const sourcesTotal = ops?.sources_total ?? 0;
  const sourcesWithLoc = ops?.sources_with_loc ?? 0;
  const sourcesStale = ops?.sources_stale ?? 0;
  const stalePct = sourcesTotal > 0 ? (sourcesStale / sourcesTotal) * 100 : 0;
  const locPct = sourcesTotal > 0 ? (sourcesWithLoc / sourcesTotal) * 100 : 0;

  const paused = status?.paused ?? false;
  const running = status?.running ?? false;
  const providerAvailable = status?.provider_available ?? false;
  const pollInterval = status?.poll_interval ?? "—";

  return (
    <div className="view view--cost">
      <header className="view__head">
        <h2>Budget &amp; costs</h2>
        <p className="trust-panel__sub">Every signal has a cost. Every celebration has a budget.</p>
      </header>

      {/* Provider budget */}
      <section className="trust-panel">
        <h3 className="trust-panel__title">Provider budget — {provider}</h3>
        <div className="funnel-kpis">
          <div className="funnel-kpi">
            <strong>{fmt(used)}</strong>
            <span>Used today</span>
          </div>
          <div className="funnel-kpi">
            <strong>{fmt(budget)}</strong>
            <span>Daily budget</span>
          </div>
          <div className="funnel-kpi">
            <strong>{fmt(remaining)}</strong>
            <span>Remaining</span>
          </div>
          <div className="funnel-kpi">
            <strong>{Math.round(usedPct)}%</strong>
            <span>Used</span>
          </div>
        </div>
        <div className="budget-bar" role="progressbar" aria-valuenow={Math.round(usedPct)} aria-valuemin={0} aria-valuemax={100}>
          <div
            className={`budget-bar__fill ${overBudget ? "budget-bar__fill--over" : usedPct >= 80 ? "budget-bar__fill--warn" : ""}`}
            style={{ width: `${usedPct}%` }}
          />
        </div>
        {overBudget && <p className="budget-bar__note budget-bar__note--over">Over daily budget — ingest will self-throttle.</p>}
      </section>

      {/* Operational costs */}
      <section className="trust-panel">
        <h3 className="trust-panel__title">Operational cost</h3>
        <div className="funnel-kpis">
          <div className="funnel-kpi">
            <strong>{fmt(couplesTotal)}</strong>
            <span>Couples total</span>
          </div>
          <div className="funnel-kpi">
            <strong>{fmt(couples24h)}</strong>
            <span>Couples (24h)</span>
          </div>
          <div className="funnel-kpi">
            <strong>{fmt(used)}</strong>
            <span>Results used today</span>
          </div>
          <div className="funnel-kpi">
            <strong>{costPerCouple > 0 ? costPerCouple.toFixed(2) : "—"}</strong>
            <span>Results / couple</span>
          </div>
        </div>
      </section>

      {/* Source health */}
      <section className="trust-panel">
        <h3 className="trust-panel__title">Source health</h3>
        <div className="funnel-kpis">
          <div className="funnel-kpi">
            <strong>{fmt(sourcesTotal)}</strong>
            <span>Sources total</span>
          </div>
          <div className="funnel-kpi">
            <strong>{fmt(sourcesWithLoc)}</strong>
            <span>With location</span>
          </div>
          <div className="funnel-kpi">
            <strong>{fmt(sourcesStale)}</strong>
            <span>Stale</span>
          </div>
        </div>
        <div className="cost-bars">
          <div className="cost-bar-row">
            <span className="cost-bar-row__label">Geocoded</span>
            <div className="budget-bar">
              <div className="budget-bar__fill budget-bar__fill--cove" style={{ width: `${locPct}%` }} />
            </div>
            <span className="cost-bar-row__pct">{Math.round(locPct)}%</span>
          </div>
          <div className="cost-bar-row">
            <span className="cost-bar-row__label">Stale</span>
            <div className="budget-bar">
              <div className={`budget-bar__fill ${stalePct >= 50 ? "budget-bar__fill--warn" : ""}`} style={{ width: `${stalePct}%` }} />
            </div>
            <span className="cost-bar-row__pct">{Math.round(stalePct)}%</span>
          </div>
        </div>
      </section>

      {/* Provider status */}
      <section className="trust-panel">
        <h3 className="trust-panel__title">Provider status</h3>
        <div className="funnel-kpis">
          <div className="funnel-kpi">
            <strong className={running ? "status-dot status-dot--live" : "status-dot status-dot--idle"}>
              {running ? "Running" : "Idle"}
            </strong>
            <span>Worker</span>
          </div>
          <div className="funnel-kpi">
            <strong className={paused ? "status-dot status-dot--paused" : "status-dot status-dot--live"}>
              {paused ? "Paused" : "Active"}
            </strong>
            <span>Ingest</span>
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
        </div>
      </section>
    </div>
  );
}
