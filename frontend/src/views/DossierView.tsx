import { useEffect, useMemo, useRef, useState } from "react";
import {
  useApproveAction,
  useBuildCongratulateKit,
  useCoupleDossier,
  useCreateHandoff,
  useIgnoreAction,
  useSetJourneyStage,
  useSuppressCouple,
} from "../api/hooks";
import { mediaURL } from "../api/media";
import type { BrandAction, CoupleDossier, DossierEvidence } from "../api/types";
import { useToast } from "../components/Toast";
import { EmptyState } from "../components/EmptyState";
import { LoadingState } from "../components/LoadingState";

function pct(n: number): string {
  return `${Math.round((n || 0) * 100)}%`;
}

function Avatar({ url, label, size = 64 }: { url?: string; label: string; size?: number }) {
  const style = { width: size, height: size };
  if (url) {
    return <img className="dossier__avatar" style={style} src={mediaURL(url)} alt="" referrerPolicy="no-referrer" />;
  }
  return (
    <span className="dossier__avatar dossier__avatar--fallback" style={style}>
      {(label || "?")[0]?.toUpperCase()}
    </span>
  );
}

/* ---------- Neptune Rank gauge (animated circular) ---------- */

function RankGauge({ rank }: { rank: number }) {
  const p = Math.round((rank || 0) * 100);
  const tone = p >= 70 ? "hot" : p >= 45 ? "warm" : "cool";
  const r = 52;
  const circ = 2 * Math.PI * r;
  const [shown, setShown] = useState(0);
  useEffect(() => {
    const t = setTimeout(() => setShown(p), 120);
    return () => clearTimeout(t);
  }, [p]);
  const offset = circ - (shown / 100) * circ;
  return (
    <div
      className={`dossier-gauge dossier-gauge--${tone}`}
      title="Neptune rank = engagement × ICP × runway × deliverability"
    >
      <svg width="120" height="120" viewBox="0 0 120 120">
        <circle className="dossier-gauge__track" cx="60" cy="60" r={r} fill="none" strokeWidth="8" />
        <circle
          className="dossier-gauge__fill"
          cx="60"
          cy="60"
          r={r}
          fill="none"
          strokeWidth="8"
          strokeLinecap="round"
          strokeDasharray={circ}
          strokeDashoffset={offset}
          transform="rotate(-90 60 60)"
        />
      </svg>
      <div className="dossier-gauge__center">
        <span className="dossier-gauge__n">{p}</span>
        <span className="dossier-gauge__l">Neptune Rank</span>
      </div>
    </div>
  );
}

/* ---------- Metric card ---------- */

function MetricCard({
  icon,
  label,
  value,
  sub,
  accent,
}: {
  icon: string;
  label: string;
  value: string;
  sub: string;
  accent: string;
}) {
  return (
    <div className={`dossier-mcard dossier-mcard--${accent}`}>
      <span className="dossier-mcard__icon">{icon}</span>
      <span className="dossier-mcard__label">{label}</span>
      <strong className="dossier-mcard__value">{value}</strong>
      <span className="dossier-mcard__sub">{sub}</span>
    </div>
  );
}

/* ---------- Evidence timeline ---------- */

function evidenceTone(e: DossierEvidence): "pos" | "neg" | "neutral" {
  if (e.points > 0) return "pos";
  if (e.points < 0) return "neg";
  return "neutral";
}

