package internal

import (
	"net/http"
	"sync"
	"time"
)

type IPGetter interface {
	GetIP() string
}

type MockIPGetter struct{}

func (m *MockIPGetter) GetIP() string {
	return "192.168.1.1" // or any other IP address
}

// =================== RateLimiter ===================

type Guest struct {
	lastSeen time.Time     // Last time the handler was accessed
	attempts int           // How many attempts at accessing the handler
	timeout  time.Duration // timeout extension
	locked   bool          // Whether the account is locked
	locks    int           // How many times it's been locked
}

type RateLimiter struct {
	mu       sync.Mutex
	visitors map[string]*Guest
	Attempts int
	Timeout  time.Duration
	mockIP   IPGetter
}

func NewMockedRateLimiter(attempts int, timeout time.Duration, getter *MockIPGetter) *RateLimiter {
	return &RateLimiter{
		visitors: make(map[string]*Guest),
		Attempts: attempts,
		Timeout:  timeout,
		mockIP:   getter,
	}
}
func NewRateLimiter(attempts int, timeout time.Duration) *RateLimiter {
	return &RateLimiter{
		visitors: make(map[string]*Guest),
		Attempts: attempts,
		Timeout:  timeout,
	}
}

func (rl *RateLimiter) AddVisitor(ip string) *Guest {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		rl.visitors[ip] = &Guest{lastSeen: time.Now(), timeout: rl.Timeout}
		return rl.visitors[ip]
	}

	if v.locked {
		return v
	}

	v.lastSeen = time.Now()
	v.attempts++

	rl.visitors[ip] = v // Update the Allow pointer in the visitors map

	return v
}

func (rl *RateLimiter) RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		visitor := rl.AddVisitor(rl.mockIP.GetIP())

		if visitor.locked {
			err := "Account locked, please wait " + time.Now().Sub(visitor.lastSeen).String()
			http.Error(w, err, http.StatusLocked)
			return
		}

		if visitor.attempts >= rl.Attempts {
			rl.lockAccount(rl.mockIP.GetIP())

			err := "too many requests. Account locked, please wait " + visitor.timeout.String()
			println(err)

			http.Error(w, err, http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (rl *RateLimiter) lockAccount(ip string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if visitor, exists := rl.visitors[ip]; exists {
		visitor.locked = true
		visitor.locks++

		time.AfterFunc(visitor.timeout, func() {
			visitor.locked = false
			visitor.attempts = 0
			visitor.timeout += visitor.timeout

			rl.visitors[rl.mockIP.GetIP()] = visitor // Update the visitor in the map
		})

		rl.visitors[ip] = visitor
	}

}
