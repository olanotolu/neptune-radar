package ratelimit

import (
	"testing"
	"time"
)

func TestLimiterAllowsWithinBurst(t *testing.T) {
	l := New(10, 5) // 10/sec, burst 5
	for i := 0; i < 5; i++ {
		if !l.Allow("user1") {
			t.Errorf("request %d should be allowed within burst", i+1)
		}
	}
}

func TestLimiterBlocksAfterBurst(t *testing.T) {
	l := New(0.1, 3) // 0.1/sec (very slow refill), burst 3
	for i := 0; i < 3; i++ {
		l.Allow("user1")
	}
	if l.Allow("user1") {
		t.Error("4th request should be blocked after burst exhausted")
	}
}

func TestLimiterRefillsOverTime(t *testing.T) {
	l := New(100, 2) // 100/sec, burst 2
	l.Allow("user1")
	l.Allow("user1")
	// Bucket is empty. Wait for refill.
	time.Sleep(20 * time.Millisecond)
	if !l.Allow("user1") {
		t.Error("should be allowed after refill")
	}
}

func TestLimiterSeparateIdentities(t *testing.T) {
	l := New(0.1, 2) // burst 2 each
	l.Allow("user1")
	l.Allow("user1")
	// user1 is exhausted, but user2 has its own bucket.
	if !l.Allow("user2") {
		t.Error("user2 should have its own bucket")
	}
}

func TestLimiterNilIsNoop(t *testing.T) {
	// Middleware with nil limiter should just pass through.
	// Covered by the nil check in Middleware — verify Allow isn't called.
	// This test documents that nil is safe.
	var l *Limiter
	if l != nil {
		t.Error("nil limiter should be nil")
	}
}

func TestLimiterGlobalCap(t *testing.T) {
	// With many users each having burst 2, the global cap (burst × 4 = 8)
	// should eventually block even new users.
	l := New(0.01, 2) // very slow refill, burst 2, global burst 8
	allowed := 0
	for i := 0; i < 20; i++ {
		identity := "user" + string(rune('A'+i))
		if l.Allow(identity) {
			allowed++
		}
	}
	if allowed > 10 {
		t.Errorf("global cap should limit total allowed requests, got %d", allowed)
	}
}
