import { useEffect, useState } from "react";
import { CoupleGraphSvg } from "../components/CoupleGraphSvg";
import { StageBadge } from "../components/StageBadge";
import { ConfidenceBar } from "../components/ConfidenceBar";
import { VisibilityBadge } from "../components/VisibilityBadge";
import { useCoupleGraph, useCouples, usePauseCouple, useRelationship, useResumeCouple } from "../api/hooks";

export function CoupleGraphView() {
  const { data: couples, isLoading, error } = useCouples();
  const [selectedId, setSelectedId] = useState<string | null>(null);

  // Default to the first couple once the list loads; a selector change wins.
  useEffect(() => {
    if (!selectedId && couples && couples.length > 0) {
      setSelectedId(couples[0].id);
    }
  }, [couples, selectedId]);

  const couple = couples?.find((c) => c.id === selectedId) ?? couples?.[0];
  const { data: graph } = useCoupleGraph(couple?.id);
  const { data: relationship } = useRelationship(couple?.id);
  const pause = usePauseCouple();
  const resume = useResumeCouple();

  if (isLoading) {
    return (
      <div className="view">
        <h2 className="view__title">Couple graph</h2>
        <div className="empty-state">Loading couples…</div>
      </div>
    );
  }

  if (error || !couples || couples.length === 0) {
    return (
      <div className="view">
        <h2 className="view__title">Couple graph</h2>
        <div className="empty-state">
          {error
            ? `Couples unavailable: ${(error as Error).message}`
            : "No couple resolved yet — couples appear here as the watch loop resolves them."}
        </div>
      </div>
    );
  }

  if (!couple) return null;

  const isPaused = relationship?.current.automation_paused ?? false;

  return (
    <div className="view">
      <div className="couple-picker-row">
        <h2 className="view__title">
          Couple graph — {couple.person_a_label} &amp; {couple.person_b_label}
        </h2>
        {couples.length > 1 && (
          <select
            className="couple-picker"
            value={couple.id}
            onChange={(e) => setSelectedId(e.target.value)}
            aria-label="Select couple"
          >
            {couples.map((c) => (
              <option key={c.id} value={c.id}>
                {c.person_a_label} & {c.person_b_label}
              </option>
            ))}
          </select>
        )}
      </div>
      {relationship && (
        <div className="couple-header">
          <StageBadge stage={relationship.current.stage} />
          <VisibilityBadge scope={relationship.current.visibility_scope} />
          {isPaused && <span className="paused-badge">⏸ automation paused</span>}
          <ConfidenceBar value={relationship.current.confidence} label="Relationship confidence" />
          <button
            className={`btn ${isPaused ? "btn--primary" : "btn--ghost"}`}
            onClick={() => (isPaused ? resume.mutate(couple.id) : pause.mutate(couple.id))}
            disabled={pause.isPending || resume.isPending}
            title={isPaused ? "Resume automation — Neptune starts acting on signals again" : "Pause automation — Neptune stops acting on new signals for this couple"}
          >
            {isPaused ? "Resume" : "Pause"}
          </button>
        </div>
      )}
      {graph && <CoupleGraphSvg graph={graph} />}

      {relationship && relationship.history.length > 1 && (
        <>
          <h3 className="view__subtitle-heading">Stage history</h3>
          <ul className="stage-history">
            {relationship.history.map((h) => (
              <li key={h.id}>
                <StageBadge stage={h.stage} /> since {new Date(h.effective_from).toLocaleString()}
                {h.effective_to && <span className="stage-history__ended"> → superseded</span>}
              </li>
            ))}
          </ul>
        </>
      )}
    </div>
  );
}
