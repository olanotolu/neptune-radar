package vision

import (
	"math"
	"testing"
)

func TestDispersionScore(t *testing.T) {
	// Graph: two clusters bridged by the couple's mutual friends.
	// Cluster A: alice, bob, carol (tightly connected)
	// Cluster B: dave, eve, frank (tightly connected)
	// Mutual follows = [alice, dave] — one from each cluster, far apart.
	graph := map[string][]string{
		"alice": {"bob", "carol"},
		"bob":   {"alice", "carol"},
		"carol": {"alice", "bob"},
		"dave":  {"eve", "frank"},
		"eve":   {"dave", "frank"},
		"frank": {"dave", "eve"},
	}

	// Two mutual follows from different clusters → no path between them →
	// pairs=0 → score 0 (undefined dispersion, not high dispersion).
	score := DispersionScore([]string{"alice", "dave"}, graph)
	if score != 0 {
		t.Errorf("disconnected mutual follows: expected 0, got %.4f", score)
	}

	// Add a bridge between clusters so distance(alice, dave) = 3.
	graph["carol"] = append(graph["carol"], "frank")
	graph["frank"] = append(graph["frank"], "carol")
	score = DispersionScore([]string{"alice", "dave"}, graph)
	// distance(alice→carol→frank→dave) = 3, avgSqDist = 9, score = 9/11 ≈ 0.818
	if math.Abs(score-9.0/11.0) > 0.001 {
		t.Errorf("bridged clusters: expected %.4f, got %.4f", 9.0/11.0, score)
	}
	if score <= 0.7 {
		t.Errorf("high-dispersion case should exceed 0.7, got %.4f", score)
	}

	// Tightly connected mutual follows → low dispersion.
	// alice, bob, carol are all directly connected: distance(alice,bob)=1,
	// distance(alice,carol)=1, distance(bob,carol)=1. avgSqDist=1, score=1/3.
	score = DispersionScore([]string{"alice", "bob", "carol"}, graph)
	if math.Abs(score-1.0/3.0) > 0.001 {
		t.Errorf("tight cluster: expected %.4f, got %.4f", 1.0/3.0, score)
	}
	if score > 0.7 {
		t.Errorf("tight cluster should be low dispersion (<0.7), got %.4f", score)
	}

	// Edge cases.
	if DispersionScore(nil, graph) != 0 {
		t.Error("nil mutual follows should return 0")
	}
	if DispersionScore([]string{"only_one"}, graph) != 0 {
		t.Error("single mutual follow should return 0")
	}
}
