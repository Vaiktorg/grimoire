package main

import (
	"errors"
	"net/http"
	"sync"
	"time"
)

type Guest struct {
	tries   int
	lastTry time.Time
	timeout time.Duration
	locked  bool
	locks   int
}

type RateLimiter struct {
	mu    sync.Mutex
	Tries int
	Time  time.Duration
	tries map[string]*Guest
}

const Timeout = time.Second
const LockedError = "locked"
const TooManyError = "too many requests, locked"
const EmptyIpError = "empty ip"

func NewRateLimiter(tries int, t time.Duration) *RateLimiter {
	return &RateLimiter{
		tries: make(map[string]*Guest),
		Tries: tries,
		Time:  t,
	}
}

func (rl *RateLimiter) Allow(ip string) error {
	if ip == "" {
		return errors.New(EmptyIpError)
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// New Guest
	g, ok := rl.tries[ip]
	if !ok {
		rl.tries[ip] = &Guest{
			timeout: Timeout,
			lastTry: time.Now(),
		}

		return nil
	}

	if g.locked {
		return errors.New(LockedError)
	}

	// Reset the tries if the time has passed
	now := time.Now()
	if now.Sub(g.lastTry) >= rl.Time {
		g.tries++
		g.lastTry = now

		rl.tries[ip] = g
		return nil
	}

	// Check if the guest has exceeded the limit
	if g.tries >= rl.Tries {
		rl.lockGuest(ip, g)
		return errors.New(TooManyError)
	}

	g.tries++
	rl.tries[ip] = g

	return nil
}

func (rl *RateLimiter) lockGuest(ip string, g *Guest) {
	g.locked = true
	g.locks++

	time.AfterFunc(g.timeout, func() {
		g.locked = false
		g.tries = 0
		g.timeout *= 2
	})

	rl.tries[ip] = g
}

func (rl *RateLimiter) RateLimitMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr

		err := rl.Allow(ip)
		if err != nil {
			switch err.Error() {
			case TooManyError:
				http.Error(w, TooManyError, http.StatusTooManyRequests)
			case LockedError:
				http.Error(w, LockedError, http.StatusLocked)
			case EmptyIpError:
				http.Error(w, EmptyIpError, http.StatusBadRequest)
			}
			return
		}

		h.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) RateLimitChecker(r *http.Request) bool {
	ip := r.RemoteAddr
	err := rl.Allow(ip)
	if err != nil {
		// Log the error if needed
		return false
	}
	return true
}
