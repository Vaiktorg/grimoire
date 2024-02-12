package tests

import (
	"github.com/vaiktorg/grimoire/authentity/internal"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

var attempts = 5
var RateLimiter = internal.NewMockedRateLimiter(attempts, timeout, &internal.MockIPGetter{})

var timeout = time.Second

func TestRateLimitMiddleware_AllowsAccessWithFewAttempts(t *testing.T) {
	handler := RateLimiter.RateLimitMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server := httptest.NewServer(handler)
	defer server.Close()

	for i := 0; i < attempts; i++ {
		req, err := http.NewRequest("GET", server.URL, nil)
		if err != nil {
			t.Fatal(err)
			return
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
			return
		}

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		_ = resp.Body.Close()
	}
}

func TestRateLimitMiddleware_BlocksAccessWithTooManyAttempts(t *testing.T) {
	handler := RateLimiter.RateLimitMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server := httptest.NewServer(handler)
	defer server.Close()

	for i := 0; i <= attempts; i++ {
		req, err := http.NewRequest("GET", server.URL, nil)
		if err != nil {
			t.Fatal(err)
			return
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
			return
		}

		time.Sleep(time.Second)

		if i >= attempts && resp.StatusCode != http.StatusTooManyRequests {
			t.Errorf("Expected status 429, got %d", resp.StatusCode)
		}

		_ = resp.Body.Close()
	}
}

func TestRateLimitMiddleware_UnblocksAfterTimeout(t *testing.T) {
	handler := RateLimiter.RateLimitMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server := httptest.NewServer(handler)
	defer server.Close()

	for i := 0; i <= attempts; i++ {
		req, err := http.NewRequest("GET", server.URL, nil)
		if err != nil {
			t.Fatal(err)
			return
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
			return
		}
		_ = resp.Body.Close()
	}

	time.Sleep(timeout * 2)

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatal(err)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestRateLimitMiddleware_IncreasesTimeoutAfterLockout(t *testing.T) {
	handler := RateLimiter.RateLimitMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server := httptest.NewServer(handler)
	defer server.Close()

	res := &http.Response{}
	for j := 0; j < 2; j++ {
		// Try 1
		for i := 0; i <= attempts; i++ {
			req, err := http.NewRequest("GET", server.URL, nil)
			if err != nil {
				t.Fatal(err)
				return
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
				return
			}

			res = resp

			_ = resp.Body.Close()
		}

		timeout += timeout
		// locked 1
		time.Sleep(timeout)
	}

	if res.StatusCode != http.StatusLocked {
		t.Errorf("Expected status 423, got %d", res.StatusCode)
	}

}
