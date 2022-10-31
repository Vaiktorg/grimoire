package log

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
)

type (
	Method  string                                      // What kind of request Method is it? "get", "send", "update", "delete"
	Service = string                                    // What Service you're targeting? ex: "authentication"
	Command = string                                    // The Command within a Service that will be used as a representation of the Action Name. ex: "login"
	Action  func(reqBody []byte, response Sender) error // Handler func that corresponds to your Command's execution.
)

const (
	GET    = Method("GET")
	SEND   = Method("SND")
	UPDATE = Method("UPD")
	DELETE = Method("DEL")

	RUN    = Command("RUN")
	CANCEL = Command("CNL")
)

// ==================================================

// Message ...
type Message struct {
	Method  Method  // Should we send data or run an action [SND], or get some data [GET]
	Service Service // Which service should we apply the action?
	Command Command // Which action/command should we execute?
	Body    []byte  // Data to send or receive, depending on Method
}

func (m *Message) SetBody(body []byte) error {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return err
	}

	m.Body = jsonBody
	return nil
}

func (m *Message) Parse(msg []byte) error {
	b := bytes.NewBuffer(msg)

	err := json.NewEncoder(b).Encode(m)
	if err != nil {
		return err
	}

	if m.Service == "" ||
		m.Method == "" ||
		m.Command == "" {
		return errors.New("message has no service, method or Command")
	}

	return nil
}

func (m *Message) String() string {
	cmd := m.Key()
	if m.Body != nil && len(m.Body) > 0 {
		b := base64.StdEncoding.EncodeToString(m.Body)
		cmd = strings.Join([]string{cmd, b}, ":")
	}
	return cmd
}

func (m *Message) Key() string {
	return strings.Join([]string{string(m.Method), m.Service, m.Command}, ".")
}