function EvidenceTimeline({ items }: { items: DossierEvidence[] }) {
  const sorted = useMemo(
    () => [...items].sort((a, b) => (b.created_at || "").localeCompare(a.created_at || "") || b.points - a.points),
    [items],
  );
  if (sorted.length === 0) return <p className="work-drawer__muted">No evidence rows yet.</p>;
  const maxW = Math.max(100, ...sorted.map((e) => Math.abs(e.weight)));
  return (
    <ol className="dossier-timeline">
      {sorted.map((e) => {
        const tone = evidenceTone(e);
        const w = Math.min(100, Math.round((Math.abs(e.weight) / maxW) * 100));
        return (
          <li key={e.id} className={`dossier-timeline__item dossier-timeline__item--${tone}`}>
            <span className="dossier-timeline__dot" />
            <div className="dossier-timeline__body">
              <div className="dossier-timeline__head">
                <span className="dossier-timeline__chip">{e.kind}</span>
                <span className="dossier-timeline__pts">
                  {e.points > 0 ? "+" : ""}
                  {e.points} pts
                </span>
                {e.created_at && <time className="dossier-timeline__time">{e.created_at.slice(0, 10)}</time>}
              </div>
              <p className="dossier-timeline__desc">{e.description}</p>
              <div className="dossier-timeline__weight">
                <span className="dossier-timeline__weight-bar" style={{ width: `${w}%` }} />
              </div>
            </div>
          </li>
        );
      })}
    </ol>
  );
}

/* ---------- Journey stepper ---------- */

const STAGES = ["detected", "approved", "congratulated", "invited", "in_chat", "booked", "closed_won"] as const;

function JourneyStepper({
  current,
  busy,
  onSet,
}: {
  current: string;
  busy: boolean;
  onSet: (stage: string) => void;
}) {
  const idx = STAGES.indexOf(current as (typeof STAGES)[number]);
  return (
    <ol className="dossier-stepper">
      {STAGES.map((s, i) => {
        const done = idx >= 0 && i < idx;
        const isCurrent = s === current;
        return (
          <li
            key={s}
            className={`dossier-stepper__step${isCurrent ? " dossier-stepper__step--current" : ""}${done ? " dossier-stepper__step--done" : ""}`}
          >
            <button
              type="button"
              className="dossier-stepper__btn"
              disabled={busy}
              title={`Set journey → ${s.replace(/_/g, " ")}`}
              onClick={() => {
                if (s === current) return;
                if (!confirm(`Move journey to ${s.replace(/_/g, " ")}?`)) return;
                onSet(s);
              }}
            >
              <span className="dossier-stepper__node">{done ? "✓" : i + 1}</span>
              <span className="dossier-stepper__label">{s.replace(/_/g, " ")}</span>
            </button>
          </li>
        );
      })}
    </ol>
  );
}

/* ---------- Brand action panel ---------- */

function BrandActionButton({
  action,
  busy,
  onRun,
}: {
  action: BrandAction;
  busy: boolean;
  onRun: (id: string) => void;
}) {
  const icon =
    action.tone === "celebrate" ? "🎉" : action.tone === "soft_invite" ? "✉" : action.tone === "risk" ? "⚠" : "•";
  return (
    <button
      type="button"
      className={`dossier-action dossier-action--${action.tone}${!action.allowed ? " dossier-action--blocked" : ""}`}
      disabled={busy || !action.allowed}
      title={action.block_reason || action.body}
      onClick={() => onRun(action.id)}
    >
      <span className="dossier-action__icon">{icon}</span>
      <span className="dossier-action__text">
        <strong>{action.title}</strong>
        <span>{action.allowed ? action.body : action.block_reason || action.body}</span>
      </span>
    </button>
  );
}

/* ---------- Notes (local state, autosave indicator) ---------- */

function NotesSection({ coupleId }: { coupleId: string }) {
  const key = `dossier-notes-${coupleId}`;
  const [text, setText] = useState(() => localStorage.getItem(key) || "");
  const [savedAt, setSavedAt] = useState<string | null>(() => localStorage.getItem(`${key}-ts`));
  const taRef = useRef<HTMLTextAreaElement>(null);

  const save = () => {
    localStorage.setItem(key, text);
    const ts = new Date().toISOString();
    localStorage.setItem(`${key}-ts`, ts);
    setSavedAt(ts);
  };

  return (
    <section className="dossier__section dossier-notes">
      <h3>Analyst notes</h3>
      <textarea
        ref={taRef}
        className="dossier-notes__ta"
        placeholder="Private notes for the concierge team — not sent to the couple."
        value={text}
        onChange={(e) => setText(e.target.value)}
        onBlur={save}
        rows={4}
      />
      <span className="dossier-notes__saved">
        {savedAt ? `saved ${new Date(savedAt).toLocaleString()}` : "unsaved"}
      </span>
    </section>
  );
}

