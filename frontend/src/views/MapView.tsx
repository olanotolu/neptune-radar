import { useEffect, useRef, useState, useCallback } from "react";
import { geoAlbersUsa, geoPath } from "d3-geo";
import { select } from "d3-selection";
import { zoom, zoomIdentity, type ZoomBehavior } from "d3-zoom";
import "d3-transition";
import { feature } from "topojson-client";
import topology from "us-atlas/states-10m.json";
import { getCities } from "./cities";
import { OHIO_CITY_COORDS, OHIO_COUNTIES, OHIO_FIPS, type LayerId } from "./ohio";
import { OhioPanel } from "./OhioPanel";
import { useOhioChurches, useOhioGovernment, useOhioOverview, useOhioSocial, useProspectPins } from "../api/hooks";
import { mediaURL } from "../api/media";
import type { ConnectorStatus, ProspectPin } from "../api/types";

// us-atlas identifies states by numeric FIPS code, which is meaningless on
// screen and — critically — does NOT match the USPS abbreviations cities.ts
// is keyed by. Every lookup (on-map label, city panel, tooltip) needs to go
// through this table, not the raw FIPS id.
const FIPS_TO_ABBREV: Record<number, string> = {
  1: "AL", 2: "AK", 4: "AZ", 5: "AR", 6: "CA", 8: "CO", 9: "CT", 10: "DE",
  11: "DC", 12: "FL", 13: "GA", 15: "HI", 16: "ID", 17: "IL", 18: "IN",
  19: "IA", 20: "KS", 21: "KY", 22: "LA", 23: "ME", 24: "MD", 25: "MA",
  26: "MI", 27: "MN", 28: "MS", 29: "MO", 30: "MT", 31: "NE", 32: "NV",
  33: "NH", 34: "NJ", 35: "NM", 36: "NY", 37: "NC", 38: "ND", 39: "OH",
  40: "OK", 41: "OR", 42: "PA", 44: "RI", 45: "SC", 46: "SD", 47: "TN",
  48: "TX", 49: "UT", 50: "VT", 51: "VA", 53: "WA", 54: "WV", 55: "WI",
  56: "WY", 60: "AS", 66: "GU", 69: "MP", 72: "PR", 78: "VI",
};

const geoFeature = feature(topology as any, topology.objects.states as any) as any;
const STATE_MAP: Record<string, string> = {};
const STATE_ABBREV: Record<string, string> = {};
const STATES = geoFeature.features.map((f: any) => {
  const abbrev = FIPS_TO_ABBREV[Number(f.id)] ?? "?";
  STATE_MAP[f.id] = f.properties.name;
  STATE_ABBREV[f.id] = abbrev;
  return { id: f.id, abbrev, name: f.properties.name, geometry: f.geometry };
});

const WIDTH = 960;
const HEIGHT = 620;
// Panel overlays the right side of the map, so zoom-to-state fits into the
// remaining left viewport instead of the full width (otherwise the panel
// covers the very state you clicked).
const PANEL_W = 356;

const projection = geoAlbersUsa().fitSize([WIDTH, HEIGHT], geoFeature);
const path = geoPath(projection);

// Precompute every path string + label centroid ONCE at module scope.
// Before this, every mousemove/zoom-frame re-render recomputed geoPath for
// all 56 states and 88 Ohio counties — dozens of times per second.
const STATE_PATHS: Record<string, string> = {};
const STATE_CENTROIDS: Record<string, [number, number]> = {};
for (const s of STATES) {
  const d = path(s.geometry);
  if (d) STATE_PATHS[s.id] = d;
  const c = path.centroid(s.geometry);
  if (c && !isNaN(c[0]) && !isNaN(c[1])) STATE_CENTROIDS[s.id] = c;
}
const COUNTY_PATHS: Record<string, string> = {};
for (const cty of OHIO_COUNTIES) {
  const d = path(cty.geometry);
  if (d) COUNTY_PATHS[cty.fips] = d;
}

