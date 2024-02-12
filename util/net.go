package util

import (
	"net"
	"net/http"
	"net/url"
	"time"
)

func TryConnection(u url.URL, attempts int, success func(conn *net.Conn) error) error {
	att := attempts
	sleepTime := time.Second

	for i := 0; i < att; i++ {
		conn, err := net.Dial("tcp", u.String())
		if err == nil {
			return success(&conn)
		}

		// Make sure that on the final attempt we won't wait
		if i != att-1 {
			time.Sleep(sleepTime)
			sleepTime *= 2
		}
	}

	panic("could not connect")
}

type Middleware func(handlerFunc http.Handler) http.Handler

// ChainMiddleware chains multiple middleware handlers in logical order.
func ChainMiddleware(h http.Handler, middlewares ...Middleware) http.Handler {
	// Define a wrapper function to apply the middlewares
	var wrapMiddleware func(http.Handler, []Middleware) http.Handler
	wrapMiddleware = func(h http.Handler, mws []Middleware) http.Handler {
		if len(mws) == 0 {
			return h
		}
		return mws[0](wrapMiddleware(h, mws[1:]))
	}

	// Apply middlewares in logical order
	return wrapMiddleware(h, middlewares)
}
