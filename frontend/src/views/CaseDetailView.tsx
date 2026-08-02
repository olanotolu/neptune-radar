import { EvidenceTimelineTable } from "../components/EvidenceTimelineTable";
import { ConfidenceBar } from "../components/ConfidenceBar";
import { EmptyState } from "../components/EmptyState";
import { LoadingState } from "../components/LoadingState";
import { useCases, useConfidence, useEvidence, useLeads } from "../api/hooks";
import type { NeptuneCase, CRMLead } from "../api/types";

function CaseCard({ neptuneCase, lead }: { neptuneCase: NeptuneCase; lead?: CRMLead }) {
  const { data: evidence } = useEvidence(lead?.hypothesis_id);
  const { data: confidence } = useConfidence(lead?.hypothesis_id);

  return (
    <div className="case-card">
      <div className="case-card__header">
        <h3>{neptuneCase.case_type} case</h3>
        <span className={`case-status case-status--${neptuneCase.status}`}>{neptuneCase.status}</span>
      </div>
      <div className="case-card__meta">
        Created {new Date(neptuneCase.created_at).toLocaleString()} · updated{" "}
        {new Date(neptuneCase.updated_at).toLocaleString()}
      </div>
      {lead && <div className="case-card__meta">Lead type: {lead.lead_type} · lead status: {lead.status}</div>}

      {confidence && <ConfidenceBar value={confidence.final} label="Hypothesis confidence" />}

      <h4 className="case-card__section-title">Evidence timeline</h4>
      {evidence && <EvidenceTimelineTable evidence={evidence} />}
    </div>
  );
}

export function CaseDetailView() {
  const { data: cases, isLoading, error } = useCases();
  const { data: leads } = useLeads();

  if (isLoading || error || !cases || cases.length === 0) {
    return (
      <div className="view">
        <h2 className="view__title">Active Neptune case</h2>
        {isLoading ? (
          <LoadingState variant="spinner" message="Loading cases…" />
        ) : error ? (
          <EmptyState variant="warning" icon="⚠" title="Cases unavailable" message={(error as Error).message} />
        ) : (
          <EmptyState
            variant="empty"
            icon="📂"
            title="No case yet"
            message="Approve the engagement review action in the Approval Queue to open one."
          />
        )}
      </div>
    );
  }

  return (
    <div className="view">
      <h2 className="view__title">Active Neptune case</h2>
      {cases.map((c) => (
        <CaseCard key={c.id} neptuneCase={c} lead={leads?.find((l) => l.id === c.lead_id)} />
      ))}
    </div>
  );
}
