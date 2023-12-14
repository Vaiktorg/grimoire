package ws

import (
	"errors"
	"github.com/vaiktorg/grimoire/gwt"
	"github.com/vaiktorg/grimoire/store"
	"github.com/vaiktorg/grimoire/uid"
	"time"
)

const TokenLen = 16
const SessTimeout = 30 * time.Minute

var Spice = gwt.Spice{
	Salt:   uid.NewUID(TokenLen).Bytes(),
	Pepper: uid.NewUID(TokenLen).Bytes(),
}

type Sessions struct {
	globalId uid.UID
	sessions *store.Repo[uid.UID, *ConnSession] // K: ClientID uid.UID -> V: ConnSession
}

type ConnSession struct {
	Token  gwt.Token
	client IClient

	LastActivity time.Time
	SessTimeout  time.Time
	SessCreated  time.Time
}

func NewConnSessions() *Sessions {
	return &Sessions{
		globalId: uid.NewUID(TokenLen),
		sessions: store.NewRepo[uid.UID, *ConnSession](),
	}
}

func (c *Sessions) NewSessionWithToken(client IClient, token gwt.Token) error {
	cId := client.ID()
	if c.sessions.Has(cId) {
		return errors.New("current Client already connected")
	}

	// Get and Delete Client session
	// If disconnected, run a cleanup for session.
	sess := &ConnSession{
		client:      client,
		SessTimeout: time.Now().Add(SessTimeout),
		SessCreated: time.Now(),
		Token:       token,
	}

	c.sessions.Add(cId, sess)

	return nil
}

func (c *Sessions) NewSession(client IClient) error {
	cId := client.ID()
	if c.sessions.Has(cId) {
		return errors.New("current Client already connected")
	}

	// Get and Delete Client session
	// If disconnected, run a cleanup for session.
	sess := &ConnSession{
		client:      client,
		SessTimeout: time.Now().Add(SessTimeout),
		SessCreated: time.Now(),
	}

	enc := gwt.NewEncoder[*ConnSession](Spice)
	ret, err := enc.Encode(sess)
	if err != nil {
		return err
	}
	sess.Token = ret

	c.sessions.Add(cId, sess)

	return nil
}

// Disconnect Append a function to execute onDisconnect
func (c *Sessions) Disconnect(clientId uid.UID) error {
	return c.sessions.With(clientId, func(val *ConnSession) error {
		val.client.Close()
		c.sessions.Delete(clientId)

		return nil
	})
}
func (c *Sessions) Session(clientId uid.UID) *ConnSession {

	if ok := c.sessions.Has(clientId); ok {
		return c.sessions.Get(clientId)
	}

	return nil
}
func (c *Sessions) Sessions() (ret []*ConnSession) {

	c.sessions.Iterate(func(k uid.UID, v *ConnSession) {
		ret = append(ret, v)
	})

	return ret
}
