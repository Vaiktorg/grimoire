package ws

import (
	"context"
	"github.com/vaiktorg/grimoire/uid"
	"nhooyr.io/websocket"
	"sync"
	"sync/atomic"
)

type IClient interface {
	ID() uid.UID
	Send(string, any)
	Error(error)
	Close()
}

type Client struct {
	mu     sync.RWMutex
	conn   *websocket.Conn
	secret []byte

	id     uid.UID
	isGob  bool
	closed atomic.Bool

	WriterChan chan Message
	ReaderChan chan Message

	wgHandler    sync.WaitGroup
	onConnect    []func(IClient)
	onDisconnect []func(IClient)
}

func NewClient(conn *websocket.Conn) *Client {
	id := uid.NewUID(8)
	client := &Client{
		conn:       conn,
		id:         id,
		WriterChan: make(chan Message, 100),
		ReaderChan: make(chan Message, 100),
	}

	client.execOnConnect()

	go client.listenIncoming()
	go client.listenOutgoing()

	return client
}

func (c *Client) ID() uid.UID {
	return c.id
}

func (c *Client) Send(key string, data any) {
	msg := Message{
		KEY:  key,
		Data: data,
	}

	c.WriterChan <- msg
}
func (c *Client) OnMessage(handler func(Message)) {
	for m := range c.ReaderChan {
		if handler != nil {
			handler(m)
		}
	}
}
func (c *Client) Error(err error) {
	if err == nil {
		return
	}

	msg := Message{
		KEY:   "error",
		Error: err.Error(),
	}

	c.WriterChan <- msg
}

// OnConnect appends a handler that will be called at the end of the Client's connection sequence.
func (c *Client) OnConnect(handler ...func(client IClient)) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.onConnect = append(c.onConnect, handler...)
}

// OnDisconnect appends a handler that will be called at the end of the Client's disconnection sequence.
func (c *Client) OnDisconnect(handler ...func(client IClient)) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.onDisconnect = append(c.onDisconnect, handler...)
}

func (c *Client) Close() {
	if c.closed.Load() {
		return
	}

	c.closed.Store(true)

	c.mu.Lock()
	defer c.mu.Unlock()

	println("disconnecting")

	c.execOnDisconnect()

	_ = c.conn.Close(websocket.StatusNormalClosure, "bye bye")

	close(c.WriterChan)
	close(c.ReaderChan)

	println("closed and disconnected")
}

func (c *Client) execOnConnect() {
	c.wgHandler.Add(len(c.onConnect))
	for _, f := range c.onConnect {
		go func(hf func(client IClient), wg *sync.WaitGroup) { hf(c); wg.Done() }(f, &c.wgHandler)
	}

	c.wgHandler.Wait()
}

func (c *Client) execOnDisconnect() {
	c.wgHandler.Add(len(c.onDisconnect))
	for _, f := range c.onDisconnect {
		go func(hf func(client IClient), wg *sync.WaitGroup) { hf(c); wg.Done() }(f, &c.wgHandler)
	}

	c.wgHandler.Wait()
}

func (c *Client) listenIncoming() {
	for {
		_, buf, err := c.conn.Read(context.Background())
		if err != nil {
			c.Close()
			return
		}

		var msg Message
		err = msg.Decode(JSON, buf)
		if err != nil {
			continue
		}

		c.ReaderChan <- msg
	}
}
func (c *Client) listenOutgoing() {
	for msg := range c.WriterChan {
		buf, err := msg.Encode(JSON)
		if err != nil {
			continue
		}

		if err = c.conn.Write(context.Background(), websocket.MessageText, buf); err != nil {
			continue
		}
	}
}
