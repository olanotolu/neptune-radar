import { useMemo } from "react";
import {
  useDLQ,
  useFunnelStats,
  useIngestStatus,
  useKits,
  useNotifications,
  useOpsSummary,
  useProspectBoard,
  useSignals,
} from "../api/hooks";
import type { ProspectCard, Signal } from "../api/types";
import { EmptyState } from "../components/EmptyState";
import { LoadingState } from "../components/LoadingState";

// ponytail: command center inherits each hook's native poll cadence rather
// than overriding shared hook intervals (that would change every other view).
// health/funnel land 15-60s, feed/priority 30-45s — close to the 15/30 spec.

function relativeTime(iso: string): string {
  const diff = Date.now() - new Date(iso).getTime();
  const s = Math.floor(diff / 1000);
  if (s < 60) return `${s}s`;
  const m = Math.floor(s / 60);
  if (m < 60) return `${m}m`;
  const h = Math.floor(m / 60);
  if (h < 24) return `${h}h`;
  return `${Math.floor(h / 24)}d`;
}

// ponytail: observation_type is free-form; categorize by substring (same trick
// as FeedView) so the chip stays useful whatever the backend emits.
function signalChip(s: Signal): { label: string; cls: string } {
  const t = s.observation_type.toLowerCase();
  if (t.includes("engag") || t.includes("post")) return { label: "ENG", cls: "cc-chip--eng" };
  if (t.includes("tag")) return { label: "TAG", cls: "cc-chip--tag" };
  if (t.includes("mention")) return { label: "MEN", cls: "cc-chip--men" };
  if (t.includes("vendor") || s.monitor.startsWith("vendor:")) return { label: "VND", cls: "cc-chip--vnd" };
  return { label: "SIG", cls: "cc-chip--sig" };
}

type DotTone = "live" | "idle" | "paused" | "warn" | "bad";

function HealthCard({
  label,
  value,
  tone,
  onClick,
}: {
  label: string;
  value: React.ReactNode;
  tone: DotTone;
  onClick?: () => void;
}) {
  return (
    <button type="button" className="cc-card cc-card--health" onClick={onClick} disabled={!onClick}>
      <span className={`cc-card__dot cc-card__dot--${tone}`} aria-hidden />
      <span className="cc-card__label">{label}</span>
      <span className="cc-card__value">{value}</span>
    </button>
  );
}

function KpiCard({
  label,
  value,
  delta,
  deltaTone,
  onClick,
}: {
  label: string;
  value: React.ReactNode;
  delta?: string;
  deltaTone?: "up" | "down" | "flat";
  onClick?: () => void;
}) {
  return (
    <button type="button" className="cc-card" onClick={onClick} disabled={!onClick}>
      <span className="cc-card__label">{label}</span>
      <span className="cc-card__value">{value}</span>
      {delta && <span className={`cc-card__delta cc-card__delta--${deltaTone ?? "flat"}`}>{delta}</span>}
    </button>
  );
}

