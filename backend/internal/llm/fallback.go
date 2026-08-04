package llm

import (
	"context"
	"sync"
	"time"
)

// CircuitBreaker skips a provider that's been failing consistently, so the
// fallback doesn't waste 60s of timeout on every call when a provider is down.
// After maxFailures consecutive errors, the provider is tripped for cooldown;
// after cooldown, one trial call is allowed (half-open). A success resets.
// ponytail: ceiling — in-memory, per-process. No shared state across replicas.
// Fine for a single-process worker; upgrade to Redis-backed if multiple
// replicas need to share circuit state.
type CircuitBreaker struct {
	mu          sync.Mutex
	failures    int
	maxFailures int
	cooldown    time.Duration
	trippedAt   time.Time
}

func NewCircuitBreaker(maxFailures int, cooldown time.Duration) *CircuitBreaker {
	return &CircuitBreaker{maxFailures: maxFailures, cooldown: cooldown}
}

// Allow returns true if the provider should be tried. If the circuit is
// tripped and cooldown hasn't elapsed, returns false. If cooldown has
// elapsed, allows one trial call (half-open state).
func (c *CircuitBreaker) Allow() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.failures < c.maxFailures {
		return true
	}
	if time.Since(c.trippedAt) > c.cooldown {
		return true // half-open: allow one trial
	}
	return false
}

func (c *CircuitBreaker) RecordSuccess() {
	c.mu.Lock()
	c.failures = 0
	c.mu.Unlock()
}

func (c *CircuitBreaker) RecordFailure() {
	c.mu.Lock()
	c.failures++
	if c.failures == c.maxFailures {
		c.trippedAt = time.Now()
	}
	c.mu.Unlock()
}

// FallbackInterpreter tries OpenAI first, then Baseten, then Claude, then the
// deterministic template — so the system produces output even when a particular
// model is unavailable. Circuit breakers prevent wasting time on a dead provider.
type FallbackInterpreter struct {
	openai    *OpenAIInterpreter
	baseten   *BasetenInterpreter
	claude    *ClaudeInterpreter
	fallback  *TemplateInterpreter
	openaiCB  *CircuitBreaker
	basetenCB *CircuitBreaker
	claudeCB  *CircuitBreaker
}

func NewInterpreter() Interpreter {
	openai := NewOpenAIInterpreter()
	baseten := NewBasetenInterpreter()
	claude := NewClaudeInterpreter()
	tmpl := NewTemplateInterpreter()

	f := &FallbackInterpreter{fallback: tmpl}
	if openai.Available() {
		f.openai = openai
		f.openaiCB = NewCircuitBreaker(5, 60*time.Second)
	}
	if baseten.Available() {
		f.baseten = baseten
		f.basetenCB = NewCircuitBreaker(5, 60*time.Second)
	}
	if claude.Available() {
		f.claude = claude
		f.claudeCB = NewCircuitBreaker(5, 60*time.Second)
	}
	if f.openai == nil && f.baseten == nil && f.claude == nil {
		return tmpl
	}
	return f
}

func (f *FallbackInterpreter) HasOpenAI() bool  { return f.openai != nil }
func (f *FallbackInterpreter) HasBaseten() bool { return f.baseten != nil }
func (f *FallbackInterpreter) HasClaude() bool  { return f.claude != nil }

func (f *FallbackInterpreter) InterpretSignal(ctx context.Context, req SignalRequest) (Interpretation, error) {
	if f.openai != nil && f.openaiCB.Allow() {
		if out, err := f.openai.InterpretSignal(ctx, req); err == nil {
			f.openaiCB.RecordSuccess()
			return out, nil
		} else {
			f.openaiCB.RecordFailure()
		}
	}
	if f.baseten != nil && f.basetenCB.Allow() {
		if out, err := f.baseten.InterpretSignal(ctx, req); err == nil {
			f.basetenCB.RecordSuccess()
			return out, nil
		} else {
			f.basetenCB.RecordFailure()
		}
	}
	if f.claude != nil && f.claudeCB.Allow() {
		if out, err := f.claude.InterpretSignal(ctx, req); err == nil {
			f.claudeCB.RecordSuccess()
			return out, nil
		} else {
			f.claudeCB.RecordFailure()
		}
	}
	return f.fallback.InterpretSignal(ctx, req)
}

func (f *FallbackInterpreter) DraftCopy(ctx context.Context, req CopyRequest) (Copy, error) {
	if f.openai != nil && f.openaiCB.Allow() {
		if out, err := f.openai.DraftCopy(ctx, req); err == nil {
			f.openaiCB.RecordSuccess()
			return out, nil
		} else {
			f.openaiCB.RecordFailure()
		}
	}
	if f.baseten != nil && f.basetenCB.Allow() {
		if out, err := f.baseten.DraftCopy(ctx, req); err == nil {
			f.basetenCB.RecordSuccess()
			return out, nil
		} else {
			f.basetenCB.RecordFailure()
		}
	}
	if f.claude != nil && f.claudeCB.Allow() {
		if out, err := f.claude.DraftCopy(ctx, req); err == nil {
			f.claudeCB.RecordSuccess()
			return out, nil
		} else {
			f.claudeCB.RecordFailure()
		}
	}
	return f.fallback.DraftCopy(ctx, req)
}
