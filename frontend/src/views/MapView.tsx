import { useEffect, useRef, useState, useCallback, useMemo, type ReactNode } from "react";
import { geoAlbersUsa, geoPath } from "d3-geo";
import { select } from "d3-selection";
import { zoom, zoomIdentity, type ZoomBehavior } from "d3-zoom";
import "d3-transition";
import { feature } from "topojson-client";
import topology from "us-atlas/states-10m.json";
import countyTopology from "us-atlas/counties-10m.json";
import { getCities } from "./cities";
import { StatePanel } from "./StatePanel";
import {
  useNationalCoverage,
  useProspectPins,
  useStateOverview,
  useStateSocial,
  useStateChurches,
  useStateGovernment,
} from "../api/hooks";
import { mediaURL } from "../api/media";
import type { ProspectPin } from "../api/types";
import { type LayerId } from "./layers";

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
// all 56 states — dozens of times per second.
const STATE_PATHS: Record<string, string> = {};
const STATE_CENTROIDS: Record<string, [number, number]> = {};
for (const s of STATES) {
  const d = path(s.geometry);
  if (d) STATE_PATHS[s.id] = d;
  const c = path.centroid(s.geometry);
  if (c && !isNaN(c[0]) && !isNaN(c[1])) STATE_CENTROIDS[s.id] = c;
}

// County centroids keyed by 5-digit FIPS — used to place government-office
// pins. us-atlas counties-10m ships the same resolution as the state file,
// so no new dependency; same topojson feature() pattern as states above.
const countyFeature = feature(countyTopology as any, countyTopology.objects.counties as any) as any;
const COUNTY_CENTROIDS: Record<string, [number, number]> = {};
for (const f of countyFeature.features) {
  const c = path.centroid(f.geometry);
  if (c && !isNaN(c[0]) && !isNaN(c[1])) COUNTY_CENTROIDS[String(f.id)] = c;
}

// Denomination → pin color for diocese markers. Read from the diocese
// organization's metadata JSON (set by bootstrap-state); unknown denominations
// fall back to the Catholic gold so the pin is still visible.
const DENOM_COLOR: Record<string, string> = {
  catholic: "#D4AF37",
  episcopal: "#6B3FA0",
  methodist: "#C8102E",
  jewish: "#003F87",
};
function denomColor(metaJson?: string): string {
  if (!metaJson) return DENOM_COLOR.catholic;
  try {
    const d = (JSON.parse(metaJson) as { denomination?: string }).denomination ?? "catholic";
    return DENOM_COLOR[d] ?? DENOM_COLOR.catholic;
  } catch {
    return DENOM_COLOR.catholic;
  }
}

// --- National coverage choropleth -----------------------------------------
// 5-step red→green ramp keyed off a 0-1 coverage-health value. Discrete
// buckets match the spec exactly; computed once at module scope, never
// per-frame. Used for both the state fills and the legend gradient.
const COVERAGE_STOPS: Array<[number, string]> = [
  [0.0, "#451a03"], // no coverage
  [0.2, "#c2410c"], // minimal
  [0.4, "#a16207"], // partial
  [0.6, "#4d7c0f"], // good
  [0.8, "#15803d"], // full
];
function coverageColor(v: number): string {
  if (v < 0.2) return COVERAGE_STOPS[0][1];
  if (v < 0.4) return COVERAGE_STOPS[1][1];
  if (v < 0.6) return COVERAGE_STOPS[2][1];
  if (v < 0.8) return COVERAGE_STOPS[3][1];
  return COVERAGE_STOPS[4][1];
}

