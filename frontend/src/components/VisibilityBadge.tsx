const LABELS: Record<string, string> = {
  private_person_a: "Private (A)",
  private_person_b: "Private (B)",
  shared_couple: "Shared (couple)",
  neptune_internal: "Neptune internal",
  attorney_only: "Attorney only",
  unconfirmed_inference: "Unconfirmed inference",
};

export function VisibilityBadge({ scope }: { scope: string }) {
  return <span className={`visibility-badge visibility-${scope}`}>{LABELS[scope] ?? scope}</span>;
}
