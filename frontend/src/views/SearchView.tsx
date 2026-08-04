import { useEffect, useMemo, useState } from "react";
import { useSearch } from "../api/hooks";
import type { CRMLead, NeptuneCase } from "../api/types";
import { EmptyState } from "../components/EmptyState";
import { LoadingState } from "../components/LoadingState";

type EntityTab = "all" | "couples" | "leads" | "cases";

function useHashParams(): URLSearchParams {
  const [params, setParams] = useState<URLSearchParams>(() => {
    const raw = window.location.hash.replace(/^#/, "");
    const [, q] = raw.split("?");
    return new URLSearchParams(q || "");
  });
  useEffect(() => {
    const onHash = () => {
      const raw = window.location.hash.replace(/^#/, "");
      const [, q] = raw.split("?");
      setParams(new URLSearchParams(q || ""));
    };
    window.addEventListener("hashchange", onHash);
    return () => window.removeEventListener("hashchange", onHash);
  }, []);
  return params;
}

export function SearchView() {
  const params = useHashParams();
  const q = params.get("q") ?? "";
  const typeParam = (params.get("type") as EntityTab) ?? "all";

  const [tab, setTab] = useState<EntityTab>(typeParam);
  const [minConfidence, setMinConfidence] = useState(0);

  useEffect(() => setTab(typeParam), [typeParam]);

  const { data, isLoading, error } = useSearch(q, "all");

  const couples = data?.couples ?? [];
  const leads = data?.leads ?? [];
  const cases = data?.cases ?? [];

  // ponytail: CRMLead/NeptuneCase don't carry a confidence field today, so the
  // slider is a no-op until the search endpoint enriches results. Couples have
  // no score either. Kept so the UI is ready when the API adds confidence.
  const conf = (o: unknown): number => (o as { confidence?: number }).confidence ?? 0;
  const filteredCouples = useMemo(
    () => couples.filter((c) => conf(c) >= minConfidence),
    [couples, minConfidence],
  );
  const filteredLeads = useMemo(
    () => leads.filter((l) => conf(l) >= minConfidence),
    [leads, minConfidence],
  );
  const filteredCases = useMemo(
    () => cases.filter((c) => conf(c) >= minConfidence),
    [cases, minConfidence],
  );

  const total = filteredCouples.length + filteredLeads.length + filteredCases.length;

  return (
    <div className="view view--search">
      <header className="search-view__header">
        <h2 className="search-view__title">
          Search{q ? <span className="search-view__q"> · {q}</span> : null}
        </h2>
        <p className="search-view__subtitle">Find any couple, lead, or case across the system.</p>
        <div className="search-view__controls">
          <label className="search-view__filter">
            <span className="sr-only">Entity type</span>
            <select value={tab} onChange={(e) => setTab(e.target.value as EntityTab)}>
              <option value="all">All</option>
              <option value="couples">Couples</option>
              <option value="leads">Leads</option>
              <option value="cases">Cases</option>
            </select>
          </label>
          <label className="search-view__slider">
            <span>Min confidence</span>
            <input
              type="range"
              min={0}
              max={1}
              step={0.05}
              value={minConfidence}
              onChange={(e) => setMinConfidence(Number(e.target.value))}
            />
            <span className="search-view__slider-val">
              {Math.round(minConfidence * 100)}%
            </span>
          </label>
        </div>
      </header>

      {!q && <EmptyState variant="empty" title="Search Neptune" message="Find any couple, lead, or case. Type a query above and press Enter." />}
      {q && isLoading && <LoadingState variant="skeleton" message="Searching…" />}
      {q && error && <EmptyState variant="warning" title="Search failed" message={(error as Error).message} />}

      {q && !isLoading && !error && total === 0 && (
        <EmptyState variant="empty" title="No results" message={`No matches for “${q}”. Try a different query or lower the confidence filter.`} />
      )}

      {q && !isLoading && !error && total > 0 && (
        <div className="search-results">
          {(tab === "all" || tab === "couples") && filteredCouples.length > 0 && (
            <section className="search-section">
              <h3 className="search-section__title">Couples ({filteredCouples.length})</h3>
              <ul className="search-list">
                {filteredCouples.map((c) => (
                  <li key={c.id}>
                    <a className="search-list__row" href={`#/dossier/${encodeURIComponent(c.id)}`}>
                      <span className="search-list__label">
                        {c.person_a_label} &amp; {c.person_b_label}
                      </span>
                      <span className="search-list__id">{c.id.slice(0, 12)}…</span>
                    </a>
                  </li>
                ))}
              </ul>
            </section>
          )}

          {(tab === "all" || tab === "leads") && filteredLeads.length > 0 && (
            <section className="search-section">
              <h3 className="search-section__title">Leads ({filteredLeads.length})</h3>
              <ul className="search-list">
                {filteredLeads.map((l: CRMLead) => (
                  <li key={l.id}>
                    <a className="search-list__row" href={`#/case`}>
                      <span className="search-list__label">{l.lead_type} · {l.status}</span>
                      <span className="search-list__id">{l.id.slice(0, 12)}…</span>
                    </a>
                  </li>
                ))}
              </ul>
            </section>
          )}

          {(tab === "all" || tab === "cases") && filteredCases.length > 0 && (
            <section className="search-section">
              <h3 className="search-section__title">Cases ({filteredCases.length})</h3>
              <ul className="search-list">
                {filteredCases.map((c: NeptuneCase) => (
                  <li key={c.id}>
                    <a className="search-list__row" href={`#/case`}>
                      <span className="search-list__label">{c.case_type} · {c.status}</span>
                      <span className="search-list__id">{c.id.slice(0, 12)}…</span>
                    </a>
                  </li>
                ))}
              </ul>
            </section>
          )}
        </div>
      )}
    </div>
  );
}
