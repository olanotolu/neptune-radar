import { useVisionAnalysis } from "../api/hooks";
import type { VisionAnalysis } from "../api/types";
import { EmptyState } from "../components/EmptyState";
import { LoadingState } from "../components/LoadingState";

function pct(n: number) {
  return n > 0 ? `${Math.round(n * 100)}%` : "—";
}

function ringTone(conf: number): string {
  if (conf >= 0.7) return "status-dot--live";
  if (conf >= 0.5) return "status-dot--warn";
  return "";
}

export function VisionView() {
  const { data, isLoading, error } = useVisionAnalysis(100);
  const rows = data ?? [];

  const rings = rows.filter((r) => r.ring_confidence >= 0.5);
  const proposals = rows.filter(
    (r) => r.photo_label === "marriage proposal" || r.photo_label === "engagement photo shoot",
  );
  const errors = rows.filter((r) => r.error);

  return (
    <div className="view view--vision">
      <header className="view__head">
        <h2>Vision analysis</h2>
        <p className="trust-panel__sub">
          Ring detection (YOLOv8) + photo classification (CLIP zero-shot) per ingested post.
        </p>
      </header>

      <section className="trust-panel">
        <div className="funnel-kpis">
          <div className="funnel-kpi">
            <strong>{rows.length}</strong>
            <span>Total analyses</span>
          </div>
          <div className="funnel-kpi">
            <strong className={rings.length > 0 ? "status-dot status-dot--live" : ""}>{rings.length}</strong>
            <span>Rings detected</span>
          </div>
          <div className="funnel-kpi">
            <strong className={proposals.length > 0 ? "status-dot status-dot--live" : ""}>{proposals.length}</strong>
            <span>Proposal photos</span>
          </div>
          <div className="funnel-kpi">
            <strong className={errors.length > 0 ? "status-dot status-dot--paused" : "status-dot status-dot--idle"}>{errors.length}</strong>
            <span>Errors</span>
          </div>
        </div>
      </section>

      {isLoading && <LoadingState variant="skeleton" message="Loading vision analysis…" />}
      {error && <EmptyState variant="warning" title="Vision data unavailable" message={(error as Error).message} />}

      {!isLoading && !error && rows.length === 0 && (
        <EmptyState variant="empty" title="No vision analyses yet" message="Ring detection and photo classification run on engagement-candidate posts during ingest." />
      )}

      {!isLoading && !error && rows.length > 0 && (
        <table className="dossier-ledger">
          <thead>
            <tr>
              <th>Image</th>
              <th>Ring</th>
              <th>Photo label</th>
              <th>Visual signals</th>
              <th>Model</th>
              <th>Analyzed</th>
            </tr>
          </thead>
          <tbody>
            {rows.map((r: VisionAnalysis) => (
              <tr key={r.id}>
                <td>
                  <a href={r.image_url} target="_blank" rel="noopener noreferrer" className="vision-thumb-link">
                    <img src={r.image_url} alt="" className="vision-thumb" loading="lazy" />
                  </a>
                </td>
                <td>
                  <span className={`status-dot ${ringTone(r.ring_confidence)}`}>
                    {pct(r.ring_confidence)}
                  </span>
                </td>
                <td>
                  {r.photo_label ? (
                    <span>{r.photo_label} <span className="ops-view__muted">({pct(r.photo_confidence)})</span></span>
                  ) : "—"}
                </td>
                <td>
                  <code className="vision-labels">{r.labels || "[]"}</code>
                </td>
                <td><span className="ops-view__muted">{r.model}</span></td>
                <td>{r.created_at?.slice(0, 16)?.replace("T", " ")}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}

export default VisionView;
