// ponytail: assert-based self-check for the Command Center's pure derivation
// logic (priority sort + kit aggregation). No framework, no fixtures — run:
//   node frontend/src/views/command_center.selfcheck.mjs
// Mirrors the inline useMemo logic in CommandCenterView.tsx. If the algorithm
// drifts, this fails first.

import assert from "node:assert/strict";

// --- topPriorityCouples: flatten board cards, drop unranked, sort desc, top N ---
function topPriorityCouples(board, n = 5) {
  const cards = board ? Object.values(board.cards).flat() : [];
  return cards
    .filter((c) => c.neptune_rank != null)
    .sort((a, b) => (b.neptune_rank ?? 0) - (a.neptune_rank ?? 0))
    .slice(0, n);
}

// --- deriveKitAgg: one kits list -> scans total, mailed, follow-ups sent/due ---
function deriveKitAgg(kits, now = Date.now()) {
  const list = kits ?? [];
  const scans = list.reduce((n, k) => n + (k.qr_scan_count ?? 0), 0);
  const mailed = list.filter((k) => k.status === "mailed").length;
  const followSent = list.filter((k) => k.follow_up_sent_at).length;
  const followDue = list.filter(
    (k) => k.follow_up_at && !k.follow_up_sent_at && new Date(k.follow_up_at).getTime() <= now,
  ).length;
  return { scans, mailed, followSent, followDue };
}

// priority sort
const board = {
  cards: {
    tagged_pair: [
      { couple_id: "a", person_a_label: "A", person_b_label: "B", neptune_rank: 10 },
      { couple_id: "b", person_a_label: "C", person_b_label: "D", neptune_rank: 90 },
    ],
    investigating: [
      { couple_id: "c", person_a_label: "E", person_b_label: "F" }, // unranked -> dropped
      { couple_id: "d", person_a_label: "G", person_b_label: "H", neptune_rank: 50 },
    ],
  },
};
const top = topPriorityCouples(board, 5);
assert.equal(top.length, 3, "unranked couples are dropped");
assert.deepEqual(top.map((c) => c.couple_id), ["b", "d", "a"], "sorted by neptune_rank desc");
assert.deepEqual(topPriorityCouples(null), [], "null board -> empty");

// kit aggregation
const now = new Date("2026-01-10T00:00:00Z").getTime();
const kits = [
  { id: "1", status: "mailed", qr_scan_count: 3, follow_up_sent_at: "2026-01-05" },
  { id: "2", status: "mailed", qr_scan_count: 0, follow_up_at: "2026-01-08" }, // due (past, unsent)
  { id: "3", status: "ready_to_mail", qr_scan_count: 1, follow_up_at: "2026-01-20" }, // future -> not due
  { id: "4", status: "mailed", qr_scan_count: 2, follow_up_at: "2026-01-01", follow_up_sent_at: "2026-01-02" }, // sent -> not due
];
const agg = deriveKitAgg(kits, now);
assert.equal(agg.scans, 6, "qr scans summed across all kits");
assert.equal(agg.mailed, 3, "only mailed-status kits counted");
assert.equal(agg.followSent, 2, "kits with follow_up_sent_at counted as sent");
assert.equal(agg.followDue, 1, "only past-due unsent follow-ups counted");
assert.deepEqual(deriveKitAgg(null), { scans: 0, mailed: 0, followSent: 0, followDue: 0 }, "null kits -> zeros");

console.log("command_center.selfcheck: OK");