const TIER_COLORS: Record<number, { bg: string; border: string; text: string }> = {
  1: { bg: "var(--cove)", border: "#b8e8ea", text: "var(--cove-deep)" },
  2: { bg: "var(--mesa-soft)", border: "#f5d4c0", text: "var(--mesa-deep)" },
  3: { bg: "var(--bg-alt)", border: "var(--border)", text: "var(--ink-dim)" },
};

const SIGNAL_DOT: Record<string, string> = {
  Active: "var(--cove-deep)",
  Pending: "var(--mesa)",
  Idle: "var(--ink-faint)",
};

function TierBadge({ tier }: { tier: 1 | 2 | 3 }) {
  const label = tier === 1 ? "Tier 1" : tier === 2 ? "Tier 2" : "Tier 3";
  const c = TIER_COLORS[tier];
  return (
    <span className="city-tier" style={{ background: c.bg, color: c.text, borderColor: c.border }}>
      {label}
    </span>
  );
}

function SignalDot({ signal }: { signal: string }) {
  const color = SIGNAL_DOT[signal] ?? "var(--ink-faint)";
  return <span className={`city-signal-dot ${signal === "Active" ? "city-signal-dot--live" : ""}`} style={{ background: color }} title={signal} />;
}

function StatePanel({ stateId, stateName, onClose }: { stateId: string; stateName: string; onClose: () => void }) {
  const cities = getCities(stateId);
  const tiers = ([1, 2, 3] as const).map((t) => ({ tier: t, rows: cities.filter((c) => c.tier === t) }));
  const active = cities.filter((c) => c.signal === "Active").length;

  return (
    <aside className="state-panel">
      <header className="state-panel__header">
        <div className="state-panel__title-row">
          <div>
            <h3 className="state-panel__title">{stateName}</h3>
            <span className="state-panel__code">{stateId}</span>
          </div>
          <button className="state-panel__close" onClick={onClose} aria-label="Close panel">×</button>
        </div>
        <div className="state-panel__stats">
          <div className="state-panel__stat">
            <span className="state-panel__stat-value">{cities.length}</span>
            <span className="state-panel__stat-label">cities</span>
          </div>
          <div className="state-panel__stat">
            <span className="state-panel__stat-value">{active}</span>
            <span className="state-panel__stat-label">active</span>
          </div>
          <div className="state-panel__stat">
            <span className="state-panel__stat-value">{tiers[0].rows.length}</span>
            <span className="state-panel__stat-label">tier 1</span>
          </div>
        </div>
      </header>

      <div className="state-panel__body">
        {tiers.map(({ tier, rows }) => rows.length > 0 && (
          <section className="state-panel__section" key={tier}>
            <div className="state-panel__section-header">
              <span className="state-panel__section-title">
                {tier === 1 ? "Primary" : tier === 2 ? "Secondary" : "Emerging"}
              </span>
              <span className="state-panel__section-count">{rows.length}</span>
            </div>
            <ul className="city-list">
              {rows.map((c, i) => (
                <li key={c.city} className={`city-row city-row--tier${tier}`} style={{ animationDelay: `${i * 30}ms` }}>
                  <span className="city-row__bar" />
                  <div className="city-row__main">
                    <div className="city-row__top">
                      <span className="city-row__name">{c.city}</span>
                      <SignalDot signal={c.signal} />
                    </div>
                    <div className="city-row__bottom">
                      <span className="city-row__pop">{c.pop}</span>
                      <TierBadge tier={c.tier} />
                    </div>
                  </div>
                </li>
              ))}
            </ul>
          </section>
        ))}

        {cities.length === 0 && (
          <div className="state-panel__empty">
            No monitored markets in this state yet.
          </div>
        )}
      </div>

      <footer className="state-panel__footer">
        <span className="state-panel__hint">Click ocean or press esc to reset</span>
      </footer>
    </aside>
  );
}

