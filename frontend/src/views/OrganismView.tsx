import { useOrganism } from "../api/hooks";
import { EmptyState } from "../components/EmptyState";
import { LoadingState } from "../components/LoadingState";

function pct(n: number): string {
  return `${Math.round((n || 0) * 100)}%`;
}

export function OrganismView({ onNavigate }: { onNavigate?: (path: string) => void }) {
  const { data, isLoading, error } = useOrganism();

  if (isLoading) return <LoadingState message="Loading organism…" />;
  if (error) return <EmptyState variant="warning" title="Organism unavailable" message={(error as Error).message} />;
  if (!data) return <EmptyState title="No organism data" message="Radar has not published swarm status yet." />;

  const y = data.yield;
  const risk = data.risk_sentinel;

  return (
    <div className="view view--organism">
      <header className="organism-hero">
        <div>
          <h2 className="view__title">Growth organism</h2>
          <p className="view__subtitle">{data.thesis}</p>
        </div>
        <a
          className="btn btn--ghost"
          href={data.meet_neptune?.site || "https://www.meetneptune.com"}
          target="_blank"
          rel="noreferrer"
        >
          meetneptune.com
        </a>
      </header>

      <section className="organism-briefing">
        <h3 className="organism-section__title">Morning briefing</h3>
        <p className="organism-briefing__headline">{data.briefing.headline}</p>
        <ul className="organism-briefing__lines">
          {(data.briefing.lines || []).map((line) => (
            <li key={line}>{line}</li>
          ))}
        </ul>
        <div className="organism-briefing__actions">
          <button type="button" className="btn btn--primary" onClick={() => onNavigate?.("/work?filter=action")}>
            Open work queue
          </button>
          <button type="button" className="btn" onClick={() => onNavigate?.("/congratulate")}>
            Congratulate kits
          </button>
          <button type="button" className="btn btn--ghost" onClick={() => onNavigate?.("/funnel")}>
            Funnel detail
          </button>
        </div>
      </section>

      <section className="organism-section">
        <h3 className="organism-section__title">Swarm</h3>
        <div className="organism-swarm">
          {(data.swarm || []).map((a) => (
            <article key={a.id} className={`swarm-card swarm-card--${a.status}`}>
              <header className="swarm-card__top">
                <span className="swarm-card__name">{a.name}</span>
                <span className={`swarm-card__status swarm-card__status--${a.status}`}>{a.status}</span>
              </header>
              <p className="swarm-card__job">{a.job}</p>
              <p className="swarm-card__rule">{a.hard_rule}</p>
              <div className="swarm-card__metric">
                <span className="swarm-card__metric-val">{a.metric_value}</span>
                <span className="swarm-card__metric-label">{a.metric_label}</span>
              </div>
            </article>
          ))}
        </div>
      </section>

      <section className="organism-section">
        <h3 className="organism-section__title">Guarantees</h3>
        <div className="organism-guarantees">
          {(data.guarantees || []).map((g) => (
            <article key={g.id} className="guarantee-card">
              <header className="guarantee-card__top">
                <h4 className="guarantee-card__title">{g.title}</h4>
                <span className={`guarantee-card__status guarantee-card__status--${g.status}`}>{g.status}</span>
              </header>
              <p className="guarantee-card__promise">{g.promise}</p>
              <p className="guarantee-card__meta">
                Enforced by {g.enforced_by}
              </p>
              <p className="guarantee-card__evidence">{g.evidence}</p>
            </article>
          ))}
        </div>
      </section>

      <section className="organism-section">
        <h3 className="organism-section__title">Yield · closed loop</h3>
        <div className="organism-yield-kpis">
          <div className="organism-kpi">
            <span className="organism-kpi__n">{y.handoffs_issued}</span>
            <span className="organism-kpi__l">Handoffs</span>
          </div>
          <div className="organism-kpi">
            <span className="organism-kpi__n">{y.chats_7d}</span>
            <span className="organism-kpi__l">Chats 7d</span>
            <span className="organism-kpi__s">{pct(y.chat_rate)} rate</span>
          </div>
          <div className="organism-kpi">
            <span className="organism-kpi__n">{y.booked_7d}</span>
            <span className="organism-kpi__l">Booked 7d</span>
            <span className="organism-kpi__s">{pct(y.book_rate)} of chat</span>
          </div>
          <div className="organism-kpi">
            <span className="organism-kpi__n">{y.closed_won_7d}</span>
            <span className="organism-kpi__l">Closed won</span>
            <span className="organism-kpi__s">{pct(y.win_rate)} win rate</span>
          </div>
          <div className="organism-kpi">
            <span className="organism-kpi__n">{y.kits_mailed}</span>
            <span className="organism-kpi__l">Kits mailed</span>
            <span className="organism-kpi__s">{y.kits_ready} ready</span>
          </div>
        </div>

        {(y.by_market || []).length > 0 && (
          <div className="organism-table-wrap">
            <h4 className="organism-table__title">By market</h4>
            <table className="organism-table">
              <thead>
                <tr>
                  <th>Market</th>
                  <th>Couples</th>
                  <th>Invited+</th>
                  <th>Won</th>
                </tr>
              </thead>
              <tbody>
                {y.by_market.map((m) => (
                  <tr key={m.market}>
                    <td>{m.market}</td>
                    <td>{m.couples}</td>
                    <td>{m.invited}</td>
                    <td>{m.closed_won}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {(y.top_sources || []).length > 0 && (
          <div className="organism-table-wrap">
            <h4 className="organism-table__title">Top sources (90d)</h4>
            <table className="organism-table">
              <thead>
                <tr>
                  <th>Source</th>
                  <th>Signals</th>
                  <th>Won</th>
                </tr>
              </thead>
              <tbody>
                {y.top_sources.map((s) => (
                  <tr key={s.source}>
                    <td className="organism-table__mono">{s.source}</td>
                    <td>{s.signals}</td>
                    <td>{s.closed_won}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>

      <section className="organism-section">
        <h3 className="organism-section__title">Risk sentinel</h3>
        <p className="organism-risk__promise">{risk.promise}</p>
        <div className="organism-yield-kpis organism-yield-kpis--tight">
          <div className="organism-kpi">
            <span className="organism-kpi__n">{risk.risk_queue_open}</span>
            <span className="organism-kpi__l">Open risk queue</span>
          </div>
          <div className="organism-kpi">
            <span className="organism-kpi__n">{risk.pitches_blocked_30d}</span>
            <span className="organism-kpi__l">Risk paths 30d</span>
          </div>
        </div>
        {(risk.refusals || []).length > 0 ? (
          <ul className="organism-refusals">
            {risk.refusals.slice(0, 12).map((r, i) => (
              <li key={`${r.at}-${i}`} className="organism-refusal">
                <span className="organism-refusal__reason">{r.reason}</span>
                <span className="organism-refusal__meta">
                  {r.action}
                  {r.couple_id ? ` · ${r.couple_id.slice(0, 12)}…` : ""}
                  {" · "}
                  {r.at?.slice(0, 16)?.replace("T", " ")}
                </span>
              </li>
            ))}
          </ul>
        ) : (
          <p className="organism-empty-meta">No recent risk refusals logged.</p>
        )}
      </section>
    </div>
  );
}