export function DossierView({
  coupleId,
  onClose,
  onCongratulate,
}: {
  coupleId: string;
  onClose?: () => void;
  onCongratulate?: (coupleId: string) => void;
}) {
  const { data, error, isLoading, refetch } = useCoupleDossier(coupleId);
  const approve = useApproveAction();
  const ignore = useIgnoreAction();
  const suppress = useSuppressCouple();
  const buildKit = useBuildCongratulateKit();
  const handoff = useCreateHandoff();
  const journey = useSetJourneyStage();
  const toast = useToast();

  const busy =
    approve.isPending ||
    ignore.isPending ||
    suppress.isPending ||
    buildKit.isPending ||
    handoff.isPending ||
    journey.isPending;

  const runAction = (id: string) => {
    if (!data) return;
    switch (id) {
      case "congratulate":
        buildKit.mutate(coupleId, {
          onSuccess: (kit) => {
            toast.push("Congratulate kit ready — celebration only", "ok");
            onCongratulate?.(kit.couple_id);
            refetch();
          },
          onError: (e) => toast.push((e as Error).message, "err"),
        });
        break;
      case "soft_invite":
        handoff.mutate(coupleId, {
          onSuccess: (r) => {
            toast.push("Tracked chat handoff created", "ok");
            void navigator.clipboard?.writeText(r.handoff_url);
            refetch();
          },
          onError: (e) => toast.push((e as Error).message, "err"),
        });
        break;
      case "approve_pending":
        if (!data.pending_action_id) return;
        approve.mutate(data.pending_action_id, {
          onSuccess: () => {
            toast.push("Approved — journey → approved", "ok");
            refetch();
          },
          onError: (e) => toast.push((e as Error).message, "err"),
        });
        break;
      case "not_a_couple":
        if (!confirm("Permanently suppress — not a real couple?")) return;
        suppress.mutate(
          { id: coupleId, reason: "not_a_couple" },
          {
            onSuccess: () => {
              toast.push("Suppressed", "ok");
              onClose?.();
            },
            onError: (e) => toast.push((e as Error).message, "err"),
          },
        );
        break;
      case "concierge_only":
        journey.mutate(
          { coupleId, stage: "do_not_contact" },
          {
            onSuccess: () => toast.push("Marked do-not-contact (risk path)", "info"),
            onError: (e) => toast.push((e as Error).message, "err"),
          },
        );
        break;
      default:
        break;
    }
  };

  const setStage = (stage: string) =>
    journey.mutate(
      { coupleId, stage },
      {
        onSuccess: () => toast.push(`Journey → ${stage}`, "ok"),
        onError: (e) => toast.push((e as Error).message, "err"),
      },
    );

  if (isLoading) return <LoadingState variant="spinner" message="Loading dossier…" />;
  if (error) return <EmptyState variant="warning" icon="⚠" title="Dossier unavailable" message={(error as Error).message} />;
  if (!data) return <EmptyState variant="empty" icon="📇" title="No dossier" message="No dossier data for this couple." />;

  const nameA = data.person_a_name || data.handle_a || "A";
  const nameB = data.person_b_name || data.handle_b || "B";

  // "Detected X days ago" — earliest audit event or evidence created_at.
  const earliest = useMemo(() => {
    const ts = [
      ...(data.audit_trail || []).map((a) => a.created_at),
      ...(data.evidence || []).map((e) => e.created_at || ""),
    ].filter(Boolean);
    if (ts.length === 0) return null;
    return ts.sort()[0];
  }, [data.audit_trail, data.evidence]);
  const detectedLabel = earliest
    ? `Detected ${Math.max(0, Math.round((Date.now() - new Date(earliest).getTime()) / 86400000))} days ago`
    : null;

  return (
    <div className="dossier dossier--premium">
      <header className="dossier__hero dossier__hero--premium">
        <div className="dossier__hero-bg" />
        <div className="dossier__hero-content">
          <div className="dossier__faces dossier__faces--lg">
            <Avatar url={data.profile_pic_a} label={nameA} />
            <Avatar url={data.profile_pic_b} label={nameB} />
          </div>
          <div className="dossier__identity">
            <h2 className="dossier__title">
              {nameA} <span className="dossier__amp">&</span> {nameB}
            </h2>
            <p className="dossier__handles">
              {data.handle_a && <span>@{data.handle_a}</span>}
              {data.handle_a && data.handle_b ? " · " : ""}
              {data.handle_b && <span>@{data.handle_b}</span>}
              {(data.city || data.region) && (
                <span className="dossier__loc">
                  {" · "}
                  <span className="dossier__pin" aria-hidden>📍</span>
                  {data.city}
                  {data.region ? `, ${data.region}` : ""}
                </span>
              )}
              {detectedLabel && <span className="dossier__detected"> · {detectedLabel}</span>}
            </p>
            <div className="dossier__chips">
              {data.stage && <span className="pcard__chip pcard__chip--stage">{data.stage.replace(/_/g, " ")}</span>}
              {data.journey_stage && <span className="pcard__chip">journey: {data.journey_stage}</span>}
              {data.automation_paused && <span className="pcard__chip pcard__chip--action">paused</span>}
              {data.pending_action_type && (
                <span className="pcard__chip pcard__chip--action">{data.pending_action_type.replace(/_/g, " ")}</span>
              )}
              {data.has_case && <span className="pcard__chip pcard__chip--case">case open</span>}
              {(data.icp?.labels || []).slice(0, 5).map((l) => (
                <span key={l} className="pcard__chip pcard__chip--icp">
                  {l}
                </span>
              ))}
            </div>
            <p className="dossier__tagline">Everything we know about this couple.</p>
          </div>
          <RankGauge rank={data.neptune_rank} />
          {onClose && (
            <button type="button" className="btn btn--ghost btn--sm dossier__close" onClick={onClose}>
              Close
            </button>
          )}
        </div>
      </header>

      <div className="dossier__metrics dossier__metrics--premium">
        <MetricCard
          icon="♥"
          label="Engagement"
          value={pct(data.engagement_score)}
          sub="Did this happen?"
          accent="cove"
        />
        <MetricCard
          icon="⚭"
          label="Partner Fit"
          value={pct(data.partner_score || 0)}
          sub="Right two people?"
          accent="mesa"
        />
        <MetricCard
          icon="◎"
          label="ICP Match"
          value={pct(data.icp?.score || 0)}
          sub={`Market: ${data.icp?.market_label || "unknown"} (${pct(data.icp?.market_priority || 0)})`}
          accent="green"
        />
        <MetricCard
          icon="⏳"
          label="Runway"
          value={data.runway_label || "Unknown"}
          sub={
            data.runway?.raw
              ? `"${data.runway.raw}" · ${data.runway.source || "inferred"}`
              : "No date extracted yet"
          }
          accent={data.runway?.band || "unknown"}
        />
      </div>

      <section className="dossier__section dossier__section--why">
        <h3>Why this couple · now</h3>
        <ul className="dossier-why">
          {(data.why_now || []).map((w) => (
            <li key={w}>{w}</li>
          ))}
        </ul>
      </section>

      <section className="dossier__section">
        <h3>Evidence ledger</h3>
        <p className="dossier__hint">Points timeline — model cannot invent rows. Ads/styled shoots show as −50.</p>
        <EvidenceTimeline items={data.evidence || []} />
      </section>

      <div className="dossier__grid">
        <section className="dossier__section">
          <h3>Identity (no face ID)</h3>
          <ul className="dossier-identity">
            {(data.identity || []).map((i) => (
              <li key={i.kind + i.description} className={`dossier-identity__item dossier-identity--${i.strength}`}>
                <strong>{i.strength}</strong>
                <span>{i.description}</span>
              </li>
            ))}
          </ul>
          {(data.bio_a || data.bio_b) && (
            <div className="dossier-bios">
              {data.bio_a && (
                <p>
                  <strong>@{data.handle_a || "a"}:</strong> {data.bio_a}
                </p>
              )}
              {data.bio_b && (
                <p>
                  <strong>@{data.handle_b || "b"}:</strong> {data.bio_b}
                </p>
              )}
            </div>
          )}
        </section>

        <section className="dossier__section">
          <h3>Recent signals</h3>
          <div className="dossier-posts">
            {(data.observations || []).slice(0, 6).map((p) => (
              <article key={p.id} className="dossier-post">
                {p.image_url && (
                  <img src={mediaURL(p.image_url)} alt="" referrerPolicy="no-referrer" className="dossier-post__img" />
                )}
                <div>
                  <p className="dossier-post__cap">{p.caption || "(no caption)"}</p>
                  <p className="dossier-post__meta">
                    {p.source_handle && `@${p.source_handle} · `}
                    {p.observed_at?.slice(0, 10)}
                    {p.post_url && (
                      <>
                        {" · "}
                        <a href={p.post_url} target="_blank" rel="noreferrer">
                          post
                        </a>
                      </>
                    )}
                  </p>
                </div>
              </article>
            ))}
            {(data.observations || []).length === 0 && <p className="work-drawer__muted">No posts linked yet.</p>}
          </div>
        </section>
      </div>

      <section className="dossier__section">
        <h3>Journey</h3>
        <p className="dossier__hint">Celebrate first. Soft invite second. Prenup chat only after both partners can opt in. Click a stage to advance.</p>
        <JourneyStepper current={data.journey_stage} busy={busy} onSet={setStage} />
        <div className="dossier-journey__advance">
          {["closed_lost", "do_not_contact"].map((st) => (
            <button
              key={st}
              type="button"
              className="btn btn--ghost btn--sm"
              disabled={busy}
              onClick={() => setStage(st)}
            >
              {st.replace(/_/g, " ")}
            </button>
          ))}
        </div>
      </section>

      <section className="dossier__section">
        <h3>Recommended actions</h3>
        <div className="dossier-actions dossier-actions--row">
          {(data.brand_actions || []).map((a) => (
            <BrandActionButton key={a.id} action={a} busy={busy} onRun={runAction} />
          ))}
        </div>
        {data.pending_action_id && (
          <div className="dossier-pending-btns">
            <button
              type="button"
              className="pcard__btn pcard__btn--ignore"
              disabled={busy}
              onClick={() =>
                ignore.mutate(data.pending_action_id!, {
                  onSuccess: () => {
                    toast.push("Ignored", "info");
                    refetch();
                  },
                  onError: (e) => toast.push((e as Error).message, "err"),
                })
              }
            >
              Ignore pending
            </button>
          </div>
        )}
      </section>

      <div className="dossier__grid">
        <section className="dossier__section">
          <h3>Celebrate copy (day 0)</h3>
          <blockquote className="dossier-copy">{data.celebrate_copy}</blockquote>
          <h3>Soft invite (after celebrate)</h3>
          <blockquote className="dossier-copy">{data.soft_invite_copy}</blockquote>
          {data.handoff_url && (
            <p className="dossier-handoff">
              Handoff:{" "}
              <a href={data.handoff_url} target="_blank" rel="noreferrer">
                {data.handoff_url}
              </a>
            </p>
          )}
        </section>

        <section className="dossier__section">
          <h3>Audit ribbon</h3>
          <p className="dossier__hint">Every stage decision — including "do nothing".</p>
          <div className="dossier-audit">
            {(data.audit_trail || []).slice(0, 25).map((a) => (
              <div key={a.id} className="dossier-audit__row">
                <time>{a.created_at?.slice(0, 19)?.replace("T", " ")}</time>
                <code>{a.event}</code>
                <span>
                  {a.entity_type}:{a.entity_id.slice(0, 12)}
                </span>
              </div>
            ))}
            {(data.audit_trail || []).length === 0 && <p className="work-drawer__muted">No audit rows yet.</p>}
          </div>
        </section>
      </div>

      <NotesSection coupleId={coupleId} />
    </div>
  );
}
