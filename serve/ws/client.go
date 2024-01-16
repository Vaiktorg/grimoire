package ws

import (
	"context"
	"github.com/vaiktorg/grimoire/serve/simws"
	"github.com/vaiktorg/grimoire/uid"
	"github.com/vaiktorg/grimoire/util"
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
	multiCoder *util.MultiCoder[Message]

	wgHandler sync.WaitGroup

	hooks util.Hooks[*Client]
}

func NewClient(conn *websocket.Conn) (*Client, error) {
	id, err := uid.NewSecureUID(simws.TokenLen)
	if err != nil {
		return nil, err
	}

	mc, err := util.NewMultiCoder[Message]()
	if err != nil {
		return nil, err
	}

	client := &Client{
		conn:       conn,
		id:         id,
		WriterChan: make(chan Message, 100),
		ReaderChan: make(chan Message, 100),
		multiCoder: mc,
	}

	client.hooks.EnqueueHook(util.OnConnect)

	go client.listenIncoming()
	go client.listenOutgoing()

	return client, nil
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

func (c *Client) Close() {
	if c.closed.Load() {
		return
	}

	c.closed.Store(true)

	c.mu.Lock()
	defer c.mu.Unlock()

	println("disconnecting")

	c.hooks.EnqueueHook(util.OnDisconnect)

	_ = c.conn.Close(websocket.StatusNormalClosure, "bye bye")

	close(c.WriterChan)
	close(c.ReaderChan)

	println("closed and disconnected")
}

func (c *Client) listenIncoming() {
	for {
		_, buf, err := c.conn.Read(context.Background())
		if err != nil {
			c.Close()
			return
		}

		msg, err := c.multiCoder.DecodeDecrypt(buf, util.DecodeGob)
		if err != nil {
			continue
		}

		c.ReaderChan <- msg
	}
}
func (c *Client) listenOutgoing() {
	for msg := range c.WriterChan {
		buf, err := c.multiCoder.EncodeEncrypt(msg, util.EncodeGob)
		if err != nil {
			continue
		}

		if err = c.conn.Write(context.Background(), websocket.MessageType(msg.TYPE), buf); err != nil {
			continue
		}
	}
}
