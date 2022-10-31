package log

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
)

// Client ...
// ==================================================
type Client struct {
	conn   *websocket.Conn
	msgman *MsgMan
}

func NewClient(appname string, conn *websocket.Conn) *Client {
	c := &Client{
		conn:   conn,
		msgman: NewMessageManager(appname),
	}

	c.msgman.SenderFunc(c.Send)

	go c.listen()

	return c
}

func (c *Client) Close() {
	_ = c.conn.Close()
}

func (c *Client) Send(msg Message) {
	jsonBody, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("error during message encoding:", err)
	}

	err = c.conn.WriteMessage(websocket.TextMessage, jsonBody)
	if err != nil {
		fmt.Println("error during message writing:", err)
	}
}
func (c *Client) listen() {
	for {
		msgType, msg, err := c.conn.ReadMessage()
		if err != nil {
			_ = c.conn.Close()
			fmt.Printf("error: %s; connection closed", err)
			return
		}
		if msgType == websocket.CloseMessage {
			_ = c.conn.Close()
			fmt.Printf("close message; connection closed")
			return
		}

		c.msgman.ReadMessage(msg)
	}
}
