package simws

import (
	"context"
	"github.com/vaiktorg/grimoire/serve/ws"
	"github.com/vaiktorg/grimoire/uid"
	"nhooyr.io/websocket"
	"runtime/debug"
	"sync"
)

type SimClient struct {
	id uid.UID
	mu sync.Mutex

	closed bool

	conn *websocket.Conn

	eventChan chan func(client ws.IClient)
	onEvent   func(eventType EventType, client ws.IClient)
}

func NewSimClient(conn *websocket.Conn) *SimClient {
	return &SimClient{
		conn:      conn,
		id:        uid.NewUID(8),
		eventChan: make(chan func(client ws.IClient), 10),
	}
}

func (c *SimClient) ID() uid.UID {
	return c.id
}

func (c *SimClient) Send(key string, data any) {
	m := ws.Message{
		KEY:  key,
		Data: data,
	}

	buf, err := m.Encode(ws.JSON)
	if err != nil {
		return
	}

	_ = c.conn.Write(context.Background(), websocket.MessageText, buf)
}

func (c *SimClient) Error(e error) {
	m := &ws.Message{
		KEY:   "error",
		Data:  debug.Stack(),
		Error: e.Error(),
	}

	buf, err := m.Encode(ws.JSON)
	if err != nil {
		return
	}

	_ = c.conn.Write(context.Background(), websocket.MessageText, buf)
}

func (c *SimClient) Close() {
	c.mu.Lock()

	if c.closed {
		c.mu.Unlock()
		return
	}

	c.closed = true
	c.mu.Unlock()

	c.enqueueEvents(OnDisconnect)

	_ = c.conn.Close(websocket.StatusNormalClosure, "bye bye")
	close(c.eventChan)
}

func (c *SimClient) Listen(onMessage func(msg ws.Message)) {
	go c.listenForEvents()

	for {
		_, buf, err := c.conn.Read(context.Background())
		if err != nil {
			c.Close()
			return
		}

		var msg ws.Message
		if err = msg.Decode(ws.JSON, buf); err != nil {
			continue
		}

		onMessage(msg)
	}
}

func (c *SimClient) OnEvent(handler func(eventType EventType, client ws.IClient)) {
	c.onEvent = handler
}

func (c *SimClient) listenForEvents() {
	for event := range c.eventChan {
		event(c)
	}
}

func (c *SimClient) enqueueEvents(evType EventType) {
	c.eventChan <- func(client ws.IClient) {
		if c.onEvent != nil {
			c.onEvent(evType, client)
		}
	}
}
