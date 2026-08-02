// Package ratelimit provides per-identity API rate limiting. It's an
// in-memory token bucket: no Redis, no new dependency — sufficient for a
// small team where the main threat is a runaway script, not a DDoS.
//
// ponytail: ceiling — in-memory state is per-process. With leader election
// (Tier 1.1), only the leader polls, but all replicas serve the API. Each
// replica tracks its own buckets, so the effective limit is replicas × limit.
// Acceptable at 5 users; upgrade to Redis-backed if scale demands it.
package ratelimit

import (
	"net/http"
	"sync"
	"time"

	"neptune-social-radar/backend/internal/auth"
)

// bucket is one token bucket. Tokens refill at rate per second, capped at
// burst. Each request consumes 1 token.
type bucket struct {
	tokens   float64
	lastFill time.Time
	mu       sync.Mutex
}

// Limiter tracks per-identity buckets and enforces a global cap.
type Limiter struct {
	buckets sync.Map // key=identity (email or IP), value=*bucket

	rate  float64 // tokens per second per identity
	burst float64 // max tokens per identity

	// Global limiter: even with N users, total RPS can't exceed this.
	globalTokens   float64
	globalLastFill time.Time
	globalMu       sync.Mutex
	globalBurst    float64
}

// New creates a Limiter with the given per-identity rate (tokens/sec) and
// burst (max tokens). Global cap is burst × 4 (allows all users to burst
// simultaneously but limits sustained throughput).
func New(rate, burst float64) *Limiter {
	return &Limiter{
		rate:           rate,
		burst:          burst,
		globalTokens:   burst * 4,
		globalLastFill: time.Now(),
		globalBurst:    burst * 4,
	}
}

// Allow checks whether the identity (email or IP) may make a request.
// Returns true if allowed, false if rate-limited.
func (l *Limiter) Allow(identity string) bool {
	// Global check first — even unlimited identities can't exceed global cap.
	if !l.allowGlobal() {
		return false
	}

	b, _ := l.buckets.LoadOrStore(identity, &bucket{tokens: l.burst, lastFill: time.Now()})
	bkt := b.(*bucket)
	bkt.mu.Lock()
	defer bkt.mu.Unlock()
	l.refill(bkt)
	if bkt.tokens < 1 {
		return false
	}
	bkt.tokens -= 1
	return true
}

func (l *Limiter) allowGlobal() bool {
	l.globalMu.Lock()
	defer l.globalMu.Unlock()
	now := time.Now()
	elapsed := now.Sub(l.globalLastFill).Seconds()
	l.globalTokens += elapsed * l.rate * 4 // global refill = per-identity × 4
	if l.globalTokens > l.globalBurst {
		l.globalTokens = l.globalBurst
	}
	l.globalLastFill = now
	if l.globalTokens < 1 {
		return false
	}
	l.globalTokens -= 1
	return true
}

func (l *Limiter) refill(b *bucket) {
	now := time.Now()
	elapsed := now.Sub(b.lastFill).Seconds()
	b.tokens += elapsed * l.rate
	if b.tokens > l.burst {
		b.tokens = l.burst
	}
	b.lastFill = now
}

// Middleware wraps the next handler with rate limiting. The identity is
// the authenticated user's email, or the client IP in legacy mode. The
// limiter is nil-safe (no limit if not configured).
func Middleware(limiter *Limiter, next http.Handler) http.Handler {
	if limiter == nil {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Public routes skip rate limiting.
		if r.URL.Path == "/api/health" || r.URL.Path == "/api/media" {
			next.ServeHTTP(w, r)
			return
		}
		identity := auth.UserFromContext(r.Context()).Email
		if identity == "" {
			// Legacy mode or unauthenticated: fall back to client IP.
			identity = r.RemoteAddr
		}
		if !limiter.Allow(identity) {
			w.Header().Set("Retry-After", "5")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error":"rate limit exceeded"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}
