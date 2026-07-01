package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestMiddlewareRateLimit(t *testing.T) {
	mu.Lock()
	visitor = make(map[string]*bucket)
	mu.Unlock()

	handler := MiddlewareRateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	burstSize := 30
	var lastStatus int

	// Exhaust the bucket with burstSize requests
	for i := 0; i < burstSize; i++ {
		req := httptest.NewRequest(http.MethodGet, "/test-limit", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		lastStatus = rr.Code

		if rr.Code != http.StatusOK {
			t.Fatalf("request %d should succeed, got %d", i+1, rr.Code)
		}
	}

	// Next request should be rate limited
	req := httptest.NewRequest(http.MethodGet, "/test-limit", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	lastStatus = rr.Code

	if lastStatus != http.StatusTooManyRequests {
		t.Errorf("expected 429 after exhausting bucket, got %d", lastStatus)
	}

	// Different IP should not be rate limited
	req2 := httptest.NewRequest(http.MethodGet, "/test-limit", nil)
	req2.RemoteAddr = "127.0.0.2:12345"
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Errorf("different IP should not be limited, got %d", rr2.Code)
	}
}

func TestMiddlewareRateLimitConcurrent(t *testing.T) {
	mu.Lock()
	visitor = make(map[string]*bucket)
	mu.Unlock()

	handler := MiddlewareRateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	var wg sync.WaitGroup
	results := make([]int, 35)

	for i := 0; i < 35; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.RemoteAddr = "10.0.0.1:9999"
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			results[idx] = rr.Code
		}(i)
	}
	wg.Wait()

	successCount := 0
	limitedCount := 0
	for _, code := range results {
		if code == http.StatusOK {
			successCount++
		} else if code == http.StatusTooManyRequests {
			limitedCount++
		}
	}

	if successCount > 30 {
		t.Errorf("expected at most 30 successes, got %d", successCount)
	}
	if limitedCount < 5 {
		t.Errorf("expected at least 5 rate-limited requests, got %d", limitedCount)
	}
}
