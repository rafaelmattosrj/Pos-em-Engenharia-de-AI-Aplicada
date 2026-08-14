package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimiter_AllowsUpToCapacity(t *testing.T) {
	rl := NewRateLimiter()
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < rateLimitCapacity; i++ {
		req := httptest.NewRequest(http.MethodGet, "/customers", nil)
		req.Header.Set("Authorization", "Bearer token-fixo")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("requisicao %d: esperava 200, obteve %d", i+1, rec.Code)
		}
	}

	// A requisicao 91 deve estourar o limite.
	req := httptest.NewRequest(http.MethodGet, "/customers", nil)
	req.Header.Set("Authorization", "Bearer token-fixo")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("esperava 429 apos estourar o limite, obteve %d", rec.Code)
	}
}

func TestRateLimiter_UnauthenticatedRequestsBypass(t *testing.T) {
	rl := NewRateLimiter()
	calls := 0
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusOK)
	}))

	for i := 0; i < rateLimitCapacity+10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("requisicao sem token nao deveria ser limitada, obteve %d na iteracao %d", rec.Code, i)
		}
	}
	if calls != rateLimitCapacity+10 {
		t.Errorf("esperava %d chamadas, obteve %d", rateLimitCapacity+10, calls)
	}
}

func TestRateLimiter_SeparateBucketsPerToken(t *testing.T) {
	rl := NewRateLimiter()
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Esgota o limite do token A.
	for i := 0; i < rateLimitCapacity; i++ {
		req := httptest.NewRequest(http.MethodGet, "/customers", nil)
		req.Header.Set("Authorization", "Bearer token-a")
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}

	// O token B deve ter seu proprio bucket, ainda cheio.
	req := httptest.NewRequest(http.MethodGet, "/customers", nil)
	req.Header.Set("Authorization", "Bearer token-b")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("token diferente nao deveria ser afetado pelo limite do token-a, obteve %d", rec.Code)
	}
}
