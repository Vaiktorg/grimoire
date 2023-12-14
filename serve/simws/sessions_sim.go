package simws

import (
	"errors"
	"github.com/vaiktorg/grimoire/gwt"
	"github.com/vaiktorg/grimoire/serve/ws"
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

type SimSessions struct {
	globalId uid.UID
	sessions *store.Repo[uid.UID, *SimConnSession]
}

type SimConnSession struct {
	Token        gwt.Token
	Client       ws.IClient
	LastActivity time.Time
	SessTimeout  time.Time
	SessCreated  time.Time
}

func NewSimConnSessions() *SimSessions {
	return &SimSessions{
		globalId: uid.NewUID(ws.TokenLen),
		sessions: store.NewRepo[uid.UID, *SimConnSession](),
	}
}

func (c *SimSessions) NewSession(client ws.IClient, token *gwt.Token) error {
	cId := client.ID()
	if c.sessions.Has(cId) {
		return errors.New("client already connected")
	}

	sess := &SimConnSession{
		Client:      client,
		SessTimeout: time.Now().Add(ws.SessTimeout),
		SessCreated: time.Now(),
	}

	if token != nil {
		sess.Token = *token
	} else {
		enc := gwt.NewEncoder[*SimConnSession](Spice)
		ret, err := enc.Encode(sess)
		if err != nil {
			return err
		}
		sess.Token = ret
	}

	c.sessions.Add(cId, sess)
	return nil
}
func (c *SimSessions) Disconnect(clientId uid.UID) error {
	if !c.sessions.Has(clientId) {
		return errors.New("session not found for clientID: " + clientId.String())
	}

	c.sessions.Get(clientId).Client.Close()
	c.sessions.Delete(clientId)

	return nil
}
func (c *SimSessions) Session(clientId uid.UID) *SimConnSession {
	return c.sessions.Get(clientId)
}
func (c *SimSessions) Sessions() []*SimConnSession {
	return c.sessions.Slice()
}
