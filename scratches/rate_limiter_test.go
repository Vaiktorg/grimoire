package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestTryLimiter(t *testing.T) {
	rl := NewRateLimiter(100, Timeout)

	// Test Allow method
	t.Run("Allow", func(t *testing.T) {
		// Test with empty IP
		err := rl.Allow("")
		if err == nil || err.Error() != EmptyIpError {
			t.Errorf("Expected error for empty IP, got %v", err)
		}

		// Test with new IP
		err = rl.Allow("192.0.2.1")
		if err != nil {
			t.Errorf("Expected no error for new IP, got %v", err)
		}

		// Test with IP that has not exceeded limit
		for i := 0; i < rl.Tries; i++ {
			err = rl.Allow("192.0.2.1")
			if err != nil {
				t.Errorf("Expected no error for IP under limit, got %v", err)
			}
		}

		// Test with IP that has exceeded limit but is not locked
		err = rl.Allow("192.0.2.1")
		if err == nil || err.Error() != TooManyError {
			t.Errorf("Expected error for IP over limit, got %v", err)
		}

		// Test with locked IP
		time.Sleep(Timeout)
		err = rl.Allow("192.0.2.1")
		if err == nil || err.Error() != LockedError {
			t.Errorf("Expected error for locked IP, got %v", err)
		}
	})

	// Test lockGuest method
	t.Run("lockGuest", func(t *testing.T) {
		g := &Guest{
			timeout: Timeout,
			locked:  false,
		}

		rl.lockGuest("192.0.2.1", g)
		if !g.locked {
			t.Errorf("Expected guest to be locked")
		}

		time.Sleep(2 * Timeout)
		if g.locked {
			t.Errorf("Expected guest to be unlocked after timeout")
		}
	})

	// Test RateLimitMiddleware method
	t.Run("RateLimitMiddleware", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/", nil)
		req.RemoteAddr = "192.0.2.1"

		rr := httptest.NewRecorder()
		handler := rl.RateLimitMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

		for i := 0; i < rl.Tries; i++ {
			_ = rl.Allow("192.0.2.1")
		}

		handler.ServeHTTP(rr, req)
		if status := rr.Code; status != http.StatusTooManyRequests {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusTooManyRequests)
		}
	})
}
