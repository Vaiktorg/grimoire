package ws

import (
	"errors"
	"fmt"
)

type Message struct {
	KEY   string `json:"KEY"`
	TYPE  int    `json:"TYPE"`
	Data  any    `json:"Data"`
	Error string `json:"Error"`
}

func (m *Message) Validate() error {
	// Is our Message nil?
	if m == nil {
		return errors.New("message is nil")
	}

	// Is our message invalid?
	if m.KEY == "" {
		return errors.New("message has no key")
	}

	// Is there some sort of Error?
	if m.Error != "" {
		return errors.New(m.Error)
	}

	// Do we have some sort of data?
	if m.Data == nil {
		return errors.New("data for message is nil")
	}

	return nil
}

func (m *Message) Key() string { return m.KEY }

func (m *Message) String() string {
	return fmt.Sprintf("{ \tKEY: %s,\n\tTYPE:%d,\n\tData: %v,\n\tError: %s\n}", m.KEY, m.TYPE, m.Data, m.Error)
}
