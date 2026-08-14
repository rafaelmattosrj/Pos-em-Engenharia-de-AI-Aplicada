// Package middleware implementa os filtros HTTP da API legada — equivalente
// a JwtFilter.java e RateLimitFilter.java.
package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	rateLimitCapacity = 90
	rateLimitPeriod   = time.Minute
)

// tokenBucket implementa um bucket de token com refill "greedy" (contínuo,
// proporcional ao tempo decorrido) — equivalente ao Bandwidth.refillGreedy
// do Bucket4j usado em RateLimitFilter.java.
type tokenBucket struct {
	mu         sync.Mutex
	tokens     float64
	capacity   float64
	refillRate float64 // tokens por segundo
	lastRefill time.Time
}

func newTokenBucket(capacity int, period time.Duration) *tokenBucket {
	return &tokenBucket{
		tokens:     float64(capacity),
		capacity:   float64(capacity),
		refillRate: float64(capacity) / period.Seconds(),
		lastRefill: time.Now(),
	}
}

func (b *tokenBucket) tryConsume() bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(b.lastRefill).Seconds()
	b.tokens += elapsed * b.refillRate
	if b.tokens > b.capacity {
		b.tokens = b.capacity
	}
	b.lastRefill = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// RateLimiter aplica um limite de 90 requisições/minuto por token —
// equivalente a RateLimitFilter.java.
type RateLimiter struct {
	mu      sync.Mutex
	buckets map[string]*tokenBucket
}

// NewRateLimiter cria um RateLimiter vazio.
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{buckets: make(map[string]*tokenBucket)}
}

// Middleware aplica rate limiting apenas a requisições autenticadas (com
// header Authorization: Bearer <token>) — um bucket por token.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		bucket := rl.bucketFor(token)
		if !bucket.tryConsume() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Too many requests. Limit: 90 req/min per token.",
			})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) bucketFor(token string) *tokenBucket {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, ok := rl.buckets[token]
	if !ok {
		bucket = newTokenBucket(rateLimitCapacity, rateLimitPeriod)
		rl.buckets[token] = bucket
	}
	return bucket
}

func bearerToken(r *http.Request) (string, bool) {
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, "Bearer ") {
		return "", false
	}
	return strings.TrimPrefix(header, "Bearer "), true
}
