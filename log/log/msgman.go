package log

import (
	"fmt"
)

type Sender func(msg Message)

// MsgMan ...
// ==================================================
type MsgMan struct {
	actions  map[string]Action //K: Method . Service . Command; V: Action
	messages chan []byte
	sender   Sender
	appname  string
}

func NewMessageManager(appname string) *MsgMan {
	mm := &MsgMan{
		messages: make(chan []byte),
		appname:  appname,
	}

	go mm.listen()

	return mm
}

func (m *MsgMan) listen() {
	for {
		select {
		case msg := <-m.messages:
			mm := new(Message)
			err := mm.Parse(msg)
			if err != nil {
				println(err.Error())
				continue
			}

			key := mm.Key()
			if action, ok := m.actions[key]; ok {
				err = action(mm.Body, m.sender)
				if err != nil {
					println(err.Error())
					continue
				}
			} else {
				fmt.Printf("no actions found for key: %s\n", key)
			}
		}
	}
}

func (m *MsgMan) ReadMessage(msg []byte) {
	m.messages <- msg
}
func (m *MsgMan) SenderFunc(sender Sender) {
	m.sender = sender
}
