package main

import (
	"net/http"
)

type CheckFunc func(r *http.Request) bool

type SusDetect struct {
	checks []CheckFunc
}

func NewSusDetectMiddleware(checks ...CheckFunc) *SusDetect {
	return &SusDetect{
		checks: checks,
	}
}

func (m *SusDetect) DetectSuspiciousActivity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, check := range m.checks {
			if !check(r) {
				http.Error(w, "Suspicious activity detected", http.StatusForbidden)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