export function CommandCenterView({ onNavigate }: { onNavigate: (path: string) => void }) {
  const { data: ingest, isLoading: ingestLoading } = useIngestStatus();
  const { data: ops, isLoading: opsLoading } = useOpsSummary();
  const { data: funnel } = useFunnelStats();
  // ponytail: DLQ count is min(limit, actual) — no count endpoint exists.
  const { data: dlq } = useDLQ("pending", 50);
  // ponytail: OpsSummary exposes no unread-notification field; use the inbox hook.
  const { data: notifications } = useNotifications(true);
  const { data: signals } = useSignals();
  // ponytail: useCouples returns CoupleSummary (no priority/handles/stage);
  // useProspectBoard has all three — sort client-side by neptune_rank.
  const { data: board } = useProspectBoard();
  // ponytail: one kits fetch feeds QR-scan total + follow-up sent/due.
  const { data: kits } = useKits();

  const kitAgg = useMemo(() => {
    const list = kits ?? [];
    const scans = list.reduce((n, k) => n + (k.qr_scan_count ?? 0), 0);
    const mailed = list.filter((k) => k.status === "mailed").length;
    const followSent = list.filter((k) => k.follow_up_sent_at).length;
    const now = Date.now();
    const followDue = list.filter(
      (k) => k.follow_up_at && !k.follow_up_sent_at && new Date(k.follow_up_at).getTime() <= now,
    ).length;
    return { scans, mailed, followSent, followDue };
  }, [kits]);

  const priorityCouples = useMemo<ProspectCard[]>(() => {
    const cards = board ? (Object.values(board.cards) as ProspectCard[][]).flat() : [];
    return cards
      .filter((c) => c.neptune_rank != null)
      .sort((a, b) => (b.neptune_rank ?? 0) - (a.neptune_rank ?? 0))
      .slice(0, 5);
  }, [board]);

  if (ingestLoading && opsLoading) {
    return <LoadingState variant="skeleton" message="Loading command center…" />;
  }

  const running = ingest?.running ?? false;
  const paused = ingest?.paused ?? false;
  const ingestTone: DotTone = paused ? "paused" : running ? "live" : "idle";
  const ingestLabel = paused ? "Paused" : running ? "Running" : "Idle";

  // ponytail: no explicit DB-health metric in OpsSummary; a successful ops
  // fetch means the DB is reachable — that IS the health signal here.
  const dbTone: DotTone = ops ? "live" : "idle";
  const dlqN = dlq?.length ?? 0;
  const dlqTone: DotTone = dlqN > 0 ? "bad" : "live";
  const pendingN = ops?.pending_actions ?? 0;
  const pendingTone: DotTone = pendingN > 0 ? "warn" : "live";
  const unreadN = notifications?.length ?? 0;
  const unreadTone: DotTone = unreadN > 0 ? "warn" : "live";

  const mailed = ops?.kits_mailed ?? kitAgg.mailed;
  const scanRate = mailed > 0 ? Math.round((kitAgg.scans / mailed) * 100) : 0;
  const couples24h = ops?.couples_24h ?? 0;

  return (
    <div className="view view--command">
      <header className="view__head">
        <h2>Command Center</h2>
        <p className="trust-panel__sub">One glance. Every system, every funnel step, every priority.</p>
      </header>

      {/* Top row: system health strip */}
      <section className="cc-section">
        <h3 className="cc-section__title">System health</h3>
        <div className="cc-grid cc-grid--health">
          <HealthCard label="Ingest" value={ingestLabel} tone={ingestTone} onClick={() => onNavigate("/ops")} />
          <HealthCard label="Database" value={ops ? "Online" : "—"} tone={dbTone} onClick={() => onNavigate("/ops")} />
          <HealthCard label="DLQ pending" value={dlqN} tone={dlqTone} onClick={() => onNavigate("/dlq")} />
          <HealthCard
            label="Pending actions"
            value={pendingN}
            tone={pendingTone}
            onClick={() => onNavigate("/work?filter=action")}
          />
          <HealthCard label="Unread alerts" value={unreadN} tone={unreadTone} onClick={() => onNavigate("/ops")} />
        </div>
      </section>

      {/* Middle row: funnel KPIs */}
      <section className="cc-section">
        <h3 className="cc-section__title">Funnel</h3>
        <div className="cc-grid cc-grid--funnel">
          <KpiCard
            label="Couples discovered"
            value={ops?.couples_total ?? 0}
            delta={couples24h > 0 ? `+${couples24h} 24h` : undefined}
            deltaTone="up"
            onClick={() => onNavigate("/work")}
          />
          <KpiCard label="Postcards mailed" value={mailed} onClick={() => onNavigate("/congratulate")} />
          <KpiCard
            label="QR scans"
            value={kitAgg.scans}
            delta={`${scanRate}% conv`}
            deltaTone="flat"
            onClick={() => onNavigate("/congratulate")}
          />
          <KpiCard
            label="Chats started"
            value={funnel?.chat_started_7d ?? 0}
            delta="7d"
            onClick={() => onNavigate("/funnel")}
          />
          <KpiCard
            label="Consults booked"
            value={funnel?.consult_booked_7d ?? 0}
            delta="7d"
            onClick={() => onNavigate("/funnel")}
          />
          <KpiCard
            label="Closed won"
            value={funnel?.closed_won_7d ?? 0}
            delta="7d"
            deltaTone="up"
            onClick={() => onNavigate("/funnel")}
          />
        </div>
      </section>

      {/* Middle row: pipeline status */}
      <section className="cc-section">
        <h3 className="cc-section__title">Pipeline</h3>
        <div className="cc-grid cc-grid--pipeline">
          {/* ponytail: no "active-only" filter in OpsSummary; couples_total is the closest proxy. */}
          <KpiCard label="Active couples" value={ops?.couples_total ?? 0} onClick={() => onNavigate("/work")} />
          <KpiCard
            label="Kits ready to mail"
            value={ops?.kits_ready_to_mail ?? 0}
            onClick={() => onNavigate("/congratulate")}
          />
          <KpiCard label="Kits mailed" value={mailed} onClick={() => onNavigate("/congratulate")} />
          <KpiCard label="Follow-ups sent" value={kitAgg.followSent} onClick={() => onNavigate("/congratulate")} />
          <KpiCard
            label="Follow-ups due"
            value={kitAgg.followDue}
            deltaTone={kitAgg.followDue > 0 ? "down" : "flat"}
            onClick={() => onNavigate("/congratulate")}
          />
        </div>
      </section>

      {/* Bottom row: recent activity + top priority */}
      <div className="cc-bottom">
        <section className="cc-section cc-section--feed">
          <h3 className="cc-section__title">Recent activity</h3>
          {(signals ?? []).length === 0 ? (
            <EmptyState title="No signals yet" message="The radar is quiet — signals will appear as sources are scanned." />
          ) : (
            <ul className="cc-feed">
              {(signals ?? []).slice(0, 10).map((s) => {
                const chip = signalChip(s);
                return (
                  <li key={s.id} className="cc-feed__row">
                    <span className="cc-feed__time">{relativeTime(s.observed_at)}</span>
                    <span className={`cc-chip ${chip.cls}`}>{chip.label}</span>
                    <span className="cc-feed__handle">@{s.handle}</span>
                    <span className="cc-feed__summary">{s.summary}</span>
                  </li>
                );
              })}
            </ul>
          )}
        </section>

        <section className="cc-section cc-section--priority">
          <h3 className="cc-section__title">Top priority couples</h3>
          {priorityCouples.length === 0 ? (
            <EmptyState title="No ranked couples" message="Prospects with a Neptune rank will surface here once scored." />
          ) : (
            <ul className="cc-priority">
              {priorityCouples.map((c) => (
                <li
                  key={c.couple_id}
                  className="cc-priority__row"
                  onClick={() => onNavigate(`/dossier?couple=${encodeURIComponent(c.couple_id)}`)}
                >
                  <span className="cc-priority__names">
                    {c.person_a_label} &amp; {c.person_b_label}
                  </span>
                  <span className="cc-priority__handles">
                    {c.handle_a ? `@${c.handle_a}` : "—"}
                    {c.handle_b ? ` @${c.handle_b}` : ""}
                  </span>
                  <span className="cc-priority__score">{c.neptune_rank}</span>
                  {c.stage && <span className="cc-chip cc-chip--stage">{c.stage}</span>}
                </li>
              ))}
            </ul>
          )}
        </section>
      </div>
    </div>
  );
}