type CoverageMode = "radar" | "government" | "church" | "social";
const COVERAGE_MODES: Array<{ id: CoverageMode; label: string }> = [
  { id: "radar", label: "Radar Value" },
  { id: "government", label: "Government" },
  { id: "church", label: "Church" },
  { id: "social", label: "Social" },
];

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
  const [coverageMode, setCoverageMode] = useState<CoverageMode>("radar");
  // Layer + pin toggle are lifted here so the map can render the active
  // layer's pins; StatePanel receives them as props.
  const [layer, setLayer] = useState<LayerId>("overview");
  const [showPins, setShowPins] = useState(true);
  const svgRef = useRef<SVGSVGElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const zoomRef = useRef<ZoomBehavior<SVGSVGElement, unknown> | null>(null);
  // Custom HTML tooltip for registry pins — replaces native SVG <title> so the
  // tooltip is styled consistently and anchored to the pin, not the cursor.
  const [hoveredPin, setHoveredPin] = useState<{ x: number; y: number; content: ReactNode } | null>(null);

  // Source-registry drill-down panel is rendered by StatePanel, which
  // manages its own layer/county/diocese selection state internally.
  const selectedAbbrev = selected ? STATE_ABBREV[selected] : undefined;
  const hasSelection = !!selectedAbbrev;

  const nationalCoverage = useNationalCoverage(true);
  const prospectPins = useProspectPins(showProspects);

  // State-scoped registry data — react-query caches by state key, so these
  // are the same responses StatePanel fetches (no duplicate HTTP). Only
  // enabled while a state is selected so we don't load every state up front.
  const overview = useStateOverview(selectedAbbrev, hasSelection);
  const social = useStateSocial(selectedAbbrev, hasSelection);
  const churches = useStateChurches(selectedAbbrev, hasSelection);
  const government = useStateGovernment(selectedAbbrev, hasSelection);

  // City id → [lat, lng] from the overview payload. Vendors and dioceses
  // reference cities by id, so this lookup table is what places their pins.
  const cityCoords = useMemo(() => {
    const m = new Map<string, { lat: number; lng: number; name: string }>();
    for (const row of overview.data ?? []) {
      if (row.city.lat != null && row.city.lng != null) {
        m.set(row.city.id, { lat: row.city.lat, lng: row.city.lng, name: row.city.name });
      }
    }
    return m;
  }, [overview.data]);

  // --- Pin positions, precomputed per active layer -------------------------
  // Each memo projects lat/lng → SVG x/y once; the render just maps the
  // result. Recomputed only when the underlying dataset or zoom changes.
  const cityPins = useMemo(() => {
    if (!hasSelection || layer !== "overview") return [];
    return (overview.data ?? [])
      .filter((r) => r.city.lat != null && r.city.lng != null)
      .map((r) => {
        const pt = projection([r.city.lng!, r.city.lat!]);
        return pt ? { id: r.city.id, name: r.city.name, x: pt[0], y: pt[1] } : null;
      })
      .filter((p): p is { id: string; name: string; x: number; y: number } => !!p);
  }, [hasSelection, layer, overview.data]);

  // Vendors grouped by city — a city with >1 vendor renders a single cluster
  // badge with the count instead of overlapping dots.
  const vendorPins = useMemo(() => {
    if (!hasSelection || layer !== "social") return [];
    const byCity = new Map<string, { x: number; y: number; vendors: { name: string; handle?: string }[] }>();
    for (const market of social.data ?? []) {
      const city = market.city;
      if (city.lat == null || city.lng == null) continue;
      const pt = projection([city.lng, city.lat]);
      if (!pt) continue;
      const entry = byCity.get(city.id) ?? { x: pt[0], y: pt[1], vendors: [] };
      for (const v of market.vendors ?? []) {
        entry.vendors.push({
          name: v.organization.name,
          handle: v.watched_source?.handle,
        });
      }
      byCity.set(city.id, entry);
    }
    return [...byCity.entries()].map(([cityId, e]) => ({ cityId, ...e }));
  }, [hasSelection, layer, social.data]);

  const churchPins = useMemo(() => {
    if (!hasSelection || layer !== "churches") return [];
    return (churches.data ?? [])
      .map((d) => {
        const cityId = d.jurisdiction.hub_city_id ?? d.organization.city_id;
        const c = cityId ? cityCoords.get(cityId) : undefined;
        if (!c) return null;
        const pt = projection([c.lng, c.lat]);
        if (!pt) return null;
        return {
          id: d.jurisdiction.id,
          name: d.organization.name,
          parishCount: (d.parishes ?? []).length,
          color: denomColor(d.organization.metadata),
          x: pt[0],
          y: pt[1],
        };
      })
      .filter((p): p is NonNullable<typeof p> => !!p);
  }, [hasSelection, layer, churches.data, cityCoords]);

  const govPins = useMemo(() => {
    if (!hasSelection || layer !== "government") return [];
    return (government.data ?? [])
      .map((r) => {
        const c = COUNTY_CENTROIDS[r.county.id];
        if (!c) return null;
        const status = r.connector?.status ?? "not_configured";
        return {
          id: r.county.id,
          name: r.organization?.name ?? `${r.county.name} County`,
          county: r.county.name,
          status,
          x: c[0],
          y: c[1],
        };
      })
      .filter((p): p is NonNullable<typeof p> => !!p);
  }, [hasSelection, layer, government.data]);

  const coverageByAbbrev = useMemo(() => {
    const m = new Map<string, number>();
    for (const row of nationalCoverage.data ?? []) {
      m.set(row.state_id, row.alive_score);
    }
    return m;
  }, [nationalCoverage.data]);

  // Choropleth metric for the selected mode — normalized to 0-1 so the same
  // color scale applies to every toggle. Church/Social are raw counts, so
  // they're divided by the national max (computed once per dataset).
  const metricByAbbrev = useMemo(() => {
    const rows = nationalCoverage.data ?? [];
    const maxChurch = rows.reduce((m, r) => Math.max(m, r.church_sources), 1);
    const maxSocial = rows.reduce((m, r) => Math.max(m, r.social_sources), 1);
    const map = new Map<string, number>();
    for (const row of rows) {
      let v = 0;
      if (coverageMode === "radar") v = row.alive_score;
      else if (coverageMode === "government")
        v = row.county_count ? row.counties_configured / row.county_count : 0;
      else if (coverageMode === "church") v = row.church_sources / maxChurch;
      else v = row.social_sources / maxSocial;
      map.set(row.state_id, v);
    }
    return map;
  }, [nationalCoverage.data, coverageMode]);

  // National totals for the summary bar — aggregated from the same coverage
  // rows, recomputed only when the dataset changes.
  const summary = useMemo(() => {
    const rows = nationalCoverage.data ?? [];
    const withCoverage = rows.filter((r) => r.alive_score > 0).length;
    const gov = rows.reduce((s, r) => s + r.government_sources, 0);
    const church = rows.reduce((s, r) => s + r.church_sources, 0);
    const social = rows.reduce((s, r) => s + r.social_sources, 0);
    const avg = rows.length
      ? rows.reduce((s, r) => s + r.alive_score, 0) / rows.length
      : 0;
    return { withCoverage, total: rows.length, gov, church, social, avg };
  }, [nationalCoverage.data]);

  const resetZoom = useCallback((clearSelection = true) => {
    if (svgRef.current && zoomRef.current) {
      select(svgRef.current)
        .transition()
        .duration(650)
        .call(zoomRef.current.transform, zoomIdentity);
    }
    if (clearSelection) {
      setSelected(null);
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

  // Anchor the custom pin tooltip to the pin's screen position (top-center),
  // not the cursor — so it stays put while reading multi-line vendor lists.
  const pinEnter = useCallback((e: React.MouseEvent<SVGGElement>, content: ReactNode) => {
    const cr = containerRef.current?.getBoundingClientRect();
    const pr = e.currentTarget.getBoundingClientRect();
    if (cr && pr) setHoveredPin({ x: pr.left - cr.left + pr.width / 2, y: pr.top - cr.top, content });
  }, []);
  const pinLeave = useCallback(() => setHoveredPin(null), []);

  const hoveredAbbrev = hovered ? STATE_ABBREV[hovered] : null;
  const hoveredCities = hoveredAbbrev ? getCities(hoveredAbbrev) : [];
  const hoveredCoverage = hoveredAbbrev
    ? nationalCoverage.data?.find((c) => c.state_id === hoveredAbbrev)
    : undefined;

  // Counter-scale keeps pins a constant on-screen size as the map zooms.
  // Clamped so extreme zoom can't shrink pins to nothing or blow them up.
  const pinScale = Math.min(2, Math.max(0.5, 1 / zoomLevel));
  // At low zoom, close-together city labels collide — show every other one;
  // once zoomed in enough, show them all.
  const labelStride = zoomLevel < 2 ? 2 : 1;

  return (
    <div className="view view--map">
      <div className="map-frame" ref={containerRef} onMouseMove={handleMouseMove}>
        <div className="map-frame__topbar">
          <div className="map-frame__title-group">
            <h2 className="map-frame__title">Coverage map</h2>
            <p className="map-frame__subtitle">Where we're watching. 50 states + DC.</p>
            <span className="map-frame__meta">
              {selected
                ? STATE_MAP[selected]
                : `${summary.withCoverage || 0} states with sources · all 50 clickable`}
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

        <div className="map-coverage-bar">
          <div className="map-mode-toggle" role="group" aria-label="Coverage metric">
            {COVERAGE_MODES.map((m) => (
              <button
                key={m.id}
                className={`map-mode-btn ${coverageMode === m.id ? "map-mode-btn--active" : ""}`}
                onClick={() => setCoverageMode(m.id)}
                type="button"
              >
                {m.label}
              </button>
            ))}
          </div>
          <div className="map-summary">
            <span className="map-summary__stat">States <b>{summary.withCoverage}/{summary.total || 51}</b></span>
            <span className="map-summary__stat">Gov <b>{summary.gov}</b></span>
            <span className="map-summary__stat">Church <b>{summary.church}</b></span>
            <span className="map-summary__stat">Social <b>{summary.social}</b></span>
            <span className="map-summary__stat">Avg alive <b>{Math.round(summary.avg * 100)}%</b></span>
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
                  const alive = coverageByAbbrev.get(state.abbrev) ?? 0;
                  // Choropleth fill from the selected metric's 0-1 value.
                  // Selected states defer to the CSS selected style.
                  const metric = metricByAbbrev.get(state.abbrev) ?? 0;
                  const fill =
                    selected === state.id ? undefined : coverageColor(metric);
                  return (
                    <path
                      key={state.id}
                      d={d}
                      className={`map-state ${selected === state.id ? "map-state--selected" : ""} ${hovered === state.id ? "map-state--hovered" : ""} ${alive > 0 ? "map-state--alive" : ""}`}
                      style={fill ? { fill } : undefined}
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

              {/* --- State-scoped registry pins (only when zoomed in) ------ */}
              {hasSelection && showPins && (
                <g className="registry-pin-layer" pointerEvents="all">
                  {/* City markers — overview layer */}
                  {layer === "overview" && cityPins.map((c, i) => {
                    // Skip every other label at low zoom so close-together city
                    // names don't collide; pins still render so the dot is visible.
                    const showLabel = i % labelStride === 0;
                    return (
                      <g
                        key={`city-${c.id}`}
                        className="registry-pin registry-pin--city"
                        transform={`translate(${c.x},${c.y})`}
                        onMouseEnter={(e) => pinEnter(e, c.name)}
                        onMouseLeave={pinLeave}
                      >
                        <g className="registry-pin__inner">
                          <circle r={4 * pinScale} className="registry-pin__dot" />
                        </g>
                        {showLabel && (
                          <text
                            className="registry-pin__label"
                            x={8 * pinScale}
                            y={-6 * pinScale}
                            style={{ fontSize: `${10 * pinScale}px`, pointerEvents: "none" }}
                          >
                            {c.name}
                          </text>
                        )}
                      </g>
                    );
                  })}

                  {/* Vendor pins — social layer. Cluster badge when >1 per city. */}
                  {layer === "social" && vendorPins.map((v) => (
                    <g
                      key={`vendor-${v.cityId}`}
                      className="registry-pin registry-pin--vendor"
                      transform={`translate(${v.x},${v.y})`}
                      onMouseEnter={(e) => pinEnter(
                        e,
                        v.vendors.length > 1
                          ? v.vendors.map((vd) => `${vd.name}${vd.handle ? ` (@${vd.handle})` : ""}`).join("\n")
                          : `${v.vendors[0]?.name ?? "Vendor"}${v.vendors[0]?.handle ? ` (@${v.vendors[0].handle})` : ""}`,
                      )}
                      onMouseLeave={pinLeave}
                    >
                      <g className="registry-pin__inner">
                        {v.vendors.length > 1 ? (
                          <>
                            <circle r={10 * pinScale} className="registry-pin__cluster" />
                            <text
                              className="registry-pin__cluster-count"
                              textAnchor="middle"
                              dominantBaseline="central"
                              style={{ fontSize: `${9 * pinScale}px` }}
                            >
                              {v.vendors.length}
                            </text>
                          </>
                        ) : (
                          <circle r={3 * pinScale} className="registry-pin__dot" />
                        )}
                      </g>
                    </g>
                  ))}

                  {/* Diocese pins — churches layer. Cross symbol, denomination-colored. */}
                  {layer === "churches" && churchPins.map((d) => {
                    const s = 6 * pinScale;
                    return (
                      <g
                        key={`church-${d.id}`}
                        className="registry-pin registry-pin--church"
                        transform={`translate(${d.x},${d.y})`}
                        onMouseEnter={(e) => pinEnter(e, `${d.name} — ${d.parishCount} parishes`)}
                        onMouseLeave={pinLeave}
                      >
                        <g className="registry-pin__inner">
                          <path
                            d={`M0,${-s} L0,${s} M${-s},0 L${s},0`}
                            className="registry-pin__cross"
                            style={{ stroke: d.color }}
                          />
                        </g>
                      </g>
                    );
                  })}

                  {/* Government office pins — government layer. Square symbol. */}
                  {layer === "government" && govPins.map((g) => {
                    const s = 4 * pinScale;
                    return (
                      <g
                        key={`gov-${g.id}`}
                        className="registry-pin registry-pin--gov"
                        transform={`translate(${g.x},${g.y})`}
                        onMouseEnter={(e) => pinEnter(e, `${g.name} — ${g.status}`)}
                        onMouseLeave={pinLeave}
                      >
                        <g className="registry-pin__inner">
                          <rect x={-s} y={-s} width={s * 2} height={s * 2} className="registry-pin__square" />
                        </g>
                      </g>
                    );
                  })}
                </g>
              )}
            </g>
          </svg>

          {/* Cursor-following tooltip — position set imperatively via ref */}
          {hovered && !selected && tipVisible && (
            <div ref={tipRef} className="map-cursor-tip map-cursor-tip--coverage">
              <span className="map-cursor-tip__name">{STATE_MAP[hovered]}</span>
              {hoveredCoverage ? (
                <div className="map-cursor-tip__stats">
                  <span>Government: {hoveredCoverage.government_sources}</span>
                  <span>Church: {hoveredCoverage.church_sources}</span>
                  <span>Social: {hoveredCoverage.social_sources}</span>
                  <span>Counties: {hoveredCoverage.counties_configured}/{hoveredCoverage.county_count}</span>
                  <span>Alive: {Math.round(hoveredCoverage.alive_score * 100)}%</span>
                </div>
              ) : (
                <span className="map-cursor-tip__count">
                  {hoveredCities.length} planned markets
                </span>
              )}
            </div>
          )}

          {/* Custom pin tooltip — anchored to the hovered pin, not the cursor */}
          {hoveredPin && (
            <div className="map-tooltip" style={{ left: hoveredPin.x, top: hoveredPin.y }}>
              {hoveredPin.content}
            </div>
          )}

          {/* Coverage choropleth legend — gradient bar with bucket ticks */}
          <svg className="map-coverage-legend" width="230" height="46" viewBox="0 0 230 46" aria-hidden="true">
            <defs>
              <linearGradient id="coverage-grad" x1="0" x2="1" y1="0" y2="0">
                {COVERAGE_STOPS.map(([t, c], i) => (
                  <stop key={i} offset={`${t * 100}%`} stopColor={c} />
                ))}
                <stop offset="100%" stopColor={COVERAGE_STOPS[COVERAGE_STOPS.length - 1][1]} />
              </linearGradient>
            </defs>
            <rect x="10" y="8" width="210" height="12" rx="2" fill="url(#coverage-grad)" stroke="rgba(0,0,0,0.12)" strokeWidth="0.5" />
            {[0, 42, 84, 126, 168, 210].map((dx, i) => (
              <line key={i} x1={10 + dx} y1="20" x2={10 + dx} y2="24" stroke="rgba(0,0,0,0.35)" strokeWidth="1" />
            ))}
            <text x="10" y="40" className="map-coverage-legend__label" textAnchor="start">No coverage</text>
            <text x="220" y="40" className="map-coverage-legend__label" textAnchor="end">Full coverage</text>
          </svg>

          {/* Legend — explains every pin color for the active layer */}
          <div className="map-legend map-legend--stacked map-legend--corner">
            {showProspects && (
              <span className="map-legend__item">
                <i className="map-legend__swatch" style={{ background: "var(--mesa)" }} /> Prospects
              </span>
            )}
            {hasSelection && showPins && layer === "overview" && (
              <span className="map-legend__item">
                <i className="map-legend__swatch" style={{ background: "var(--ink-faint)" }} /> Cities
              </span>
            )}
            {hasSelection && showPins && layer === "social" && (
              <span className="map-legend__item">
                <i className="map-legend__swatch" style={{ background: "#E1306C" }} /> Instagram vendors
              </span>
            )}
            {hasSelection && showPins && layer === "churches" && (
              <>
                <span className="map-legend__item">
                  <i className="map-legend__swatch" style={{ background: DENOM_COLOR.catholic }} /> Catholic
                </span>
                <span className="map-legend__item">
                  <i className="map-legend__swatch" style={{ background: DENOM_COLOR.episcopal }} /> Episcopal
                </span>
                <span className="map-legend__item">
                  <i className="map-legend__swatch" style={{ background: DENOM_COLOR.methodist }} /> Methodist
                </span>
                <span className="map-legend__item">
                  <i className="map-legend__swatch" style={{ background: DENOM_COLOR.jewish }} /> Jewish
                </span>
              </>
            )}
            {hasSelection && showPins && layer === "government" && (
              <span className="map-legend__item">
                <i className="map-legend__swatch" style={{ background: "#1a3a5c" }} /> County offices
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

          {/* Source registry panel — works for any state */}
          {selected && selectedAbbrev && (
            <StatePanel
              stateAbbrev={selectedAbbrev}
              stateName={STATE_MAP[selected]}
              onClose={() => resetZoom()}
              layer={layer}
              setLayer={setLayer}
              showPins={showPins}
              setShowPins={setShowPins}
            />
          )}
        </div>
      </div>
    </div>
  );
}
