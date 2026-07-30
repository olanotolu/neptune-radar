import type { RelationshipStage } from "../api/types";

const LABELS: Record<RelationshipStage, string> = {
  unknown: "Unknown",
  dating_suspected: "Dating (suspected)",
  engaged: "Engaged",
  married: "Married",
  status_uncertain: "Status uncertain",
  ended_suspected: "Ended (suspected)",
};

export function StageBadge({ stage }: { stage: RelationshipStage }) {
  return <span className={`stage-badge stage-${stage}`}>{LABELS[stage] ?? stage}</span>;
}