export function MapView() {
  const [selected, setSelected] = useState<string | null>(null);
  const [hovered, setHovered] = useState<string | null>(null);
  // Cursor position lives in a ref + direct DOM style, NOT state — a
  // mousemove-per-frame setState was re-rendering the entire SVG map
  // (all 144 precomputed paths) on every pixel of mouse travel.
  const tipRef = useRef<HTMLDivElement>(null);
  const [tipVisible, setTipVisible] = useState(false);
  const [zoomLevel, setZoomLevel] = useState(1);
  const [showProspects, setShowProspects] = useState(true);
  const [selectedPin, setSelectedPin] = useState<ProspectPin | null>(null);
  const svgRef = useRef<SVGSVGElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const zoomRef = useRef<ZoomBehavior<SVGSVGElement, unknown> | null>(null);

  // Ohio drill-down: layer tabs + which county/diocese is inspected. Only
  // meaningful when selected === OHIO_FIPS; reset alongside the selection.
  const [ohioLayer, setOhioLayer] = useState<LayerId>("overview");
  const [selectedCountyFips, setSelectedCountyFips] = useState<string | null>(null);
  const [selectedDioceseId, setSelectedDioceseId] = useState<string | null>(null);
  const isOhio = selected === OHIO_FIPS;

  const ohioOverview = useOhioOverview(isOhio);
  const ohioGovernment = useOhioGovernment(isOhio);
  const ohioChurches = useOhioChurches(isOhio);
  const ohioSocial = useOhioSocial(isOhio);
  const prospectPins = useProspectPins(showProspects);

  const resetZoom = useCallback((clearSelection = true) => {
    if (svgRef.current && zoomRef.current) {
      select(svgRef.current)
        .transition()
        .duration(650)
        .call(zoomRef.current.transform, zoomIdentity);
    }
    if (clearSelection) {
      setSelected(null);
      setOhioLayer("overview");
      setSelectedCountyFips(null);
      setSelectedDioceseId(null);
    }
  }, []);

  useEffect(() => {
    if (!svgRef.current) return;

    const svg = select(svgRef.current);
    const g = svg.select<SVGGElement>(".map-viewport");

    const z = zoom<SVGSVGElement, unknown>()
      .scaleExtent([0.75, 10])
      .on("zoom", (event) => {
        g.attr("transform", event.transform);
        // Throttle re-renders: label/pin counter-scaling can't visually
        // resolve sub-percent zoom changes anyway.
        setZoomLevel((prev) => {
          const next = Math.round(event.transform.k * 100) / 100;
          return prev === next ? prev : next;
        });
      });

    svg.call(z);
    zoomRef.current = z;

    return () => {
      svg.on(".zoom", null);
    };
  }, []);

  // Esc closes the panel — same as clicking the ocean.
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") resetZoom();
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [resetZoom]);

  const zoomToState = useCallback((stateId: string) => {
    const state = STATES.find((s: any) => s.id === stateId);
    if (!state || !svgRef.current || !zoomRef.current) return;

    const [[x0, y0], [x1, y1]] = path.bounds(state.geometry);
    const dx = x1 - x0;
    const dy = y1 - y0;
    if (dx === 0 || dy === 0) return;

    // Fit into the viewport left of the panel so the selected state is
    // never hidden behind it.
    const availW = WIDTH - PANEL_W;
    const scale = Math.min(8, 0.82 / Math.max(dx / availW, dy / HEIGHT));
    const cx = (x0 + x1) / 2;
    const cy = (y0 + y1) / 2;
    const tx = availW / 2 - scale * cx;
    const ty = HEIGHT / 2 - scale * cy;

    select(svgRef.current)
      .transition()
      .duration(750)
      .call(zoomRef.current.transform, zoomIdentity.translate(tx, ty).scale(scale));
  }, []);

  const handleStateClick = useCallback((stateId: string) => {
    if (selected === stateId) {
      resetZoom();
    } else {
      setSelected(stateId);
      setOhioLayer("overview");
      setSelectedCountyFips(null);
      setSelectedDioceseId(null);
      zoomToState(stateId);
    }
  }, [selected, resetZoom, zoomToState]);

  const handleZoomIn = useCallback(() => {
    if (svgRef.current && zoomRef.current) {
      select(svgRef.current).transition().duration(250).call(zoomRef.current.scaleBy, 1.5);
    }
  }, []);

  const handleZoomOut = useCallback(() => {
    if (svgRef.current && zoomRef.current) {
      select(svgRef.current).transition().duration(250).call(zoomRef.current.scaleBy, 0.67);
    }
  }, []);

  const handleMouseMove = useCallback((e: React.MouseEvent) => {
    const rect = containerRef.current?.getBoundingClientRect();
    if (!rect || !tipRef.current) return;
    tipRef.current.style.left = `${e.clientX - rect.left + 14}px`;
    tipRef.current.style.top = `${e.clientY - rect.top + 14}px`;
  }, []);

  const hoveredAbbrev = hovered ? STATE_ABBREV[hovered] : null;
  const hoveredCities = hoveredAbbrev ? getCities(hoveredAbbrev) : [];

  return (
    <div className="view view--map">
      <div className="map-frame" ref={containerRef} onMouseMove={handleMouseMove}>
        <div className="map-frame__topbar">
          <div className="map-frame__title-group">
            <h2 className="map-frame__title">Coverage map</h2>
            <span className="map-frame__meta">
              {selected ? STATE_MAP[selected] : "50 states · 3 tiers"}
            </span>
          </div>
          <div className="map-controls">
            <button
              className={`map-btn ${showProspects ? "map-btn--active" : ""}`}
              onClick={() => setShowProspects((v) => !v)}
              title="Toggle prospect pins from bio/post location"
            >
              Prospects {prospectPins.data?.length ?? 0}
            </button>
            <button className="map-btn" onClick={handleZoomIn} title="Zoom in">+</button>
            <button className="map-btn" onClick={handleZoomOut} title="Zoom out">−</button>
            <button className="map-btn" onClick={() => resetZoom()} title="Reset view">⟲</button>
            <span className="map-zoom-level">{Math.round(zoomLevel * 100)}%</span>
          </div>
        </div>

        <div className="map-stage">
          <svg
            ref={svgRef}
            viewBox={`0 0 ${WIDTH} ${HEIGHT}`}
            className="map-svg"
            onMouseLeave={() => { setHovered(null); setTipVisible(false); }}
          >
            <g className={`map-viewport ${selected ? "map-viewport--has-selection" : ""}`}>
              {/* Invisible backdrop: clicking ocean resets the view. */}
              <rect
                className="map-ocean"
                width={WIDTH}
                height={HEIGHT}
                onClick={() => resetZoom()}
              />
              <g>
                {STATES.map((state: any) => {
                  const d = STATE_PATHS[state.id];
                  if (!d) return null;
                  return (
                    <path
                      key={state.id}
                      d={d}
                      className={`map-state ${selected === state.id ? "map-state--selected" : ""} ${hovered === state.id ? "map-state--hovered" : ""}`}
                      onClick={() => handleStateClick(state.id)}
                      onMouseEnter={(e) => {
                        setHovered(state.id);
                        setTipVisible(true);
                        // Position immediately — without this the tooltip
                        // flashes at 0,0 for one frame before mousemove.
                        const rect = containerRef.current?.getBoundingClientRect();
                        if (rect && tipRef.current) {
                          tipRef.current.style.left = `${e.clientX - rect.left + 14}px`;
                          tipRef.current.style.top = `${e.clientY - rect.top + 14}px`;
                        }
                      }}
                    />
                  );
                })}
              </g>
              <g pointerEvents="none">
                {STATES.map((state: any) => {
                  const centroid = STATE_CENTROIDS[state.id];
                  if (!centroid) return null;
                  return (
                    <text
                      key={`label-${state.id}`}
                      x={centroid[0]}
                      y={centroid[1]}
                      className={`map-label ${selected === state.id ? "map-label--selected" : ""}`}
                      textAnchor="middle"
                      dominantBaseline="central"
                      // Labels live inside the zoomed <g>, so their rendered
                      // size scales with the map unless counter-scaled here —
                      // without this they balloon and overlap at high zoom.
                      style={{ fontSize: `${9.5 / zoomLevel}px` }}
                    >
                      {state.abbrev}
                    </text>
                  );
                })}
              </g>

              {showProspects && (
                <g className="prospect-pin-layer">
                  {(prospectPins.data ?? [])
                    .filter((p) => p.lat != null && p.lng != null)
                    .map((p) => {
                      const pt = projection([p.lng!, p.lat!]);
                      if (!pt) return null;
                      const r = Math.max(4, 6 / zoomLevel);
                      return (
                        <g
                          key={p.couple_id}
                          className={`prospect-pin ${selectedPin?.couple_id === p.couple_id ? "prospect-pin--selected" : ""}`}
                          transform={`translate(${pt[0]},${pt[1]})`}
                          onClick={(e) => {
                            e.stopPropagation();
                            setSelectedPin(p);
                          }}
                          style={{ cursor: "pointer" }}
                        >
                          <circle r={r * 2.2} className="prospect-pin__halo" />
                          <circle r={r} className="prospect-pin__dot" />
                          <title>
                            {p.person_a_label} & {p.person_b_label} — {p.city}
                            {p.region ? `, ${p.region}` : ""}
                          </title>
                        </g>
                      );
                    })}
                </g>
              )}

              {isOhio && ohioLayer === "government" && (
                <g className="ohio-county-layer">
                  {OHIO_COUNTIES.map((county) => {
                    const d = COUNTY_PATHS[county.fips];
                    if (!d) return null;
                    const row = ohioGovernment.data?.find((r) => r.county.id === county.fips);
                    const status: ConnectorStatus | "not_configured" = row?.connector?.status ?? "not_configured";
                    return (
                      <path
                        key={county.fips}
                        d={d}
                        className={`ohio-county ohio-county--${status} ${selectedCountyFips === county.fips ? "ohio-county--selected" : ""}`}
                        onClick={(e) => {
                          e.stopPropagation();
                          setSelectedCountyFips(county.fips === selectedCountyFips ? null : county.fips);
                        }}
                      />
                    );
                  })}
                </g>
              )}

              {isOhio && ohioLayer === "churches" && (
                <g className="ohio-marker-layer">
                  {/* Dioceses group by hub-city coordinate. Only Columbus has
                      verified coords today, so N dioceses share one point —
                      render ONE marker with a count badge instead of N
                      identical stacked circles where only the top was
                      clickable. Clicking opens the diocese list (no single
                      selection — they're all at the same place). */}
                  {(() => {
                    const groups = new Map<string, typeof ohioChurches.data>();
                    for (const d of ohioChurches.data ?? []) {
                      const key = OHIO_CITY_COORDS.Columbus.join(",");
                      groups.set(key, [...(groups.get(key) ?? []), d]);
                    }
                    return [...groups.values()].map((group) => {
                      const [lng, lat] = OHIO_CITY_COORDS.Columbus;
                      const p = projection([lng, lat]);
                      if (!p || !group) return null;
                      const anySelected = group.some((d) => d.jurisdiction.id === selectedDioceseId);
                      return (
                        <g
                          key={group.map((d) => d.jurisdiction.id).join("+")}
                          className={`ohio-diocese-marker ${anySelected ? "ohio-diocese-marker--selected" : ""}`}
                          transform={`translate(${p[0]},${p[1]})`}
                          onClick={(e) => {
                            e.stopPropagation();
                            setSelectedDioceseId(null);
                          }}
                          style={{ cursor: "pointer" }}
                        >
                          <circle r={11 / zoomLevel} />
                          <text
                            textAnchor="middle"
                            dominantBaseline="central"
                            style={{ fontSize: `${9 / zoomLevel}px` }}
                            className="ohio-diocese-marker__count"
                          >
                            {group.length}
                          </text>
                        </g>
                      );
                    });
                  })()}
                </g>
              )}

              {isOhio && ohioLayer === "instagram" && (ohioSocial.data?.vendors?.length ?? 0) > 0 && (
                <g className="ohio-marker-layer">
                  {/* One marker per city market that actually has vendors —
                      driven by the API response, not a static decoration. */}
                  {(() => {
                    const vendors = ohioSocial.data?.vendors ?? [];
                    const p = projection(OHIO_CITY_COORDS.Columbus);
                    if (!p) return null;
                    return (
                      <g
                        transform={`translate(${p[0]},${p[1]})`}
                        className="ohio-instagram-marker"
                      >
                        <circle r={11 / zoomLevel} />
                        <text
                          textAnchor="middle"
                          dominantBaseline="central"
                          style={{ fontSize: `${9 / zoomLevel}px` }}
                          className="ohio-diocese-marker__count"
                        >
                          {vendors.length}
                        </text>
                      </g>
                    );
                  })()}
                </g>
              )}
            </g>
          </svg>

          {/* Cursor-following tooltip — position set imperatively via ref */}
          {hovered && !selected && tipVisible && (
            <div ref={tipRef} className="map-cursor-tip">
              <span className="map-cursor-tip__name">{STATE_MAP[hovered]}</span>
              <span className="map-cursor-tip__count">
                {hoveredCities.length} {hoveredCities.length === 1 ? "market" : "markets"}
              </span>
            </div>
          )}

          {/* Legend */}
          <div className="map-legend">
            <span className="map-legend__item"><i className="map-legend__swatch map-legend__swatch--t1" /> Tier 1</span>
            <span className="map-legend__item"><i className="map-legend__swatch map-legend__swatch--t2" /> Tier 2</span>
            <span className="map-legend__item"><i className="map-legend__swatch map-legend__swatch--t3" /> Tier 3</span>
            {showProspects && (
              <span className="map-legend__item">
                <i className="map-legend__swatch" style={{ background: "var(--mesa)" }} /> Prospects
              </span>
            )}
          </div>

          {selectedPin && (
            <div className="prospect-pin-panel">
              <div className="prospect-pin-panel__header">
                <h3 className="prospect-pin-panel__title">
                  {selectedPin.person_a_label} & {selectedPin.person_b_label}
                </h3>
                <button className="btn btn--ghost btn--sm" type="button" onClick={() => setSelectedPin(null)}>
                  Close
                </button>
              </div>
              <div className="prospect-pin-panel__loc">
                {selectedPin.city}
                {selectedPin.region ? `, ${selectedPin.region}` : ""}
                {selectedPin.stage ? ` · ${selectedPin.stage.replace(/_/g, " ")}` : ""}
              </div>
              <div className="prospect-card__faces" style={{ marginTop: 10 }}>
                {selectedPin.profile_pic_a ? (
                  <img className="prospect-card__avatar" src={mediaURL(selectedPin.profile_pic_a)} alt="" referrerPolicy="no-referrer" />
                ) : null}
                {selectedPin.profile_pic_b ? (
                  <img className="prospect-card__avatar" src={mediaURL(selectedPin.profile_pic_b)} alt="" referrerPolicy="no-referrer" />
                ) : null}
              </div>
              {(selectedPin.handle_a || selectedPin.handle_b) && (
                <div className="prospect-card__meta" style={{ marginTop: 8 }}>
                  {selectedPin.handle_a && <span>@{selectedPin.handle_a}</span>}
                  {selectedPin.handle_b && <span>@{selectedPin.handle_b}</span>}
                </div>
              )}
            </div>
          )}

          {/* Overlay panel — no page scroll needed */}
          {selected && isOhio && (
            <OhioPanel
              layer={ohioLayer}
              onLayerChange={setOhioLayer}
              onClose={() => resetZoom()}
              overview={ohioOverview}
              government={ohioGovernment}
              churches={ohioChurches}
              social={ohioSocial}
              selectedCountyFips={selectedCountyFips}
              onSelectCounty={setSelectedCountyFips}
              selectedDioceseId={selectedDioceseId}
              onSelectDiocese={setSelectedDioceseId}
            />
          )}
          {selected && !isOhio && (
            <StatePanel
              stateId={STATE_ABBREV[selected]}
              stateName={STATE_MAP[selected]}
              onClose={() => resetZoom()}
            />
          )}
        </div>
      </div>
    </div>
  );
}
