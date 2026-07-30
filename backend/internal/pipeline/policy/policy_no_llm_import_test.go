package policy_test

import (
	"os/exec"
	"strings"
	"testing"
)

// TestNoLLMImport enforces the architectural invariant stated at the top of
// guard.go: the Policy Guard must never import internal/llm, directly or
// transitively. This is what makes "the model proposes, policy decides" a
// checkable property instead of a comment.
func TestNoLLMImport(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps", "neptune-social-radar/backend/internal/pipeline/policy").Output()
	if err != nil {
		t.Fatalf("go list -deps: %v", err)
	}
	for _, line := range strings.Split(string(out), "\n") {
		if line == "neptune-social-radar/backend/internal/llm" {
			t.Fatal("internal/pipeline/policy must never depend on internal/llm — the model must not be able to influence what the system is permitted to do")
		}
	}
}
