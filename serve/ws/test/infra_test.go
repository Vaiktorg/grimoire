package ws

import (
	"context"
	"fmt"
	"github.com/vaiktorg/grimoire/serve"
	"github.com/vaiktorg/grimoire/serve/ws"
	"net/http"
	"nhooyr.io/websocket"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var client *ws.Client
var received uint32

func TestMain(m *testing.M) {
	ctx, err := context.WithTimeout(context.Background(), time.Minute)
	if err != nil {
		panic(err)
	}

	conn, _, er := websocket.Dial(ctx, "ws://localhost:8080/", nil)
	if er != nil {
		panic(er)
	}

	client = ws.NewClient(conn)

	m.Run()
}

// TestSend tests the client's ability to send a message.
func TestSend(t *testing.T) {

	numOfMsgs := 10

	wg := new(sync.WaitGroup)
	t.Run("Sender", func(tt *testing.T) {
		tt.Parallel()
		wg.Add(numOfMsgs)
		for i := 0; i < numOfMsgs; i++ {
			client.Send("test data", strconv.Itoa(i))
		}
	})

	t.Run("Receiver", func(tt *testing.T) {
		tt.Parallel()
		client.OnMessage(func(_ *ws.Message) {
			atomic.AddUint32(&received, 1)
			wg.Done()
		})

	})

	wg.Wait()

	fmt.Println(atomic.LoadUint32(&received))
}

func NewTestServer() {
	serv := serve.NewServer(&serve.Config{
		AppName:   "",
		Addr:      "",
		TLSConfig: nil,
		Handler:   nil,
		Logger:    nil,
	})

	serv.Startup(func(mux *serve.MUX) {
		mux.GET("/", func(w http.ResponseWriter, r *http.Request) {

		})

	})
}
