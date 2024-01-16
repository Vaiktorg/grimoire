package simws

import (
	"errors"
	"fmt"
)

type Message struct {
	KEY  string `json:"KEY"`
	Data any    `json:"Data"`
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

	// Do we have some sort of data?
	if m.Data == nil {
		return errors.New("data for message is nil")
	}

	return nil
}

func (m *Message) Key() string { return m.KEY }

func (m *Message) String() string {
	return fmt.Sprintf("{ KEY: %s, Data: %v }", m.KEY, m.Data)
}
