import { useEffect, useState } from "react";
import { useActions, useBackfillLocations, useDLQ, useEnrichMissing, useOpsSummary, useRunJanitor, useSources } from "../api/hooks";
import { useToast } from "../components/Toast";

const TODAY_DATE = new Date().toLocaleDateString("en-US", {
  weekday: "long",
  month: "long",
  day: "numeric",
});

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

  // ponytail: animate the budget bar from 0 → actual on mount; the existing
  // width transition on .budget-bar__fill does the easing.
  const [mounted, setMounted] = useState(false);
  useEffect(() => {
    const id = requestAnimationFrame(() => setMounted(true));
    return () => cancelAnimationFrame(id);
  }, []);

  const radarLive = data ? !data.paused && data.running : false;

  return (
    <div className="view view--today">
      <header className="prospects-hero prospects-hero--today">
        <div className="prospects-hero__copy">
          <h2 className="view__title view__title--today">Today</h2>
          <p className="view__subtitle view__subtitle--today">
            {TODAY_DATE} · Your concierge desk. Celebrate first, then everything else.
          </p>
        </div>
        {data && (
          <div className={`radar-pill ${radarLive ? "radar-pill--live" : "radar-pill--paused"}`}>
            <span className="radar-pill__dot" />
            {radarLive ? "Radar Live" : "Radar Paused"}
          </div>
        )}
      </header>

      {error && <div className="empty-state">{(error as Error).message}</div>}
      {isLoading && <div className="empty-state">Loading ops…</div>}

      {data && (
        <>
          <section className="today-queues">
            <h3 className="today-queues__title">Concierge queues</h3>
            <div className="queue-grid">
              <button
                type="button"
                className="queue-card queue-card--celebrate"
                onClick={() => onNavigate("/work?filter=action")}
              >
                <span className="queue-card__top">
                  <span className="queue-card__icon" aria-hidden>🎉</span>
                  <span className="queue-card__n">{data.queue_congratulate ?? 0}</span>
                </span>
                <span className="queue-card__label">Ready to congratulate</span>
                <span className="queue-card__hint">First move · postcard kits</span>
              </button>
              <button type="button" className="queue-card queue-card--detective" onClick={() => onNavigate("/work?filter=pics")}>
                <span className="queue-card__top">
                  <span className="queue-card__icon" aria-hidden>🔍</span>
                  <span className="queue-card__n">{data.queue_detective ?? 0}</span>
                </span>
                <span className="queue-card__label">Needs detective</span>
                <span className="queue-card__hint">Pics, location, identity</span>
              </button>
              <button type="button" className="queue-card queue-card--runway" onClick={() => onNavigate("/work?filter=action")}>
                <span className="queue-card__top">
                  <span className="queue-card__icon" aria-hidden>⏰</span>
                  <span className="queue-card__n">{data.queue_runway_urgent ?? 0}</span>
                </span>
                <span className="queue-card__label">Runway urgent</span>
                <span className="queue-card__hint">Date is close · prioritize</span>
              </button>
              <button type="button" className="queue-card queue-card--risk" onClick={() => onNavigate("/work?filter=action")}>
                <span className="queue-card__top">
                  <span className="queue-card__icon" aria-hidden>⚠️</span>
                  <span className="queue-card__n">{data.queue_risk ?? 0}</span>
                </span>
                <span className="queue-card__label">Relationship risk</span>
                <span className="queue-card__hint">Pause · celebrate only</span>
              </button>
            </div>
          </section>

          <div className="kpi-grid">
            <button type="button" className="kpi" onClick={() => onNavigate("/work?filter=action")}>
              <span className="kpi__value">{data.pending_actions}</span>
              <span className="kpi__label">Pending approvals</span>
            </button>
            <button type="button" className="kpi" onClick={() => onNavigate("/work")}>
              <span className="kpi__value">{data.couples_total}</span>
              <span className="kpi__label">Prospects</span>
              <span className="kpi__sub">
                <span className="kpi__trend kpi__trend--up">↑</span>+{data.couples_24h} in 24h
              </span>
            </button>
            <button type="button" className="kpi" onClick={() => onNavigate("/congratulate")}>
              <span className="kpi__value">{data.kits_ready_to_mail ?? 0}</span>
              <span className="kpi__label">Kits ready to mail</span>
              <span className="kpi__sub">
                <span className="kpi__trend">→</span>{data.kits_mailed ?? 0} mailed
              </span>
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
              <span className="kpi__sub">
                <span className="kpi__trend">→</span>{data.sources_stale} stale
              </span>
            </button>
            <button type="button" className="kpi" onClick={() => onNavigate("/map")}>
              <span className="kpi__value">{data.map_pins}</span>
              <span className="kpi__label">Map pins</span>
            </button>
            <button type="button" className="kpi" onClick={() => onNavigate("/audit")}>
              <span className="kpi__value">{data.funnel_chat_started_7d ?? 0}</span>
              <span className="kpi__label">Chats (7d)</span>
              <span className="kpi__sub">
                <span className="kpi__trend">→</span>
                {data.funnel_consult_booked_7d ?? 0} booked · {Math.round((data.funnel_chat_rate ?? 0) * 100)}% rate
              </span>
            </button>
            <button type="button" className="kpi" onClick={() => onNavigate("/audit")}>
              <span className="kpi__value">{data.funnel_handoffs_issued ?? 0}</span>
              <span className="kpi__label">Handoffs issued</span>
              <span className="kpi__sub">
                <span className="kpi__trend">→</span>Tracked chat links
              </span>
            </button>
          </div>

          <BriefingSection onNavigate={onNavigate} />

          <div className="today-panels">
            <section className="today-panel">
              <h3 className="today-panel__title">
                <span className="today-panel__icon" aria-hidden>📡</span>Provider &amp; budget
              </h3>
              <div className={`budget-bar budget-bar--${budgetTone}`}>
                <div className="budget-bar__fill" style={{ width: `${mounted ? budgetPct : 0}%` }} />
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
              <h3 className="today-panel__title">
                <span className="today-panel__icon" aria-hidden>🧹</span>Data quality actions
              </h3>
              <div className="today-panel__actions today-panel__actions--stacked">
                <button
                  type="button"
                  className="btn btn--primary today-action"
                  disabled={enrich.isPending}
                  onClick={() =>
                    enrich.mutate(15, {
                      onSuccess: (r) => toast.push(`Enriched ${r.succeeded}/${r.attempted} profiles`, r.succeeded ? "ok" : "err"),
                      onError: (e) => toast.push((e as Error).message, "err"),
                    })
                  }
                >
                  <span aria-hidden>🖼️</span>{enrich.isPending ? "Enriching…" : "Enrich missing pics"}
                </button>
                <button
                  type="button"
                  className="btn btn--ghost today-action"
                  disabled={backfill.isPending}
                  onClick={() =>
                    backfill.mutate(100, {
                      onSuccess: (r) => toast.push(`Locations updated ${r.updated}/${r.checked}`, "ok"),
                      onError: (e) => toast.push((e as Error).message, "err"),
                    })
                  }
                >
                  <span aria-hidden>📍</span>{backfill.isPending ? "Backfilling…" : "Backfill map locations"}
                </button>
                <button type="button" className="btn btn--ghost today-action" onClick={() => onNavigate("/work?filter=action")}>
                  <span aria-hidden>📋</span>Open work queue
                </button>
                <button type="button" className="btn btn--ghost today-action" onClick={() => onNavigate("/congratulate")}>
                  <span aria-hidden>✉️</span>Congratulate kits
                </button>
                <button
                  type="button"
                  className="btn btn--ghost today-action"
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
                  <span aria-hidden>🧽</span>{janitor.isPending ? "Cleaning…" : "Run janitor cleanup"}
                </button>
              </div>
            </section>

            <section className="today-panel today-panel--brand">
              <h3 className="today-panel__title">
                <span className="today-panel__icon" aria-hidden>📐</span>Brand rules (Meet Neptune)
              </h3>
              <ul className="today-brand-rules">
                <li>
                  <strong>Celebrate first</strong> — postcard / soft note never mentions prenup on day one
                </li>
                <li>
                  <strong>Both partners</strong> — dual-name copy; alignment before attorneys
                </li>
                <li>
                  <strong>Runway gate</strong> — short wedding dates deprioritize hard invite
                </li>
                <li>
                  <strong>Risk path</strong> — unfollow/state change → pause + concierge only
                </li>
                <li>
                  <strong>Human-in-loop</strong> — model proposes, policy decides, you approve
                </li>
              </ul>
            </section>
          </div>
        </>
      )}
    </div>
  );
}

