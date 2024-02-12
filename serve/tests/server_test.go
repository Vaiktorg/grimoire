package tests

import (
	"fmt"
	"github.com/vaiktorg/grimoire/log"
	"github.com/vaiktorg/grimoire/names"
	"github.com/vaiktorg/grimoire/serve"
	"github.com/vaiktorg/grimoire/serve/ws"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMain(m *testing.M) {
	m.Run()
}

func TestServer(t *testing.T) {
	t.Cleanup(func() {
		// Add cleanup logic if needed
	})

	serverName := names.NewLastName("Server")
	serverLogger := log.NewLogger(&log.Config{
		CanOutput:   true,
		Persist:     true,
		ServiceName: serverName,
	})

	serverCfg := &serve.Config{
		AppName: serverName,
		Logger:  serverLogger,
	}

	s := serve.NewServer(serverCfg)

	t.Run("WebSocket", func(t *testing.T) {
		// Setup WebSocket test
		s.WebSocket(func(socket ws.IWebSocket) {
			socket.OnMessage(func(message ws.Message, session *ws.ConnSession) {
				if message.KEY != "bish" {
					t.Errorf("wrong msg")
					return
				}
			})

			conn, err := socket.Dial("ws://localhost:8080/websocket")
			if err != nil {
				t.Errorf("error dialing WebSocket: %v", err)
				return
			}

			conn.Send("bish", nil)
		})
	})

	t.Run("HTTP_GET", func(t *testing.T) {
		s.MUX(func(mux *http.ServeMux) {
			mux.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, "hello world")
			})
		})

		req, err := http.NewRequest("GET", "/get", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "hello world")
		})

		handler.ServeHTTP(rr, req)
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		expected := "hello world"
		if rr.Body.String() != expected {
			t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
		}
	})

	t.Run("HTTP_POST", func(t *testing.T) {
		s.MUX(func(mux *http.ServeMux) {
			mux.HandleFunc("/post", func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("incorrect method: got %v want %v", r.Method, http.MethodPost)
				}
			})
		})

		req, err := http.NewRequest("POST", "/post", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("incorrect method: got %v want %v", r.Method, http.MethodPost)
			}
		})

		handler.ServeHTTP(rr, req)
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
	})

	// Start the server in a separate goroutine
	go func() {
		s.ListenAndServe()
	}()

	// Add necessary delays or checks for server startup if needed
}
