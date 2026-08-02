import { useEffect, useMemo, useState } from "react";
import {
  useApproveAction,
  useBuildCongratulateKit,
  useIgnoreAction,
  useProspectBoard,
  useSetJourneyStage,
  useSuppressCouple,
} from "../api/hooks";
import { mediaURL } from "../api/media";
import type { ActionPayload, ProspectCard, ProspectColumnId } from "../api/types";
import { useToast } from "../components/Toast";
import { useKeyboard } from "../hooks/useKeyboard";
import { DossierView } from "./DossierView";

const COLUMN_ORDER: ProspectColumnId[] = [
  "tagged_pair",
  "investigating",
  "engaged_signal",
  "ready_outreach",
  "approved_paused",
];

const COLUMN_META: Record<
  ProspectColumnId,
  { label: string; hint: string; accent: string; stage: string }
> = {
  tagged_pair: { label: "Tagged pair", hint: "Found via tags", accent: "pair", stage: "tagged_pair" },
  investigating: { label: "Investigating", hint: "Needs review", accent: "investigate", stage: "investigating" },
  engaged_signal: { label: "Engaged signal", hint: "Strong signal", accent: "engaged", stage: "engaged_signal" },
  ready_outreach: { label: "Ready for outreach", hint: "High confidence", accent: "ready", stage: "ready_outreach" },
  approved_paused: { label: "Approved / Paused", hint: "Case or held", accent: "paused", stage: "approved_paused" },
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

function RankBadge({ score }: { score: number }) {
  const pct = Math.round(Math.min(1, Math.max(0, score)) * 100);
  if (pct <= 0) return null;
  const tone = pct >= 90 ? "hot" : pct >= 70 ? "warm" : "cool";
  return (
    <div className={`pcard__rank pcard__rank--${tone}`} title="Neptune rank">
      {pct}
    </div>
  );
}

function RunwayChip({ card }: { card: ProspectCard }) {
  const band = card.runway?.band || "unknown";
  if (band === "unknown" && !card.runway_label) return null;
  return (
    <span className={`pcard__chip pcard__chip--runway-${band}`} title="Wedding runway for prenup process">
      {card.runway_label || band}
    </span>
  );
}

function Card({
  card,
  selected,
  onSelect,
  dragging,
  bulkSelected,
  onBulkToggle,
  onDragStart,
  onDragEnd,
}: {
  card: ProspectCard;
  selected: boolean;
  onSelect: () => void;
  dragging: boolean;
  bulkSelected: boolean;
  onBulkToggle: () => void;
  onDragStart: () => void;
  onDragEnd: () => void;
}) {
  const score = card.neptune_rank || card.hypothesis_score || card.confidence || 0;
  const canAct = !!card.pending_action_id;
  const nameA = displayName(card.person_a_label, card.handle_a);
  const nameB = displayName(card.person_b_label, card.handle_b);
  const loc = card.city && card.region ? `${card.city}, ${card.region}` : card.city || card.region || null;
  const icpLabel = card.icp?.labels?.[0];
  const confidence = card.confidence ?? card.hypothesis_score ?? score;
  const confPct = Math.round(Math.min(1, Math.max(0, confidence)) * 100);

  return (
    <article
      className={`pcard pcard--col-${card.column}${canAct ? " pcard--actionable" : ""}${selected ? " pcard--selected" : ""}${card.queue === "risk" ? " pcard--risk" : ""}${dragging ? " pcard--dragging" : ""}${bulkSelected ? " pcard--bulk" : ""}`}
      onClick={onSelect}
      role="button"
      tabIndex={0}
      draggable
      onDragStart={(e) => {
        e.dataTransfer.effectAllowed = "move";
        e.dataTransfer.setData("text/plain", card.couple_id);
        onDragStart();
      }}
      onDragEnd={onDragEnd}
      onKeyDown={(e) => {
        if (e.key === "Enter" || e.key === " ") {
          e.preventDefault();
          onSelect();
        }
      }}
    >
      <label
        className="pcard__bulk-check"
        title="Select for bulk action"
        onClick={(e) => e.stopPropagation()}
      >
        <input type="checkbox" checked={bulkSelected} onChange={onBulkToggle} />
      </label>

      <header className="pcard__top">
        <div className="pcard__faces" aria-hidden>
          <Avatar url={card.profile_pic_a} label={card.person_a_label} handle={card.handle_a} />
          <Avatar url={card.profile_pic_b} label={card.person_b_label} handle={card.handle_b} />
          <span className="pcard__heart">♥</span>
        </div>
        <RankBadge score={score} />
      </header>

      <div className="pcard__identity">
        <div className="pcard__name-row">
          <span className="pcard__name">{nameA}</span>
          <span className="pcard__amp">&</span>
          <span className="pcard__name">{nameB}</span>
        </div>
        {loc && (
          <div className="pcard__loc">
            <span className="pcard__loc-pin">⌖</span>
            {loc}
          </div>
        )}
      </div>

      <div className="pcard__chips">
        {canAct && <span className="pcard__chip pcard__chip--action">needs action</span>}
        {card.queue === "risk" && <span className="pcard__chip pcard__chip--risk">risk · no pitch</span>}
        <RunwayChip card={card} />
        {icpLabel && <span className="pcard__chip pcard__chip--icp">{icpLabel}</span>}
        {card.needs_pics && <span className="pcard__chip">no pics</span>}
        {card.stage && card.stage !== "unknown" && (
          <span className="pcard__chip pcard__chip--stage">{card.stage.replace(/_/g, " ")}</span>
        )}
      </div>

      <div className="pcard__confidence" title={`Confidence ${confPct}%`}>
        <div className="pcard__confidence-fill" style={{ width: `${confPct}%` }} />
      </div>
    </article>
  );
}

function Drawer({
  card,
  onClose,
  onCongratulate,
  fullDossier,
  onToggleDossier,
}: {
  card: ProspectCard;
  onClose: () => void;
  onCongratulate?: (coupleId: string) => void;
  fullDossier: boolean;
  onToggleDossier: () => void;
}) {
  const approve = useApproveAction();
  const ignore = useIgnoreAction();
  const suppress = useSuppressCouple();
  const buildKit = useBuildCongratulateKit();
  const toast = useToast();
  const payload = parsePayload(card.proposed_payload);
  const busy = approve.isPending || ignore.isPending || suppress.isPending || buildKit.isPending;
  const nameA = displayName(card.person_a_label, card.handle_a);
  const nameB = displayName(card.person_b_label, card.handle_b);

  return (
    <div className="work-overlay" onClick={onClose}>
      <aside className={`work-drawer${fullDossier ? " work-drawer--dossier" : ""}`} onClick={(e) => e.stopPropagation()}>
        <button type="button" className="work-drawer__close" onClick={onClose} aria-label="Close">
          ✕
        </button>

        {fullDossier ? (
          <>
            <DossierView
              coupleId={card.couple_id}
              onClose={onClose}
              onCongratulate={onCongratulate}
            />
            <div className="work-drawer__dossier-toggle">
              <button type="button" className="btn btn--ghost btn--sm" onClick={onToggleDossier}>
                Compact drawer
              </button>
            </div>
          </>
        ) : (
          <div className="work-drawer__body">
            <header className="work-drawer__header">
              <div className="work-drawer__faces">
                <Avatar url={card.profile_pic_a} label={card.person_a_label} handle={card.handle_a} />
                <Avatar url={card.profile_pic_b} label={card.person_b_label} handle={card.handle_b} />
              </div>
              <div>
                <h3 className="work-drawer__title">
                  {nameA} <span className="pcard__amp">&</span> {nameB}
                </h3>
                <p className="work-drawer__sub">
                  {card.handle_a && `@${card.handle_a}`}
                  {card.handle_a && card.handle_b ? " · " : ""}
                  {card.handle_b && `@${card.handle_b}`}
                  {card.neptune_rank != null && ` · rank ${Math.round(card.neptune_rank * 100)}`}
                </p>
              </div>
            </header>

            <div className="pcard__chips" style={{ marginBottom: 12 }}>
              {card.city && (
                <span className="pcard__chip pcard__chip--loc">
                  {card.city}
                  {card.region ? `, ${card.region}` : ""}
                </span>
              )}
              <RunwayChip card={card} />
              {(card.icp?.labels || []).slice(0, 3).map((l) => (
                <span key={l} className="pcard__chip pcard__chip--icp">
                  {l}
                </span>
              ))}
              {card.stage && <span className="pcard__chip pcard__chip--stage">{card.stage.replace(/_/g, " ")}</span>}
              {card.pending_action_type && (
                <span className="pcard__chip pcard__chip--action">{card.pending_action_type.replace(/_/g, " ")}</span>
              )}
              {card.has_case && <span className="pcard__chip pcard__chip--case">case open</span>}
              {card.queue === "risk" && <span className="pcard__chip pcard__chip--risk">risk · no pitch</span>}
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
                  <div className="work-drawer__copy-label">Customer-facing (celebrate-first)</div>
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
                disabled={busy || card.queue === "risk"}
                title={card.queue === "risk" ? "Risk path — no celebrate pitch" : "Build postcard + address research kit"}
                onClick={() =>
                  buildKit.mutate(card.couple_id, {
                    onSuccess: (kit) => {
                      toast.push("Congratulate kit ready — celebration only", "ok");
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

            <button type="button" className="work-drawer__open-dossier-link" onClick={onToggleDossier}>
              View full dossier →
            </button>
          </div>
        )}
      </aside>
    </div>
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
  const [fullDossier, setFullDossier] = useState(!!focusCoupleId);
  const [dragId, setDragId] = useState<string | null>(null);
  const [dragOverCol, setDragOverCol] = useState<ProspectColumnId | null>(null);
  const [bulkSet, setBulkSet] = useState<Set<string>>(new Set());
  const approve = useApproveAction();
  const ignore = useIgnoreAction();
  const setStage = useSetJourneyStage();
  const suppress = useSuppressCouple();
  const toast = useToast();

  // Props drive state: the header's queue button and ?couple= deep links
  // re-render this mounted component with new values — without these syncs
  // the URL changed but the board silently ignored it.
  useEffect(() => setFilter(initialFilter), [initialFilter]);
  useEffect(() => {
    if (focusCoupleId) {
      setSelectedId(focusCoupleId);
      setFullDossier(true);
    }
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
  const bulkCount = bulkSet.size;

  function toggleBulk(id: string) {
    setBulkSet((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  }

  function clearBulk() {
    setBulkSet(new Set());
  }

  // Drag-and-drop: map target column → API call. Forward moves approve any
  // pending action so the couple advances cleanly; setJourneyStage handles
  // the column transition itself. ponytail: stage string mirrors column id.
  function handleDrop(targetCol: ProspectColumnId, coupleId: string) {
    setDragOverCol(null);
    setDragId(null);
    const card = allCards.find((c) => c.couple_id === coupleId);
    if (!card || card.column === targetCol) return;
    const meta = COLUMN_META[targetCol];
    const isForward = COLUMN_ORDER.indexOf(targetCol) > COLUMN_ORDER.indexOf(card.column);

    const done = () => toast.push(`Moved to ${meta.label}`, "ok");
    const fail = (e: unknown) => toast.push((e as Error).message, "err");

    // Forward into outreach/approved also clears the pending action.
    if (isForward && card.pending_action_id && (targetCol === "ready_outreach" || targetCol === "approved_paused")) {
      approve.mutate(card.pending_action_id, {
        onSuccess: () => {
          setStage.mutate({ coupleId, stage: meta.stage }, { onSuccess: done, onError: fail });
        },
        onError: fail,
      });
    } else {
      setStage.mutate({ coupleId, stage: meta.stage }, { onSuccess: done, onError: fail });
    }
  }

  function bulkApprove() {
    let n = 0;
    for (const id of bulkSet) {
      const card = allCards.find((c) => c.couple_id === id);
      if (card?.pending_action_id) {
        n++;
        approve.mutate(card.pending_action_id, {
          onError: (e) => toast.push((e as Error).message, "err"),
        });
      }
    }
    toast.push(n > 0 ? `Approved ${n} couple${n > 1 ? "s" : ""}` : "No actionable selections", "ok");
    clearBulk();
  }

  function bulkSuppress() {
    if (!confirm(`Suppress ${bulkCount} couple${bulkCount > 1 ? "s" : ""} as not-a-couple?`)) return;
    for (const id of bulkSet) {
      suppress.mutate(
        { id, reason: "not_a_couple" },
        { onError: (e) => toast.push((e as Error).message, "err") },
      );
    }
    toast.push(`Suppressed ${bulkCount} couple${bulkCount > 1 ? "s" : ""}`, "ok");
    clearBulk();
  }

  // Keyboard navigation over the filtered queue. j/k move the selection,
  // a/i act on the selected card's pending action, Enter opens the dossier.
  useKeyboard((e) => {
    if (filtered.length === 0) return;
    const idx = filtered.findIndex((c) => c.couple_id === selectedId);
    if (e.key === "j") {
      e.preventDefault();
      const next = filtered[Math.min(filtered.length - 1, idx + 1)];
      if (next) setSelectedId(next.couple_id);
    } else if (e.key === "k") {
      e.preventDefault();
      const prev = filtered[Math.max(0, idx - 1)];
      if (prev) setSelectedId(prev.couple_id);
    } else if (e.key === "a" && selected?.pending_action_id) {
      e.preventDefault();
      approve.mutate(selected.pending_action_id, {
        onSuccess: () => toast.push("Approved", "ok"),
        onError: (err) => toast.push((err as Error).message, "err"),
      });
    } else if (e.key === "i" && selected?.pending_action_id) {
      e.preventDefault();
      ignore.mutate(selected.pending_action_id, {
        onSuccess: () => toast.push("Ignored", "info"),
        onError: (err) => toast.push((err as Error).message, "err"),
      });
    } else if (e.key === "Enter" && selected) {
      e.preventDefault();
      setFullDossier(true);
    }
  });

  return (
    <div className="view view--work">
      <div className="work-main">
        <header className="prospects-hero">
          <div>
            <h2 className="view__title">Prospect board</h2>
            <p className="view__subtitle">
              Every couple here is a future partnership. Celebrate first.
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
              ["all", "All", "default"],
              ["action", "Needs action", "cove"],
              ["pics", "Missing pics", "mesa"],
              ["loc", "Missing location", "mesa"],
            ] as const
          ).map(([id, label, tone]) => (
            <button
              key={id}
              type="button"
              className={`work-filter work-filter--${tone} ${filter === id ? "work-filter--active" : ""}`}
              onClick={() => setFilter(id)}
            >
              {label}
            </button>
          ))}
        </div>

        {bulkCount > 0 && (
          <div className="work-bulk-bar">
            <span className="work-bulk-bar__count">{bulkCount} selected</span>
            <button type="button" className="work-bulk-bar__btn work-bulk-bar__btn--approve" onClick={bulkApprove}>
              Approve all
            </button>
            <button type="button" className="work-bulk-bar__btn work-bulk-bar__btn--suppress" onClick={bulkSuppress}>
              Suppress all
            </button>
            <button type="button" className="work-bulk-bar__btn work-bulk-bar__btn--clear" onClick={clearBulk}>
              Clear
            </button>
          </div>
        )}

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
              const isOver = dragOverCol === colId;
              return (
                <section
                  className={`prospect-column prospect-column--${meta.accent}${isOver ? " prospect-column--drop" : ""}`}
                  key={colId}
                  onDragOver={(e) => {
                    e.preventDefault();
                    e.dataTransfer.dropEffect = "move";
                    if (dragOverCol !== colId) setDragOverCol(colId);
                  }}
                  onDragLeave={(e) => {
                    // only clear if we actually left the column, not a child
                    if (!e.currentTarget.contains(e.relatedTarget as Node)) setDragOverCol(null);
                  }}
                  onDrop={(e) => {
                    e.preventDefault();
                    const id = e.dataTransfer.getData("text/plain");
                    if (id) handleDrop(colId, id);
                  }}
                >
                  <header className="prospect-column__header">
                    <div className="prospect-column__titles">
                      <span className="prospect-column__title">{meta.label}</span>
                      <span className="prospect-column__hint">{meta.hint}</span>
                    </div>
                    <span className="prospect-column__count">{cards.length}</span>
                  </header>
                  <div className="prospect-column__cards">
                    {cards.length === 0 ? (
                      <div className="prospect-column__empty">
                        <span className="prospect-column__empty-icon">◇</span>
                        <span>No couples here yet</span>
                      </div>
                    ) : (
                      cards.map((c) => (
                        <Card
                          key={c.couple_id}
                          card={c}
                          selected={selectedId === c.couple_id}
                          onSelect={() => setSelectedId(c.couple_id)}
                          dragging={dragId === c.couple_id}
                          bulkSelected={bulkSet.has(c.couple_id)}
                          onBulkToggle={() => toggleBulk(c.couple_id)}
                          onDragStart={() => setDragId(c.couple_id)}
                          onDragEnd={() => {
                            setDragId(null);
                            setDragOverCol(null);
                          }}
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
          onClose={() => {
            setSelectedId(null);
            setFullDossier(false);
          }}
          onCongratulate={(id) => onCongratulate?.(id)}
          fullDossier={fullDossier}
          onToggleDossier={() => setFullDossier((v) => !v)}
        />
      )}
    </div>
  );
}
