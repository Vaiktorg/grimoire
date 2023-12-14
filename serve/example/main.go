package main

import (
	"fmt"
	"github.com/vaiktorg/grimoire/serve"
	"github.com/vaiktorg/grimoire/serve/simws"
	"github.com/vaiktorg/grimoire/serve/ws"
	"net/http"
)

func main() {
	serv := serve.NewServer(&serve.Config{
		AppName: "ExampleApp",
	})

	// Server Startup
	serv.Startup(func(server serve.AppConfig) {
		// Websocket Configuration
		server.WebSocket(func(socket simws.ISimWebSocket) {

			// Event Handler
			socket.OnEvent(func(evType simws.EventType, sess *simws.SimConnSession) {
				switch evType {
				case simws.OnConnect:
					fmt.Println("OnConnect: " + sess.Token.Token)
				case simws.OnDisconnect:
					fmt.Println("OnDisconnect: " + sess.Token.Token)
				}
			})

			// Message Handler
			socket.OnMessage(func(msg ws.Message, sess *simws.SimConnSession) {
				println(msg.String() + " : " + sess.Token.Token)
			})
		})

		server.MUX(func(mux *http.ServeMux) {
			mux.HandleFunc("/", HomeHandler)
		})
	})

	serv.ListenAndServe()
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	_, err := fmt.Fprintln(w, "hello world: "+r.RemoteAddr)
	if err != nil {
		return
	}
}
