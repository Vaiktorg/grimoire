package util

import (
	"net"
	"net/http"
	"net/url"
	"time"
)

func TryConnection(u url.URL, header http.Header, attempts int, success func(conn *net.Conn) error) error {
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
