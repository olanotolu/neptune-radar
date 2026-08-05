package vision

// DispersionScore implements the Facebook AI Research dispersion metric
// (Backstrom & Kleinberg, 2014) for relationship scoring.
//
// Dispersion measures how socially "spread out" the mutual friends of two
// people are. If the mutual friends are tightly connected to each other (low
// distance), they likely come from one social circle (work, school, family).
// If they are NOT connected to each other (high distance), the relationship
// bridges distinct social circles — a strong romantic-relationship signal.
//
// High dispersion (>0.7) = strong romantic relationship signal.
//
// mutualFollows is the set of accounts both people follow (common neighbors).
// graph is the adjacency list of the social graph (handle → handles it follows).
//
// ponytail: O(n²·(V+E)) — one BFS per pair of mutual follows. Fine for the
// small n (typically <100 mutual follows) and sparse subgraphs we operate on.
// Upgrade path: cache BFS results across pairs, or approximate with a
// landmark-based distance estimator for very large n.
func DispersionScore(mutualFollows []string, graph map[string][]string) float64 {
	n := len(mutualFollows)
	if n < 2 {
		return 0 // can't compute dispersion with < 2 mutual friends
	}
	var sumSqDist float64
	var pairs int
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			d := bfsDistance(mutualFollows[i], mutualFollows[j], graph)
			if d > 0 { // 0 = same node, -1 = no path (skip unreachable pairs)
				sumSqDist += float64(d) * float64(d)
				pairs++
			}
		}
	}
	if pairs == 0 {
		return 0 // all mutual friends are disconnected from each other → undefined
	}
	avgSqDist := sumSqDist / float64(pairs)
	// Squash to 0–1: score = avgSqDist / (avgSqDist + 2). At avgSqDist=2
	// (mutual friends ~1.4 hops apart on average) score=0.5; the >0.7
	// threshold corresponds to avgSqDist ≈ 4.67 (mutual friends ~2.2 hops
	// apart — clearly bridging separate clusters).
	return avgSqDist / (avgSqDist + 2.0)
}

// bfsDistance returns the shortest-path distance between src and dst in the
// graph, or -1 if unreachable. A direct edge counts as distance 1.
func bfsDistance(src, dst string, graph map[string][]string) int {
	if src == dst {
		return 0
	}
	visited := map[string]bool{src: true}
	queue := []string{src}
	dist := 0
	for len(queue) > 0 {
		dist++
		for size := len(queue); size > 0; size-- {
			node := queue[0]
			queue = queue[1:]
			for _, neighbor := range graph[node] {
				if neighbor == dst {
					return dist
				}
				if !visited[neighbor] {
					visited[neighbor] = true
					queue = append(queue, neighbor)
				}
			}
		}
	}
	return -1
}
