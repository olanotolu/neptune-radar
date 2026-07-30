import { feature } from "topojson-client";
// Ohio-only county topology, pre-trimmed from us-atlas/counties-10m.json
// (arcs remapped, 299 of 9,869 kept). The full file is 842 KB of
// non-Ohio counties we never render — it was the bulk of the map bundle.
// Regenerate if us-atlas is upgraded: see scripts note in ohio-counties.json's
// commit message (filter geometries to FIPS prefix 39, remap arcs).
import countiesTopology from "./ohio-counties.json";

// Ohio's real state FIPS code — matches the "39: OH" entry already in
// MapView.tsx's FIPS_TO_ABBREV table.
export const OHIO_FIPS = "39";

// us-atlas (already a dependency for the state map) also ships real county
// geometry at the same resolution — no new npm dependency needed. Same
// topojson-client.feature() pattern MapView.tsx already uses for states.
const countyFeatureCollection = feature(
  countiesTopology as any,
  (countiesTopology as any).objects.counties as any,
) as any;

export interface OhioCountyGeo {
  fips: string;
  name: string;
  geometry: any;
}

export const OHIO_COUNTIES: OhioCountyGeo[] = countyFeatureCollection.features
  .filter((f: any) => String(f.id).startsWith(OHIO_FIPS))
  .map((f: any) => ({ fips: String(f.id), name: f.properties.name as string, geometry: f.geometry }));

export type LayerId = "overview" | "government" | "churches" | "instagram";

export const LAYERS: { id: LayerId; label: string }[] = [
  { id: "overview", label: "Overview" },
  { id: "government", label: "Government" },
  { id: "churches", label: "Churches" },
  { id: "instagram", label: "Instagram" },
];

// Real, publicly known city coordinates — same trust tier as FIPS_TO_ABBREV.
// Used to place the diocese hub marker and the Instagram city-market marker;
// matches the coordinates cmd/bootstrap-ohio registered for Columbus.
export const OHIO_CITY_COORDS: Record<string, [number, number]> = {
  Columbus: [-82.9988, 39.9612],
};
