import { useExperimentResults } from "../api/hooks";
import type { VariantResult } from "../api/types";
import { EmptyState } from "../components/EmptyState";
import { LoadingState } from "../components/LoadingState";

function pct(n: number): string {
  return `${Math.round((n || 0) * 100)}%`;
}

export function ExperimentsView() {
  const { data, isLoading } = useExperimentResults("postcard_copy_v1");

  if (isLoading) return <LoadingState variant="skeleton" message="Loading experiments…" />;
  if (!data || data.variants.length === 0)
    return (
      <EmptyState
        variant="empty"
        title="No experiment data yet"
        message="A/B variant results will appear once postcards are mailed and QR scans / chats are recorded."
      />
    );

  const totalMailed = data.variants.reduce((s, v) => s + v.mailed, 0);

  return (
    <div className="experiments-view">
      <header className="experiments-view__header">
        <h1>{data.experiment_name}</h1>
        <span className="experiments-view__subtitle">
          {data.experiment_id} · {totalMailed} mailed
        </span>
      </header>

      {data.winner_id ? (
        <div className="experiments-view__winner">
          Leading variant: <strong>{data.winner_id}</strong>
        </div>
      ) : null}

      <table className="experiments-table">
        <thead>
          <tr>
            <th>Variant</th>
            <th>Mailed</th>
            <th>Scans</th>
            <th>Scan Rate</th>
            <th>Chats</th>
            <th>Conversion Rate</th>
          </tr>
        </thead>
        <tbody>
          {data.variants.map((v: VariantResult) => (
            <tr
              key={v.variant_id}
              className={data.winner_id === v.variant_id ? "experiments-table__row--winner" : ""}
            >
              <td>
                <strong>{v.variant_id}</strong>
                <span className="experiments-table__name"> {v.variant_name}</span>
              </td>
              <td>{v.mailed}</td>
              <td>{v.scans}</td>
              <td>{pct(v.scan_rate)}</td>
              <td>{v.chats}</td>
              <td>{pct(v.conversion_rate)}</td>
            </tr>
          ))}
        </tbody>
      </table>

      <p className="experiments-view__note">
        Winner requires ≥10 mailed per variant. Conversion = chats / mailed.
      </p>
    </div>
  );
}
