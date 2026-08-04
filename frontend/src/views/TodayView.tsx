import { useEffect, useState } from "react";
import { useActions, useBackfillLocations, useDLQ, useEnrichMissing, useOpsSummary, useRunJanitor } from "../api/hooks";
import type { OpsSummary } from "../api/types";
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

  // Animate the budget bar from 0 → actual on mount.
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
                  <span className="queue-card__n">{data.queue_congratulate ?? 0}</span>
                </span>
                <span className="queue-card__label">Ready to congratulate</span>
                <span className="queue-card__hint">First move · postcard kits</span>
              </button>
              <button type="button" className="queue-card queue-card--detective" onClick={() => onNavigate("/work?filter=pics")}>
                <span className="queue-card__top">
                  <span className="queue-card__n">{data.queue_detective ?? 0}</span>
                </span>
                <span className="queue-card__label">Needs detective</span>
                <span className="queue-card__hint">Pics, location, identity</span>
              </button>
              <button type="button" className="queue-card queue-card--runway" onClick={() => onNavigate("/work?filter=action")}>
                <span className="queue-card__top">
                  <span className="queue-card__n">{data.queue_runway_urgent ?? 0}</span>
                </span>
                <span className="queue-card__label">Runway urgent</span>
                <span className="queue-card__hint">Date is close · prioritize</span>
              </button>
              <button type="button" className="queue-card queue-card--risk" onClick={() => onNavigate("/work?filter=action")}>
                <span className="queue-card__top">
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
                <span className="kpi__trend kpi__trend--up">+</span>{data.couples_24h} in 24h
              </span>
            </button>
            <button type="button" className="kpi" onClick={() => onNavigate("/congratulate")}>
              <span className="kpi__value">{data.kits_ready_to_mail ?? 0}</span>
              <span className="kpi__label">Kits ready to mail</span>
              <span className="kpi__sub">
                {data.kits_mailed ?? 0} mailed
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
                {data.sources_stale} stale
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
                {data.funnel_consult_booked_7d ?? 0} booked · {Math.round((data.funnel_chat_rate ?? 0) * 100)}% rate
              </span>
            </button>
            <button type="button" className="kpi" onClick={() => onNavigate("/audit")}>
              <span className="kpi__value">{data.funnel_handoffs_issued ?? 0}</span>
              <span className="kpi__label">Handoffs issued</span>
              <span className="kpi__sub">Tracked chat links</span>
            </button>
          </div>

          <OrganismStrip data={data} onNavigate={onNavigate} />

          <BriefingSection onNavigate={onNavigate} />

          <div className="today-panels">
            <section className="today-panel">
              <h3 className="today-panel__title">Provider &amp; budget</h3>
              <div className={`budget-bar budget-bar--${budgetTone}`}>
                <div
                  className="budget-bar__fill"
                  style={{ ["--budget-pct" as string]: mounted ? budgetPct / 100 : 0 }}
                />
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
              <h3 className="today-panel__title">Data quality</h3>
              <div className="today-panel__actions">
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
                  {enrich.isPending ? "Enriching…" : "Enrich pics"}
                </button>
                <button
                  type="button"
                  className="btn today-action"
                  disabled={backfill.isPending}
                  onClick={() =>
                    backfill.mutate(100, {
                      onSuccess: (r) => toast.push(`Locations updated ${r.updated}/${r.checked}`, "ok"),
                      onError: (e) => toast.push((e as Error).message, "err"),
                    })
                  }
                >
                  {backfill.isPending ? "Backfilling…" : "Backfill locations"}
                </button>
                <button
                  type="button"
                  className="btn today-action"
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
                  {janitor.isPending ? "Cleaning…" : "Run janitor"}
                </button>
                <button type="button" className="btn btn--ghost today-action" onClick={() => onNavigate("/work?filter=action")}>
                  Open work queue
                </button>
              </div>
            </section>

            <section className="today-panel today-panel--brand">
              <h3 className="today-panel__title">Brand rules</h3>
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

// ponytail: derive strip from ops (already loaded) — no /api/organism on home.
function OrganismStrip({ data, onNavigate }: { data: OpsSummary; onNavigate: (path: string) => void }) {
  const celebrate = data.queue_congratulate ?? 0;
  const detective = data.queue_detective ?? 0;
  const risk = data.queue_risk ?? 0;
  const chats = data.funnel_chat_started_7d ?? 0;
  const booked = data.funnel_consult_booked_7d ?? 0;
  const won = data.funnel_closed_won_7d ?? 0;
  const mailed = data.kits_mailed ?? 0;
  const parts: string[] = [];
  if (celebrate > 0) parts.push(`${celebrate} celebrate-ready`);
  if (risk > 0) parts.push(`${risk} risk-pause`);
  if (won > 0) parts.push(`${won} closed-won (7d)`);
  const headline = parts.length ? parts.join(" · ") : "Radar quiet — sources and budget healthy";
  const agents = [
    { id: "scout", name: "Scout", val: String(data.results_used_today ?? 0), label: "signals today", status: "live" },
    { id: "detective", name: "Detective", val: String(detective), label: "needs detective", status: detective > 0 ? "live" : "idle" },
    { id: "concierge", name: "Concierge", val: String(celebrate), label: "ready to congratulate", status: celebrate > 0 ? "live" : "idle" },
    { id: "risk", name: "Risk Sentinel", val: String(risk), label: "risk queue", status: risk > 0 ? "live" : "idle" },
  ];
  return (
    <section className="organism-strip">
      <header className="organism-strip__head">
        <div>
          <h3 className="organism-strip__title">Growth organism</h3>
          <p className="organism-strip__headline">{headline}</p>
        </div>
        <button type="button" className="btn btn--ghost btn--sm" onClick={() => onNavigate("/organism")}>
          Full organism
        </button>
      </header>
      <div className="organism-strip__swarm">
        {agents.map((a) => (
          <div key={a.id} className={`organism-strip__agent organism-strip__agent--${a.status}`}>
            <span className="organism-strip__agent-name">{a.name}</span>
            <span className="organism-strip__agent-val">{a.val}</span>
            <span className="organism-strip__agent-label">{a.label}</span>
          </div>
        ))}
      </div>
      <div className="organism-strip__yield">
        <span>{chats} chats</span>
        <span>{booked} booked</span>
        <span>{won} won</span>
        <span>{mailed} mailed</span>
        <span className="organism-strip__risk">{risk} risk pause</span>
      </div>
      <p className="organism-strip__rule">Celebrate first. Never pitch prenup on day one. Never pitch on relationship-risk signals.</p>
    </section>
  );
}

function BriefingSection({ onNavigate }: { onNavigate: (path: string) => void }) {
  // ponytail: limit=8 — full pending list was ~100KB+ for a 5-item strip.
  const { data: actions } = useActions("pending", 8);
  const { data: dlq } = useDLQ("pending", 5);

  const pendingActions = (actions ?? []).slice(0, 5);
  const dlqItems = (dlq ?? []).slice(0, 3);

  const hasItems = pendingActions.length > 0 || dlqItems.length > 0;
  if (!hasItems) return null;

  return (
    <section className="briefing">
      <h3 className="briefing__title">Concierge briefing</h3>
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
      </div>
    </section>
  );
}
