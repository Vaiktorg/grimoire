package internal

import (
	"net/http"
)

func SecurityMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set the Content-Security-Policy header to only allow resources from the same origin
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		// Set the X-Content-Type-Options to prevent the browser from MIME-sniffing a response away from the declared content-type
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Set the Strict-Transport-Security header to enforce secure (HTTP over SSL/TLS) connections to the server
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Set the X-Frame-Options header to provide clickjacking protection
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")

		// Set the X-XSS-Protection header to enable the Cross-site scripting (XSS) filter in the browser
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}
