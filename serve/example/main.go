package main

import (
	"fmt"
	"github.com/vaiktorg/grimoire/serve"
	"github.com/vaiktorg/grimoire/serve/ws"
	"github.com/vaiktorg/grimoire/util"
	"net/http"
)

func main() {
	serv := serve.NewServer(&serve.Config{
		AppName: "ExampleApp",
	})

	// Server Startup
	serv.Startup(func(server serve.AppConfig) {
		// Websocket Configuration
		server.WebSocket(func(socket ws.IWebSocket) {
			// Event Handler
			socket.OnHook(func(evType util.Hook, sess *ws.ConnSession) {
				switch evType {
				case util.OnConnect:
					fmt.Println("OnConnect: " + sess.SessionID)
				case util.OnDisconnect:
					fmt.Println("OnDisconnect: " + sess.SessionID)
				}
			})

			// Message Handler
			socket.OnMessage(func(msg ws.Message, client *ws.Client) {
				fmt.Println(msg.String())
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
