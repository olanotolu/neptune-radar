import { useEffect, useMemo, useState } from "react";
import {
  useApproveAction,
  useBuildCongratulateKit,
  useIgnoreAction,
  useProspectBoard,
  useSuppressCouple,
} from "../api/hooks";
import { mediaURL } from "../api/media";
import type { ActionPayload, ProspectCard, ProspectColumnId } from "../api/types";
import { useToast } from "../components/Toast";

const COLUMN_ORDER: ProspectColumnId[] = [
  "tagged_pair",
  "investigating",
  "engaged_signal",
  "ready_outreach",
  "approved_paused",
];

const COLUMN_META: Record<ProspectColumnId, { label: string; hint: string; accent: string }> = {
  tagged_pair: { label: "Tagged pair", hint: "Found via tags", accent: "pair" },
  investigating: { label: "Investigating", hint: "Needs review", accent: "investigate" },
  engaged_signal: { label: "Engaged signal", hint: "Strong signal", accent: "engaged" },
  ready_outreach: { label: "Ready for outreach", hint: "High confidence", accent: "ready" },
  approved_paused: { label: "Approved / Paused", hint: "Case or held", accent: "paused" },
};

type Filter = "all" | "action" | "pics" | "loc";

function parsePayload(raw?: string): ActionPayload | null {
  if (!raw) return null;
  try {
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

function displayName(label: string, handle?: string): string {
  const t = (label || "").trim();
  if (!t || (handle && t.toLowerCase() === handle.toLowerCase())) {
    return handle ? `@${handle}` : "Unknown";
  }
  return t;
}

function Avatar({ url, label, handle }: { url?: string; label: string; handle?: string }) {
  const name = displayName(label, handle);
  const initial = name.replace(/^@/, "").slice(0, 1).toUpperCase() || "?";
  if (url) {
    return <img className="pcard__avatar" src={mediaURL(url)} alt="" referrerPolicy="no-referrer" title={handle ? `@${handle}` : name} />;
  }
  return (
    <span className="pcard__avatar pcard__avatar--fallback" title={handle ? `@${handle}` : name}>
      {initial}
    </span>
  );
}

function ScoreRing({ score }: { score: number }) {
  const pct = Math.round(Math.min(1, Math.max(0, score)) * 100);
  if (pct <= 0) return null;
  const tone = pct >= 90 ? "hot" : pct >= 70 ? "warm" : "cool";
  return (
    <div className={`pcard__score pcard__score--${tone}`}>
      <span className="pcard__score-val">{pct}</span>
      <span className="pcard__score-unit">%</span>
    </div>
  );
}

function Card({
  card,
  selected,
  onSelect,
}: {
  card: ProspectCard;
  selected: boolean;
  onSelect: () => void;
}) {
  const score = card.hypothesis_score || card.confidence || 0;
  const canAct = !!card.pending_action_id;
  const nameA = displayName(card.person_a_label, card.handle_a);
  const nameB = displayName(card.person_b_label, card.handle_b);
  const loc = card.city && card.region ? `${card.city}, ${card.region}` : card.city || card.region || null;

  return (
    <article
      className={`pcard pcard--col-${card.column}${canAct ? " pcard--actionable" : ""}${selected ? " pcard--selected" : ""}`}
      onClick={onSelect}
      role="button"
      tabIndex={0}
      onKeyDown={(e) => {
        if (e.key === "Enter" || e.key === " ") {
          e.preventDefault();
          onSelect();
        }
      }}
    >
      <header className="pcard__top">
        <div className="pcard__faces" aria-hidden>
          <Avatar url={card.profile_pic_a} label={card.person_a_label} handle={card.handle_a} />
          <Avatar url={card.profile_pic_b} label={card.person_b_label} handle={card.handle_b} />
          <span className="pcard__heart">♥</span>
        </div>
        <ScoreRing score={score} />
      </header>
      <div className="pcard__identity">
        <div className="pcard__name-row">
          <span className="pcard__name">{nameA}</span>
          <span className="pcard__amp">&</span>
          <span className="pcard__name">{nameB}</span>
        </div>
        {(card.handle_a || card.handle_b) && (
          <div className="pcard__handles">
            {card.handle_a && <span>@{card.handle_a}</span>}
            {card.handle_a && card.handle_b && <span className="pcard__dot">·</span>}
            {card.handle_b && <span>@{card.handle_b}</span>}
          </div>
        )}
      </div>
      <div className="pcard__chips">
        {canAct && <span className="pcard__chip pcard__chip--action">needs action</span>}
        {loc && (
          <span className="pcard__chip pcard__chip--loc">
            <span className="pcard__chip-icon">⌖</span>
            {loc}
          </span>
        )}
        {card.needs_pics && <span className="pcard__chip">no pics</span>}
        {card.stage && card.stage !== "unknown" && (
          <span className="pcard__chip pcard__chip--stage">{card.stage.replace(/_/g, " ")}</span>
        )}
      </div>
    </article>
  );
}

function Drawer({
  card,
  onClose,
  onCongratulate,
}: {
  card: ProspectCard;
  onClose: () => void;
  onCongratulate?: (coupleId: string) => void;
}) {
  const approve = useApproveAction();
  const ignore = useIgnoreAction();
  const suppress = useSuppressCouple();
  const buildKit = useBuildCongratulateKit();
  const toast = useToast();
  const payload = parsePayload(card.proposed_payload);
  const score = card.hypothesis_score || card.confidence || 0;
  const busy = approve.isPending || ignore.isPending || suppress.isPending || buildKit.isPending;

  return (
    <aside className="work-drawer">
      <header className="work-drawer__header">
        <div>
          <h3 className="work-drawer__title">
            {displayName(card.person_a_label, card.handle_a)} & {displayName(card.person_b_label, card.handle_b)}
          </h3>
          <p className="work-drawer__sub">
            {card.handle_a && `@${card.handle_a}`}
            {card.handle_a && card.handle_b ? " · " : ""}
            {card.handle_b && `@${card.handle_b}`}
          </p>
        </div>
        <button type="button" className="btn btn--ghost btn--sm" onClick={onClose}>
          Close
        </button>
      </header>

      <div className="work-drawer__faces">
        <Avatar url={card.profile_pic_a} label={card.person_a_label} handle={card.handle_a} />
        <Avatar url={card.profile_pic_b} label={card.person_b_label} handle={card.handle_b} />
        <ScoreRing score={score} />
      </div>

      <div className="pcard__chips" style={{ marginBottom: 12 }}>
        {card.city && (
          <span className="pcard__chip pcard__chip--loc">
            {card.city}
            {card.region ? `, ${card.region}` : ""}
          </span>
        )}
        {card.stage && <span className="pcard__chip pcard__chip--stage">{card.stage.replace(/_/g, " ")}</span>}
        {card.pending_action_type && (
          <span className="pcard__chip pcard__chip--action">{card.pending_action_type.replace(/_/g, " ")}</span>
        )}
        {card.has_case && <span className="pcard__chip pcard__chip--case">case open</span>}
      </div>

      {(card.bio_a || card.bio_b) && (
        <section className="work-drawer__section">
          <h4>Bios</h4>
          {card.bio_a && (
            <p className="work-drawer__bio">
              <strong>@{card.handle_a || "a"}:</strong> {card.bio_a}
            </p>
          )}
          {card.bio_b && (
            <p className="work-drawer__bio">
              <strong>@{card.handle_b || "b"}:</strong> {card.bio_b}
            </p>
          )}
        </section>
      )}

      {payload && (
        <section className="work-drawer__section">
          <h4>Recommended copy</h4>
          <div className="work-drawer__copy">
            <div className="work-drawer__copy-label">Internal</div>
            <pre>{payload.internal_note}</pre>
          </div>
          <div className="work-drawer__copy">
            <div className="work-drawer__copy-label">Customer-facing (if approved)</div>
            <p>{payload.customer_facing}</p>
          </div>
          {payload.reasons && payload.reasons.length > 0 && (
            <ul className="work-drawer__reasons">
              {payload.reasons.map((r) => (
                <li key={r}>{r}</li>
              ))}
            </ul>
          )}
        </section>
      )}

      {!payload && !card.pending_action_id && (
        <p className="work-drawer__muted">No pending approval copy for this couple.</p>
      )}

      <footer className="work-drawer__footer">
        <button
          type="button"
          className="pcard__btn pcard__btn--approve"
          disabled={busy}
          title="Build postcard + address research kit"
          onClick={() =>
            buildKit.mutate(card.couple_id, {
              onSuccess: (kit) => {
                toast.push("Congratulate kit ready", "ok");
                onCongratulate?.(kit.couple_id);
              },
              onError: (e) => toast.push((e as Error).message, "err"),
            })
          }
        >
          {buildKit.isPending ? "Building kit…" : "Congratulate → postcard"}
        </button>
        {card.pending_action_id && (
          <>
            <button
              type="button"
              className="pcard__btn pcard__btn--approve"
              disabled={busy}
              onClick={() =>
                approve.mutate(card.pending_action_id!, {
                  onSuccess: () => {
                    toast.push("Approved", "ok");
                    onClose();
                  },
                  onError: (e) => toast.push((e as Error).message, "err"),
                })
              }
            >
              Approve
            </button>
            <button
              type="button"
              className="pcard__btn pcard__btn--ignore"
              disabled={busy}
              onClick={() =>
                ignore.mutate(card.pending_action_id!, {
                  onSuccess: () => {
                    toast.push("Ignored", "info");
                    onClose();
                  },
                  onError: (e) => toast.push((e as Error).message, "err"),
                })
              }
            >
              Ignore
            </button>
          </>
        )}
        <button
          type="button"
          className="pcard__btn pcard__btn--suppress"
          disabled={busy}
          title="Permanently hide — not a real couple"
          onClick={() => {
            if (!confirm("Mark as not a couple and remove from the board?")) return;
            suppress.mutate(
              { id: card.couple_id, reason: "not_a_couple" },
              {
                onSuccess: () => {
                  toast.push("Suppressed — not a couple", "ok");
                  onClose();
                },
                onError: (e) => toast.push((e as Error).message, "err"),
              },
            );
          }}
        >
          Not a couple
        </button>
      </footer>
    </aside>
  );
}

export function WorkView({
  initialFilter = "all",
  focusCoupleId,
  onCongratulate,
}: {
  initialFilter?: Filter;
  focusCoupleId?: string;
  onCongratulate?: (coupleId: string) => void;
}) {
  const { data, error, isLoading } = useProspectBoard();
  const [filter, setFilter] = useState<Filter>(initialFilter);
  const [selectedId, setSelectedId] = useState<string | null>(focusCoupleId ?? null);

  // Props drive state: the header's queue button and ?couple= deep links
  // re-render this mounted component with new values — without these syncs
  // the URL changed but the board silently ignored it.
  useEffect(() => setFilter(initialFilter), [initialFilter]);
  useEffect(() => {
    if (focusCoupleId) setSelectedId(focusCoupleId);
  }, [focusCoupleId]);

  const allCards = useMemo(() => {
    if (!data) return [] as ProspectCard[];
    const list: ProspectCard[] = [];
    for (const col of COLUMN_ORDER) {
      for (const c of data.cards[col] ?? []) list.push(c);
    }
    return list;
  }, [data]);

  const filtered = useMemo(() => {
    return allCards.filter((c) => {
      if (filter === "action") return c.needs_action || !!c.pending_action_id;
      if (filter === "pics") return c.needs_pics;
      if (filter === "loc") return c.needs_location;
      return true;
    });
  }, [allCards, filter]);

  const byColumn = useMemo(() => {
    const m = new Map<ProspectColumnId, ProspectCard[]>();
    for (const col of COLUMN_ORDER) m.set(col, []);
    for (const c of filtered) {
      (m.get(c.column as ProspectColumnId) ?? m.get("tagged_pair")!).push(c);
    }
    return m;
  }, [filtered]);

  const selected = allCards.find((c) => c.couple_id === selectedId) ?? null;
  const actionCount = allCards.filter((c) => c.pending_action_id).length;

  return (
    <div className={`view view--work ${selected ? "view--work-split" : ""}`}>
      <div className="work-main">
        <header className="prospects-hero">
          <div>
            <h2 className="view__title">Work</h2>
            <p className="view__subtitle">
              Stage board + approval in one place. Open a card for copy, bios, and decisions. Nothing customer-facing without Approve.
            </p>
          </div>
          {data && (
            <div className="prospects-hero__stat">
              <span className="prospects-hero__n">{actionCount}</span>
              <span className="prospects-hero__l">need action</span>
            </div>
          )}
        </header>

        <div className="work-filters">
          {(
            [
              ["all", "All"],
              ["action", "Needs action"],
              ["pics", "Missing pics"],
              ["loc", "Missing location"],
            ] as const
          ).map(([id, label]) => (
            <button
              key={id}
              type="button"
              className={`work-filter ${filter === id ? "work-filter--active" : ""}`}
              onClick={() => setFilter(id)}
            >
              {label}
            </button>
          ))}
        </div>

        {error && <div className="empty-state">{(error as Error).message}</div>}
        {isLoading && <div className="empty-state">Loading work queue…</div>}
        {data && filtered.length === 0 && (
          <div className="empty-state empty-state--board">
            <div className="empty-state__glyph">◇</div>
            <strong>Nothing in this filter</strong>
            <p>Try All, or run the agent on a source to discover tagged couples.</p>
          </div>
        )}

        {data && filtered.length > 0 && (
          <div className="prospect-board">
            {COLUMN_ORDER.map((colId) => {
              const cards = byColumn.get(colId) ?? [];
              const meta = COLUMN_META[colId];
              return (
                <section className={`prospect-column prospect-column--${meta.accent}`} key={colId}>
                  <header className="prospect-column__header">
                    <div className="prospect-column__titles">
                      <span className="prospect-column__title">{meta.label}</span>
                      <span className="prospect-column__hint">{meta.hint}</span>
                    </div>
                    <span className="prospect-column__count">{cards.length}</span>
                  </header>
                  <div className="prospect-column__cards">
                    {cards.length === 0 ? (
                      <div className="prospect-column__empty">No cards</div>
                    ) : (
                      cards.map((c) => (
                        <Card
                          key={c.couple_id}
                          card={c}
                          selected={selectedId === c.couple_id}
                          onSelect={() => setSelectedId(c.couple_id)}
                        />
                      ))
                    )}
                  </div>
                </section>
              );
            })}
          </div>
        )}
      </div>
      {selected && (
        <Drawer
          card={selected}
          onClose={() => setSelectedId(null)}
          onCongratulate={(id) => onCongratulate?.(id)}
        />
      )}
    </div>
  );
}
