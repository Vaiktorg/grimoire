package simws

import (
	"errors"
	"github.com/vaiktorg/grimoire/log"
	"github.com/vaiktorg/grimoire/store"
	"github.com/vaiktorg/grimoire/uid"
	"github.com/vaiktorg/grimoire/util"
)

const TokenLen = 16

type SimSessions struct {
	GlobalId []byte
	sessions *store.Repo[uid.UID, *ConnSession] // K: ClientID []byte -> V: ConnSession
}
type Config struct {
	GlobalID uid.UID
	Sessions *SimSessions
	Logger   log.ILogger
}
type ConnSession struct {
	SessionID uid.UID
	Client    *Client
}

func NewSimConnSessions() *SimSessions {
	return &SimSessions{
		GlobalId: uid.NewUID(TokenLen).Bytes(),
		sessions: store.NewRepo[uid.UID, *ConnSession](),
	}
}

func (c *SimSessions) NewSession(client *Client) error {
	cId := client.ID()
	if c.sessions.Has(cId) {
		return errors.New("client already exists")
	}

	// Get and Delete Client session
	// If disconnected, run a cleanup for session.
	sess := &ConnSession{
		SessionID: client.ID(),
		Client:    client,
	}

	client.hooks.OnHook(util.OnDisconnect, func(client *Client) {
		_ = c.Disconnect(client.ID())
	})

	c.sessions.Add(cId, sess)
	return nil
}
func (c *SimSessions) Session(clientId uid.UID) *ConnSession {

	if ok := c.sessions.Has(clientId); ok {
		return c.sessions.Get(clientId)
	}

	return nil
}
func (c *SimSessions) Sessions() (ret []*ConnSession) {

	c.sessions.Iterate(func(k uid.UID, v *ConnSession) {
		ret = append(ret, v)
	})

	return ret
}
func (c *SimSessions) Iterate(handler func(id uid.UID, session *ConnSession)) {
	c.sessions.Iterate(handler)
}

// Disconnect Append a function to execute onDisconnect
func (c *SimSessions) Disconnect(clientId uid.UID) error {
	return c.sessions.With(clientId, func(val *ConnSession) error {
		val.Client.Close()
		c.sessions.Delete(clientId)

		return nil
	})
}
