package simws

import (
	"context"
	"github.com/vaiktorg/grimoire/log"
	"github.com/vaiktorg/grimoire/uid"
	"github.com/vaiktorg/grimoire/util"
	"nhooyr.io/websocket"
	"sync"
)

type Client struct {
	id uid.UID
	mu sync.Mutex

	closed bool

	conn   *websocket.Conn
	mc     *util.MultiCoder[Message]
	logger log.ILogger

	hooks *util.Hooks[*Client]
}

func NewSimClient(conn *websocket.Conn) (*Client, error) {
	mc, err := util.NewMultiCoder[Message]()
	if err != nil {
		return nil, err
	}

	client := &Client{
		conn: conn,
		id:   uid.NewUID(8),
		mc:   mc,
	}
	client.hooks = util.NewHookEvents[*Client](client)

	return client, nil
}

func (c *Client) ID() uid.UID {
	return c.id
}

func (c *Client) Send(key string, data any) {
	m := Message{
		KEY:  key,
		Data: data,
	}

	buf, err := c.mc.Encode(m, util.EncodeJson)
	if err != nil {
		return
	}

	_ = c.conn.Write(context.Background(), websocket.MessageText, buf)
}

func (c *Client) Error(e error) {
	m := Message{
		KEY:  "error",
		Data: e.Error(),
	}

	buf, err := c.mc.Encode(m, util.EncodeJson)
	if err != nil {
		return
	}

	_ = c.conn.Write(context.Background(), websocket.MessageText, buf)
}

func (c *Client) Close() {
	c.mu.Lock()

	if c.closed {
		c.mu.Unlock()
		return
	}

	c.closed = true
	c.mu.Unlock()

	c.hooks.EnqueueHook(util.OnDisconnect)

	_ = c.conn.Close(websocket.StatusNormalClosure, "bye bye")
	// TODO: CLose HOOKS
}

func (c *Client) Listen(onMessage func(msg Message)) {
	defer c.Close()

	for {
		_, buf, err := c.conn.Read(context.Background())
		if err != nil {
			c.logger.ERROR(err.Error())
			continue
		}

		m, err := c.mc.Decode(buf, util.DecodeJson)
		if err != nil {
			c.logger.ERROR(err.Error())
			continue
		}

		onMessage(m)
	}
}
