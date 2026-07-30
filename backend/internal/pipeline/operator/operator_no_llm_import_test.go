package operator_test

import (
	"os/exec"
	"strings"
	"testing"
)

// TestNoLLMImport mirrors the same invariant check in pipeline/policy: the
// Workflow Operator only ever writes plain-string copy handed to it by the
// orchestrator — it must never be able to reach into internal/llm itself.
func TestNoLLMImport(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps", "neptune-social-radar/backend/internal/pipeline/operator").Output()
	if err != nil {
		t.Fatalf("go list -deps: %v", err)
	}
	for _, line := range strings.Split(string(out), "\n") {
		if line == "neptune-social-radar/backend/internal/llm" {
			t.Fatal("internal/pipeline/operator must never depend on internal/llm")
		}
	}
}
