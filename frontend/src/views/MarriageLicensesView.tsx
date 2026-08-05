import { useMemo, useState } from "react";
import { useMarriageLicenses, useWeddingPredictions } from "../api/hooks";
import type { MarriageLicenseFiling, WeddingPrediction } from "../api/types";
import { EmptyState } from "../components/EmptyState";
import { LoadingState } from "../components/LoadingState";

type PriorityFilter = "all" | "urgent" | "priority" | "early" | "monitor";

const PRIORITY_CHIPS: { key: PriorityFilter; label: string }[] = [
  { key: "all", label: "All" },
  { key: "urgent", label: "Urgent" },
  { key: "priority", label: "Priority" },
  { key: "early", label: "Early" },
  { key: "monitor", label: "Monitor" },
];

// ponytail: priority dot color is the only functional color in the view —
// urgent=red, priority=amber, early=green, monitor=ink-dim. Matches the
// DESIGN.md status-dot palette exactly.
const PRIORITY_DOT: Record<WeddingPrediction["priority"], string> = {
  urgent: "ml-dot ml-dot--red",
  priority: "ml-dot ml-dot--amber",
  early: "ml-dot ml-dot--green",
  monitor: "ml-dot ml-dot--dim",
};

function fmtDate(iso: string): string {
  const d = new Date(iso);
  return d.toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric" });
}

// A unified row shape so the license feed and the social-prediction feed render
// in one table. ponytail: merging client-side is cheaper than a new backend
// join when both lists are small; ceiling = ~10k rows, upgrade = server union.
type UnifiedRow = {
  id: string;
  person_a_name: string;
  person_b_name: string;
  county: string;
  predicted_wedding_date: string;
  wedding_date?: string;
  days_until_wedding?: number;
  priority: WeddingPrediction["priority"];
  prediction_source: "marriage_license" | "social";
  social_reason?: string;
  social_confidence?: number;
};

function fromLicense(f: MarriageLicenseFiling): UnifiedRow {
  return { ...f, prediction_source: "marriage_license" };
}

function fromPrediction(p: WeddingPrediction): UnifiedRow {
  return p;
}

export function MarriageLicensesView() {
  const { data: filings, error: licErr, isLoading: licLoading } = useMarriageLicenses();
  const { data: predictions, error: predErr, isLoading: predLoading } = useWeddingPredictions();
  const [filter, setFilter] = useState<PriorityFilter>("all");

  // Merge both sources, deduping by couple id (a couple may appear in both the
  // license feed and the predictions union). The license row wins on conflict —
  // it carries the filing_date + county that the social row lacks.
  const merged = useMemo(() => {
    const byId = new Map<string, UnifiedRow>();
    for (const f of filings ?? []) byId.set(f.id, fromLicense(f));
    for (const p of predictions ?? []) {
      if (!byId.has(p.id)) byId.set(p.id, fromPrediction(p));
    }
    return [...byId.values()];
  }, [filings, predictions]);

  const filtered = useMemo(() => {
    if (filter === "all") return merged;
    return merged.filter((r) => r.priority === filter);
  }, [merged, filter]);

  // ponytail: sort client-side by days-until-wedding ascending so the most
  // time-critical couples are on top. Ceiling: server-side ORDER BY would scale
  // better past ~10k rows; upgrade path = add ?sort=days to the endpoint.
  const sorted = useMemo(() => {
    return [...filtered].sort((a, b) => {
      const da = a.days_until_wedding ?? 9999;
      const db = b.days_until_wedding ?? 9999;
      return da - db;
    });
  }, [filtered]);

  const counts = useMemo(() => {
    const c: Record<string, number> = { urgent: 0, priority: 0, early: 0, monitor: 0 };
    for (const r of merged) c[r.priority]++;
    return c;
  }, [merged]);

  const isLoading = licLoading || predLoading;
  const error = licErr || predErr;

  if (isLoading) return <LoadingState variant="skeleton" message="Loading wedding predictions…" />;
  if (error) {
    return (
      <EmptyState
        title="Couldn't load predictions"
        message={(error as Error).message}
        variant="warning"
      />
    );
  }

  return (
    <div className="view view--marriage-licenses">
      <header className="view__head">
        <h2>Wedding Predictions</h2>
        <p className="ml-sub">
          Predicted wedding dates from public marriage-license filings and Instagram signals —
          the 30-60 day window before the wedding is the prenup signing moment.
        </p>
      </header>

      <div className="ml-chips" role="tablist" aria-label="Priority filter">
        {PRIORITY_CHIPS.map((c) => (
          <button
            key={c.key}
            type="button"
            role="tab"
            aria-selected={filter === c.key}
            className={`ml-chip${filter === c.key ? " ml-chip--active" : ""}`}
            onClick={() => setFilter(c.key)}
          >
            {c.label}
            {c.key !== "all" && counts[c.key] > 0 && <span className="ml-chip__count">{counts[c.key]}</span>}
          </button>
        ))}
      </div>

      {sorted.length === 0 ? (
        <EmptyState
          title="No predictions yet"
          message="Wedding predictions will appear here once marriage-license filings arrive or Instagram signals indicate a wedding date."
          variant="empty"
        />
      ) : (
        <table className="ml-table">
          <thead>
            <tr>
              <th className="ml-table__th">Couple</th>
              <th className="ml-table__th">County</th>
              <th className="ml-table__th">Predicted wedding</th>
              <th className="ml-table__th ml-table__th--num">Days</th>
              <th className="ml-table__th">Source</th>
              <th className="ml-table__th">Status</th>
            </tr>
          </thead>
          <tbody>
            {sorted.map((r) => (
              <tr key={r.id} className="ml-table__row">
                <td className="ml-table__cell ml-table__cell--names">
                  {r.person_a_name} <span className="ml-table__amp">&amp;</span> {r.person_b_name}
                </td>
                <td className="ml-table__cell ml-table__cell--county">{r.county || "—"}</td>
                <td className="ml-table__cell ml-table__cell--date">
                  {r.wedding_date ? fmtDate(r.wedding_date) : fmtDate(r.predicted_wedding_date)}
                </td>
                <td className="ml-table__cell ml-table__cell--num">
                  {r.days_until_wedding != null ? r.days_until_wedding : "—"}
                </td>
                <td className="ml-table__cell ml-table__cell--source">
                  <span
                    className={`ml-source-badge${r.prediction_source === "social" ? " ml-source-badge--social" : ""}`}
                    title={r.social_reason || (r.prediction_source === "social" ? "Inferred from Instagram signals" : "Public marriage-license filing")}
                  >
                    {r.prediction_source === "social"
                      ? `Social${r.social_confidence ? ` · ${Math.round(r.social_confidence * 100)}%` : ""}`
                      : "License"}
                  </span>
                </td>
                <td className="ml-table__cell ml-table__cell--status">
                  <span className={PRIORITY_DOT[r.priority]} aria-hidden />
                  <span className="ml-table__priority">{r.priority}</span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
