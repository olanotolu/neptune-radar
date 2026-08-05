import { useMemo, useState } from "react";
import { useMarriageLicenses } from "../api/hooks";
import type { MarriageLicenseFiling } from "../api/types";
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
const PRIORITY_DOT: Record<MarriageLicenseFiling["priority"], string> = {
  urgent: "ml-dot ml-dot--red",
  priority: "ml-dot ml-dot--amber",
  early: "ml-dot ml-dot--green",
  monitor: "ml-dot ml-dot--dim",
};

function fmtDate(iso: string): string {
  const d = new Date(iso);
  return d.toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric" });
}

export function MarriageLicensesView() {
  const { data: filings, error, isLoading } = useMarriageLicenses();
  const [filter, setFilter] = useState<PriorityFilter>("all");

  const filtered = useMemo(() => {
    if (!filings) return [];
    if (filter === "all") return filings;
    return filings.filter((f) => f.priority === filter);
  }, [filings, filter]);

  // ponytail: sort client-side by days-until-wedding ascending so the most
  // time-critical couples are on top. Ceiling: server-side ORDER BY would scale
  // better past ~10k filings; upgrade path = add ?sort=days to the endpoint.
  const sorted = useMemo(() => {
    return [...filtered].sort((a, b) => {
      const da = a.days_until_wedding ?? 9999;
      const db = b.days_until_wedding ?? 9999;
      return da - db;
    });
  }, [filtered]);

  const counts = useMemo(() => {
    const c: Record<string, number> = { urgent: 0, priority: 0, early: 0, monitor: 0 };
    for (const f of filings ?? []) c[f.priority]++;
    return c;
  }, [filings]);

  if (isLoading) return <LoadingState variant="skeleton" message="Loading marriage license filings…" />;
  if (error) {
    return (
      <EmptyState
        title="Couldn't load filings"
        message={(error as Error).message}
        variant="warning"
      />
    );
  }

  return (
    <div className="view view--marriage-licenses">
      <header className="view__head">
        <h2>Marriage Licenses</h2>
        <p className="ml-sub">
          Public filings — the 30-90 day window before the wedding is the prenup signing moment.
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
          title="No filings yet"
          message="Marriage license filings will appear here once the MarriageSignals feed is connected (MARRIAGE_SIGNALS_URL + MARRIAGE_SIGNALS_KEY)."
          variant="empty"
        />
      ) : (
        <table className="ml-table">
          <thead>
            <tr>
              <th className="ml-table__th">Couple</th>
              <th className="ml-table__th">County</th>
              <th className="ml-table__th">Filed</th>
              <th className="ml-table__th">Predicted wedding</th>
              <th className="ml-table__th ml-table__th--num">Days</th>
              <th className="ml-table__th">Status</th>
            </tr>
          </thead>
          <tbody>
            {sorted.map((f) => (
              <tr key={f.id} className="ml-table__row">
                <td className="ml-table__cell ml-table__cell--names">
                  {f.person_a_name} <span className="ml-table__amp">&amp;</span> {f.person_b_name}
                </td>
                <td className="ml-table__cell ml-table__cell--county">{f.county}</td>
                <td className="ml-table__cell ml-table__cell--date">{fmtDate(f.filing_date)}</td>
                <td className="ml-table__cell ml-table__cell--date">
                  {f.wedding_date ? fmtDate(f.wedding_date) : fmtDate(f.predicted_wedding_date)}
                </td>
                <td className="ml-table__cell ml-table__cell--num">
                  {f.days_until_wedding != null ? f.days_until_wedding : "—"}
                </td>
                <td className="ml-table__cell ml-table__cell--status">
                  <span className={PRIORITY_DOT[f.priority]} aria-hidden />
                  <span className="ml-table__priority">{f.priority}</span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  );
}
