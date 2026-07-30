import { useState } from "react";
import { AuditTable } from "../components/AuditTable";
import { useAudit } from "../api/hooks";

export function AuditTrailView() {
  const [monitorFilter, setMonitorFilter] = useState("");
  const { data: events, error, isLoading } = useAudit(monitorFilter || undefined);

  return (
    <div className="view">
      <h2 className="view__title">Audit trail</h2>
      <p className="view__subtitle">
        Every stage of the pipeline logs its decision — not just terminal outcomes. This is the full
        observe → normalize → resolve → hypothesize → score → policy → recommend loop, in order.
      </p>
      <div className="feed-toolbar">
        <input
          className="feed-filter"
          placeholder="Filter by monitor (e.g. vendor:weddingsbynoor)…"
          value={monitorFilter}
          onChange={(e) => setMonitorFilter(e.target.value)}
        />
      </div>
      {error ? (
        <div className="empty-state">Audit trail unavailable: {(error as Error).message}</div>
      ) : isLoading ? (
        <div className="empty-state">Loading audit trail…</div>
      ) : (
        <AuditTable events={events ?? []} />
      )}
    </div>
  );
}
