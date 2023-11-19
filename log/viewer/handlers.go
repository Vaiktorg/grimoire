package viewer

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
)

func (v *Viewer) ActionsMux(w http.ResponseWriter, r *http.Request) {
	action := r.URL.Query().Get("action")
	switch action {
	case "onoff":
		v.onOffHandler(w, r)
	case "running":
		v.runningHandler(w, r)
	}
}

func (v *Viewer) onOffHandler(w http.ResponseWriter, _ *http.Request) {
	v.running = !v.running

	_, err := io.WriteString(w, strconv.FormatBool(v.running))
	if err != nil {
		println(err)
		return
	}
}
func (v *Viewer) runningHandler(w http.ResponseWriter, _ *http.Request) {
	_, err := fmt.Fprint(w, v.running)
	if err != nil {
		println(err)
		return
	}
}

// TODO: getServices handler
func (v *Viewer) getServices(w http.ResponseWriter, r *http.Request) {
	//v.logger.GetServices()
}
