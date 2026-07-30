package llm

import "context"

// FallbackInterpreter tries Baseten first, then Claude, then falls back to the
// deterministic template on any error — so the system produces output even
// when a particular model is unavailable.
type FallbackInterpreter struct {
	baseten  *BasetenInterpreter
	claude   *ClaudeInterpreter
	fallback *TemplateInterpreter
}

func NewInterpreter() Interpreter {
	baseten := NewBasetenInterpreter()
	claude := NewClaudeInterpreter()
	tmpl := NewTemplateInterpreter()

	if baseten.Available() {
		if claude.Available() {
			return &FallbackInterpreter{baseten: baseten, claude: claude, fallback: tmpl}
		}
		return &FallbackInterpreter{baseten: baseten, fallback: tmpl}
	}
	if claude.Available() {
		return &FallbackInterpreter{claude: claude, fallback: tmpl}
	}
	return tmpl
}

func (f *FallbackInterpreter) HasBaseten() bool { return f.baseten != nil }
func (f *FallbackInterpreter) HasClaude() bool  { return f.claude != nil }

func (f *FallbackInterpreter) InterpretSignal(ctx context.Context, req SignalRequest) (Interpretation, error) {
	if f.baseten != nil {
		if out, err := f.baseten.InterpretSignal(ctx, req); err == nil {
			return out, nil
		}
	}
	if f.claude != nil {
		if out, err := f.claude.InterpretSignal(ctx, req); err == nil {
			return out, nil
		}
	}
	return f.fallback.InterpretSignal(ctx, req)
}

func (f *FallbackInterpreter) DraftCopy(ctx context.Context, req CopyRequest) (Copy, error) {
	if f.baseten != nil {
		if out, err := f.baseten.DraftCopy(ctx, req); err == nil {
			return out, nil
		}
	}
	if f.claude != nil {
		if out, err := f.claude.DraftCopy(ctx, req); err == nil {
			return out, nil
		}
	}
	return f.fallback.DraftCopy(ctx, req)
}
