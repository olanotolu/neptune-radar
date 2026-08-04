import { useEffect, useState } from "react";
import { useFunnelEvents, useFunnelStats } from "../api/hooks";
import { EmptyState } from "../components/EmptyState";
import { LoadingState } from "../components/LoadingState";

function pct(n: number): string {
  return `${Math.round((n || 0) * 100)}%`;
}

type Stage = {
  label: string;
  shortLabel: string;
  count: number;
  width: number;
  convFromPrev: number | null;
  cumulative: number;
  color: string;
  gradient: string;
};

export function FunnelView() {
  const { data: stats, isLoading } = useFunnelStats();
  const { data: events } = useFunnelEvents();
  const [animated, setAnimated] = useState(false);

  useEffect(() => {
    const t = requestAnimationFrame(() => setAnimated(true));
    return () => cancelAnimationFrame(t);
  }, [stats]);

  if (isLoading) return <LoadingState variant="skeleton" message="Loading funnel…" />;
  const hasData = stats && (stats.handoffs_issued > 0 || stats.chat_started_7d > 0 || stats.consult_booked_7d > 0 || stats.closed_won_7d > 0 || stats.closed_lost_7d > 0);
  if (!stats || !hasData)
    return (
      <EmptyState
        variant="empty"
        title="No funnel data yet"
        message="Funnel analytics will appear once handoffs are issued and product events (chat_started, consult_booked, closed_won) are recorded via webhook."
      />
    );

  const chatRate = stats.chat_rate || 0;
  const bookRate = stats.book_rate || 0;
  const closeRate = stats.consult_booked_7d > 0 ? stats.closed_won_7d / stats.consult_booked_7d : 0;

  const stages: Stage[] = [
    {
      label: "Handoffs Issued",
      shortLabel: "Handoffs",
      count: stats.handoffs_issued,
      width: 1,
      convFromPrev: null,
      cumulative: 1,
      color: "var(--ink)",
      gradient: "var(--ink)",
    },
    {
      label: "Chats Started (7d)",
      shortLabel: "Chats",
      count: stats.chat_started_7d,
      width: chatRate,
      convFromPrev: stats.handoffs_issued > 0 ? stats.chat_started_7d / stats.handoffs_issued : 0,
      cumulative: chatRate,
      color: "var(--ink-soft)",
      gradient: "var(--ink-soft)",
    },
    {
      label: "Consults Booked (7d)",
      shortLabel: "Booked",
      count: stats.consult_booked_7d,
      width: bookRate * chatRate,
      convFromPrev: stats.chat_started_7d > 0 ? stats.consult_booked_7d / stats.chat_started_7d : 0,
      cumulative: bookRate * chatRate,
      color: "var(--ink-dim)",
      gradient: "var(--ink-dim)",
    },
    {
      label: "Closed Won (7d)",
      shortLabel: "Won",
      count: stats.closed_won_7d,
      width: closeRate * bookRate * chatRate,
      convFromPrev: stats.consult_booked_7d > 0 ? stats.closed_won_7d / stats.consult_booked_7d : 0,
      cumulative: closeRate * bookRate * chatRate,
      color: "var(--ink)",
      gradient: "var(--ink)",
    },
  ];

  return (
    <div className="view view--funnel">
      {/* Hero header */}
      <header className="funnel-hero">
        <div>
          <h2 className="funnel-hero__title">Growth funnel</h2>
          <p className="funnel-hero__sub">
            From first signal to signed prenup. Every step measured.
          </p>
        </div>
        <div className="funnel-hero__rates">
          <div className="funnel-rate">
            <span className="funnel-rate__label">Chat rate</span>
            <span className="funnel-rate__value">{pct(chatRate)}</span>
          </div>
          <div className="funnel-rate">
            <span className="funnel-rate__label">Book rate</span>
            <span className="funnel-rate__value">{pct(bookRate)}</span>
          </div>
          <div className="funnel-rate">
            <span className="funnel-rate__label">Close rate</span>
            <span className="funnel-rate__value">{pct(closeRate)}</span>
          </div>
        </div>
      </header>

      {/* Visual funnel — trapezoid steps */}
      <section className="funnel-viz">
        <div className="funnel-viz__stages">
          {stages.map((s, i) => {
            const w = animated ? Math.max(s.width * 100, 5) : 0;
            const nextWidth = i < stages.length - 1 ? Math.max(stages[i + 1].width * 100, 5) : w;
            return (
              <div key={s.label} className="funnel-step" style={{ "--w": `${w}%`, "--next-w": `${nextWidth}%` } as React.CSSProperties}>
                <div className="funnel-step__bar" style={{ background: s.gradient }}>
                  <div className="funnel-step__content">
                    <span className="funnel-step__label">{s.label}</span>
                    <span className="funnel-step__count">{s.count}</span>
                  </div>
                  {s.convFromPrev !== null && (
                    <span className="funnel-step__conv">
                      {pct(s.convFromPrev)}
                      <small>from prev</small>
                    </span>
                  )}
                </div>
                {i < stages.length - 1 && (
                  <div className="funnel-step__drop">
                    <span className="funnel-step__drop-num">
                      {stages[i].count - stages[i + 1].count > 0 ? `−${stages[i].count - stages[i + 1].count}` : ""}
                    </span>
                    <span className="funnel-step__drop-line" />
                  </div>
                )}
              </div>
            );
          })}
        </div>

        {/* Conversion rate sparkline */}
        <div className="funnel-viz__summary">
          <div className="funnel-summary-stat">
            <span className="funnel-summary-stat__num">{stats.handoffs_issued}</span>
            <span className="funnel-summary-stat__label">total handoffs</span>
          </div>
          <div className="funnel-summary-stat">
            <span className="funnel-summary-stat__num">{pct(stages[3].cumulative)}</span>
            <span className="funnel-summary-stat__label">end-to-end conversion</span>
          </div>
          <div className="funnel-summary-stat">
            <span className="funnel-summary-stat__num">
              {stats.closed_won_7d}/{stats.closed_won_7d + stats.closed_lost_7d || 0}
            </span>
            <span className="funnel-summary-stat__label">won / decided (7d)</span>
          </div>
        </div>
      </section>

      {/* 7-day KPI cards */}
      <section className="funnel-kpis-section">
        <h3 className="funnel-kpis-section__title">7-day performance</h3>
        <div className="funnel-kpis-grid">
          <div className="funnel-kpi-card funnel-kpi-card--blue">
            <span className="funnel-kpi-card__num">{stats.chat_started_7d}</span>
            <span className="funnel-kpi-card__label">Chats started</span>
          </div>
          <div className="funnel-kpi-card funnel-kpi-card--cove">
            <span className="funnel-kpi-card__num">{stats.consult_booked_7d}</span>
            <span className="funnel-kpi-card__label">Consults booked</span>
          </div>
          <div className="funnel-kpi-card funnel-kpi-card--green">
            <span className="funnel-kpi-card__num">{stats.closed_won_7d}</span>
            <span className="funnel-kpi-card__label">Closed won</span>
          </div>
          <div className="funnel-kpi-card funnel-kpi-card--red">
            <span className="funnel-kpi-card__num">{stats.closed_lost_7d}</span>
            <span className="funnel-kpi-card__label">Closed lost</span>
          </div>
        </div>
      </section>

      {/* Recent funnel events */}
      <section className="funnel-events-section">
        <h3 className="funnel-events-section__title">Recent events</h3>
        {(events ?? []).length === 0 ? (
          <EmptyState
            variant="empty"
            title="No events recorded"
            message="Product events (chat_started, consult_booked, closed_won, closed_lost) will appear here when posted via the Neptune webhook."
          />
        ) : (
          <div className="funnel-events-list">
            {(events ?? []).slice(0, 20).map((e) => (
              <div key={e.id} className="funnel-event-row">
                <span className={`funnel-event-row__chip funnel-event-row__chip--${e.event_type}`}>
                  {e.event_type.replace(/_/g, " ")}
                </span>
                <span className="funnel-event-row__time">
                  {e.occurred_at?.slice(0, 16)?.replace("T", " ") ?? "—"}
                </span>
                <span className="funnel-event-row__couple">
                  {e.couple_id ? (
                    <a href={`#/dossier/${encodeURIComponent(e.couple_id)}`}>{e.couple_id.slice(0, 12)}…</a>
                  ) : (
                    <em>unresolved</em>
                  )}
                </span>
                <span className="funnel-event-row__journey">
                  {e.journey_stage_before || "—"} → {e.journey_stage_after || "—"}
                </span>
              </div>
            ))}
          </div>
        )}
      </section>
    </div>
  );
}
