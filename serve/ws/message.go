package ws

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
)

type Message struct {
	KEY   string `json:"KEY"`
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
	return fmt.Sprintf("{ KEY: %s, Data: %v, Error: %s }", m.KEY, m.Data, m.Error)
}

type EncDecType uint8

const (
	JSON EncDecType = iota
	GOB
)

func (m *Message) encGob() ([]byte, error) {
	buff := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buff)

	err := enc.Encode(&m)
	return buff.Bytes(), err
}

func (m *Message) encJson() ([]byte, error) {
	return json.Marshal(&m)
}

func (m *Message) decGob(b []byte) error {
	return gob.NewDecoder(bytes.NewReader(b)).Decode(&m)
}

func (m *Message) decJson(b []byte) error {
	return json.Unmarshal(b, &m)
}

func (m *Message) Decode(decType EncDecType, buf []byte) error {
	switch decType {
	case JSON:
		return m.decJson(buf)
	case GOB:
		return m.decGob(buf)

	default:
		return errors.New("unknown decoding type")
	}
}

func (m *Message) Encode(encType EncDecType) ([]byte, error) {
	switch encType {
	case JSON:
		return m.encJson()
	case GOB:
		return m.encGob()

	default:
		return nil, errors.New("unknown encoding type")
	}
}
