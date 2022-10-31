package ws

import (
	"encoding/json"
)

// IAction is the interface to implement your events.
type IAction interface {
	Validate() error
	Handler(ctx *json.Encoder) error
}

// IMsg is the interface for wrapping your IAction and make them act as dynamic store.
type IMsg interface {
	Key() string
	Process(*json.Decoder, *json.Encoder) error
}