// BriefingSection shows the top prioritized items needing human attention,
// pulled from existing hooks — no new API needed. Each item has a one-click
// navigate to the right view.
function BriefingSection({ onNavigate }: { onNavigate: (path: string) => void }) {
  const { data: actions } = useActions("pending");
  const { data: dlq } = useDLQ("pending", 5);
  const { data: sources } = useSources(false);

  const pendingActions = (actions ?? []).slice(0, 5);
  const dlqItems = (dlq ?? []).slice(0, 3);
  const staleSources = (sources ?? []).filter((s) => !s.active).slice(0, 3);

  const hasItems = pendingActions.length > 0 || dlqItems.length > 0 || staleSources.length > 0;
  if (!hasItems) return null;

  return (
    <section className="briefing">
      <h3 className="briefing__title">
        <span className="briefing__icon" aria-hidden>📋</span>Concierge briefing
      </h3>
      <p className="briefing__sub">What deserves attention right now.</p>
      <div className="briefing__grid">
        {pendingActions.length > 0 && (
          <div className="briefing__col">
            <h4 className="briefing__col-title">
              Pending approvals ({pendingActions.length})
            </h4>
            <ul className="briefing__list">
              {pendingActions.map((a) => (
                <li key={a.id} className="briefing__item briefing__item--action" onClick={() => onNavigate("/work?filter=action")}>
                  <span className="briefing__item-type">{a.action_type}</span>
                  <span className="briefing__item-meta">{a.id.slice(0, 12)}…</span>
                </li>
              ))}
            </ul>
          </div>
        )}
        {dlqItems.length > 0 && (
          <div className="briefing__col">
            <h4 className="briefing__col-title">
              Failed signals ({dlqItems.length})
            </h4>
            <ul className="briefing__list">
              {dlqItems.map((d) => (
                <li key={d.id} className="briefing__item briefing__item--dlq" onClick={() => onNavigate("/dlq")}>
                  <span className="briefing__item-type">{d.source}</span>
                  <span className="briefing__item-meta">{d.error_message.slice(0, 60)}</span>
                </li>
              ))}
            </ul>
          </div>
        )}
        {staleSources.length > 0 && (
          <div className="briefing__col">
            <h4 className="briefing__col-title">
              Inactive sources ({staleSources.length})
            </h4>
            <ul className="briefing__list">
              {staleSources.map((s) => (
                <li key={s.id} className="briefing__item briefing__item--source" onClick={() => onNavigate("/sources")}>
                  <span className="briefing__item-type">@{s.handle}</span>
                  <span className="briefing__item-meta">{s.source_class}</span>
                </li>
              ))}
            </ul>
          </div>
        )}
      </div>
    </section>
  );
}
