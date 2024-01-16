package ws

import (
	"errors"
	"github.com/vaiktorg/grimoire/log"
	"github.com/vaiktorg/grimoire/store"
	"github.com/vaiktorg/grimoire/uid"
	"github.com/vaiktorg/grimoire/util"
)

type Sessions struct {
	globalId uid.UID
	sessions *store.Repo[uid.UID, *ConnSession]
	logger   log.ILogger
}

type ConnSession struct {
	SessionID uid.UID
	Client    IClient
}

type Config struct {
	GlobalID uid.UID
	Sessions *Sessions
	Logger   log.ILogger
}

func NewConnSessions(config *Config) *Sessions {
	return &Sessions{
		globalId: config.GlobalID,
		sessions: store.NewRepo[uid.UID, *ConnSession](),
		logger:   config.Logger,
	}
}

func (c *Sessions) NewSession(client *Client) error {
	cId := client.ID()
	if c.sessions.Has(cId) {
		return errors.New("client already connected")
	}

	sess := &ConnSession{
		SessionID: uid.NewUID(8),
		Client:    client,
	}

	client.hooks.OnHook(util.OnDisconnect, func(client *Client) {
		c.sessions.Delete(cId)
	})

	c.sessions.Add(cId, sess)
	c.logger.TRACE(string("New Session: "+sess.SessionID+"\n for client: "+client.ID()), sess)
	return nil
}
func (c *Sessions) Disconnect(clientId uid.UID) error {
	if !c.sessions.Has(clientId) {
		return errors.New("session not found for clientID: " + string(clientId))
	}

	c.sessions.Get(clientId).Client.Close()
	c.sessions.Delete(clientId)

	c.logger.TRACE("Disconnected client: " + string(clientId))
	return nil
}
func (c *Sessions) Session(clientId uid.UID) *ConnSession {
	defer c.logger.TRACE("reading session for client: " + string(clientId))
	return c.sessions.Get(clientId)
}
func (c *Sessions) Iterate(handler func(uid.UID, *ConnSession)) {
	c.sessions.Iterate(handler)
}
func (c *Sessions) Sessions() []*ConnSession {
	defer c.logger.TRACE("reading all sessions")
	return c.sessions.Slice()
}
