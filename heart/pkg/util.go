package heart

import (
	"encoding/json"
	"fmt"
	"io"
)

func bindHeartbeat(r io.ReadCloser) (*Heartbeat, error) {
	b := &Heartbeat{}
	err := json.NewDecoder(r).Decode(&b)
	if err != nil {
		return &Heartbeat{}, err
	}

	return b, nil
}

//-------------------------------------------------------------------
func (h *Heartbeat) recover() {
	if e := recover().(error); e != nil {
		fmt.Println(e.Error())
	}
}
