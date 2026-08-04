import { useEffect, useState } from "react";
import { CoupleGraphSvg } from "../components/CoupleGraphSvg";
import { StageBadge } from "../components/StageBadge";
import { VisibilityBadge } from "../components/VisibilityBadge";
import { useCoupleGraph, useCouples, usePauseCouple, useRelationship, useResumeCouple } from "../api/hooks";
import { EmptyState } from "../components/EmptyState";
import { LoadingState } from "../components/LoadingState";

export function CoupleGraphView() {
  const { data: couples, isLoading, error } = useCouples();
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [pickerOpen, setPickerOpen] = useState(false);

  // Default to the first couple once the list loads; a selector change wins.
  useEffect(() => {
    if (!selectedId && couples && couples.length > 0) {
      setSelectedId(couples[0].id);
    }
  }, [couples, selectedId]);

  // close picker on outside click
  useEffect(() => {
    if (!pickerOpen) return;
    const close = () => setPickerOpen(false);
    window.addEventListener("click", close);
    return () => window.removeEventListener("click", close);
  }, [pickerOpen]);

  const couple = couples?.find((c) => c.id === selectedId) ?? couples?.[0];
  const { data: graph } = useCoupleGraph(couple?.id);
  const { data: relationship } = useRelationship(couple?.id);
  const pause = usePauseCouple();
  const resume = useResumeCouple();

  if (isLoading) {
    return (
      <div className="view view--couple-graph">
        <h2 className="view__title">Relationship Graph</h2>
        <LoadingState variant="spinner" message="Loading couples…" />
      </div>
    );
  }

  if (error || !couples || couples.length === 0) {
    return (
      <div className="view view--couple-graph">
        <h2 className="view__title">Relationship Graph</h2>
        {error ? (
          <EmptyState variant="warning" title="Couples unavailable" message={(error as Error).message} />
        ) : (
          <EmptyState
            variant="empty"
            title="No couples detected yet"
            message="The radar is watching. Couples will appear here as they're discovered."
          />
        )}
      </div>
    );
  }

  if (!couple) return null;

  const isPaused = relationship?.current.automation_paused ?? false;
  const confidence = relationship?.current.confidence ?? 0;
  const stage = relationship?.current.stage;
  const interactions =
    graph?.edges.filter((e) => e.kind !== "owns_account").length ?? 0;
  const accountCount = graph?.nodes.filter((n) => n.type === "account").length ?? 0;
  const history = relationship?.history ?? [];

  return (
    <div className="view view--couple-graph">
      <header className="graph-header">
        <div className="graph-header__titles">
          <h1 className="graph-header__title">Relationship Graph</h1>
          <p className="graph-header__subtitle">
            How they're connected. How we found them.
          </p>
        </div>

        {couples.length > 1 && (
          <div
            className={`couple-picker-premium ${pickerOpen ? "is-open" : ""}`}
            onClick={(e) => e.stopPropagation()}
          >
            <button
              type="button"
              className="couple-picker-premium__trigger"
              onClick={() => setPickerOpen((o) => !o)}
              aria-haspopup="listbox"
              aria-expanded={pickerOpen}
            >
              <span className="couple-picker-premium__avatars">
                <span className="couple-picker-premium__avatar couple-picker-premium__avatar--a">
                  {initials(couple.person_a_label)}
                </span>
                <span className="couple-picker-premium__avatar couple-picker-premium__avatar--b">
                  {initials(couple.person_b_label)}
                </span>
              </span>
              <span className="couple-picker-premium__label">
                {couple.person_a_label} &amp; {couple.person_b_label}
              </span>
              <span className="couple-picker-premium__chevron" aria-hidden="true">
                ▾
              </span>
            </button>
            {pickerOpen && (
              <ul className="couple-picker-premium__menu" role="listbox">
                {couples.map((c) => (
                  <li key={c.id}>
                    <button
                      type="button"
                      className={`couple-picker-premium__option ${c.id === couple.id ? "is-selected" : ""}`}
                      onClick={() => {
                        setSelectedId(c.id);
                        setPickerOpen(false);
                      }}
                      role="option"
                      aria-selected={c.id === couple.id}
                    >
                      <span className="couple-picker-premium__avatars">
                        <span className="couple-picker-premium__avatar couple-picker-premium__avatar--a">
                          {initials(c.person_a_label)}
                        </span>
                        <span className="couple-picker-premium__avatar couple-picker-premium__avatar--b">
                          {initials(c.person_b_label)}
                        </span>
                      </span>
                      <span className="couple-picker-premium__label">
                        {c.person_a_label} &amp; {c.person_b_label}
                      </span>
                    </button>
                  </li>
                ))}
              </ul>
            )}
          </div>
        )}
      </header>

      <div className="graph-stats">
        <div className="graph-stat">
          <span className="graph-stat__value">{Math.round(confidence * 100)}%</span>
          <span className="graph-stat__label">Confidence</span>
        </div>
        <div className="graph-stat">
          <span className="graph-stat__value">{stage ? <StageBadge stage={stage} /> : "—"}</span>
          <span className="graph-stat__label">Stage</span>
        </div>
        <div className="graph-stat">
          <span className="graph-stat__value">{interactions}</span>
          <span className="graph-stat__label">Interactions</span>
        </div>
        <div className="graph-stat">
          <span className="graph-stat__value">{accountCount}</span>
          <span className="graph-stat__label">Accounts</span>
        </div>
        {relationship && (
          <div className="graph-stat graph-stat--badges">
            <VisibilityBadge scope={relationship.current.visibility_scope} />
            {isPaused && <span className="paused-badge">paused</span>}
          </div>
        )}
        <button
          className={`btn ${isPaused ? "btn--primary" : "btn--ghost"} graph-stats__action`}
          onClick={() => (isPaused ? resume.mutate(couple.id) : pause.mutate(couple.id))}
          disabled={pause.isPending || resume.isPending}
          title={
            isPaused
              ? "Resume automation — Neptune starts acting on signals again"
              : "Pause automation — Neptune stops acting on new signals for this couple"
          }
        >
          {isPaused ? "Resume" : "Pause"}
        </button>
      </div>

      <div className="graph-layout">
        <div className="graph-card">
          {graph ? (
            <CoupleGraphSvg graph={graph} confidence={Math.round(confidence * 100)} />
          ) : (
            <div className="graph-card__loading">Resolving graph…</div>
          )}
        </div>

        <aside className="graph-side">
          <h3 className="graph-side__title">Stage history</h3>
          {history.length > 0 ? (
            <ol className="stage-timeline">
              {[...history].reverse().map((h, i) => (
                <li key={h.id} className={`stage-timeline__item ${i === 0 ? "is-current" : ""}`}>
                  <span className="stage-timeline__dot" />
                  <div className="stage-timeline__body">
                    <StageBadge stage={h.stage} />
                    <span className="stage-timeline__date">
                      {new Date(h.effective_from).toLocaleDateString(undefined, {
                        year: "numeric",
                        month: "short",
                        day: "numeric",
                      })}
                    </span>
                    {h.effective_to && (
                      <span className="stage-timeline__ended">superseded</span>
                    )}
                  </div>
                </li>
              ))}
            </ol>
          ) : (
            <p className="graph-side__empty">No stage transitions recorded yet.</p>
          )}
        </aside>
      </div>
    </div>
  );
}

function initials(name: string): string {
  return name
    .split(" ")
    .map((p) => p[0])
    .join("")
    .slice(0, 2)
    .toUpperCase();
}
