package internal

import (
	"net/http"
)

func SecurityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set Content Security Policy
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; object-src 'none';")

		// Set X-Content-Type-Options
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Set X-Frame-Options
		w.Header().Set("X-Frame-Options", "DENY")

		// Set X-XSS-Protection
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}
