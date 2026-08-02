package llm

import (
	"context"
	"testing"
)

// TestEvalTemplate runs the golden eval cases against the template
// interpreter. The template is deterministic, so this is a stable
// regression check — if someone breaks the template, this fails.
func TestEvalTemplate(t *testing.T) {
	interp := NewTemplateInterpreter()
	results := RunEval(context.Background(), interp)
	passed := 0
	for _, r := range results {
		if r.Passed {
			passed++
		} else {
			t.Errorf("eval case %s FAILED: %s", r.CaseID, r.Diff)
		}
	}
	if passed == 0 {
		t.Error("no eval cases passed — template interpreter is broken")
	}
}
