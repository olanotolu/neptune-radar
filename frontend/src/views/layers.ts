export type LayerId = "overview" | "government" | "churches" | "social";

export const LAYERS: { id: LayerId; label: string }[] = [
  { id: "overview", label: "Overview" },
  { id: "government", label: "Government" },
  { id: "churches", label: "Churches" },
  { id: "social", label: "Social" },
];
