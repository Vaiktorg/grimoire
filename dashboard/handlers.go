package dashboard

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
)

func (d *Dashboard) actionsMux(w http.ResponseWriter, r *http.Request) {
	action := r.URL.Query().Get("action")

	d.logger.TRACE("Action triggered: ", action)

	switch action {
	case "onoff":
		d.onOffHandler(w, r)
	case "running":
		d.runningHandler(w, r)
	}
}

func (d *Dashboard) onOffHandler(w http.ResponseWriter, _ *http.Request) {
	d.running = !d.running

	d.logger.TRACE("Running state toggled..", d.running)

	_, err := io.WriteString(w, strconv.FormatBool(d.running))
	if err != nil {
		println(err)
		return
	}
}
func (d *Dashboard) runningHandler(w http.ResponseWriter, _ *http.Request) {
	d.logger.TRACE("Running state: ", d.running)

	_, err := fmt.Fprint(w, d.running)
	if err != nil {
		println(err)
		return
	}
}

// TODO: getServices handler
func (d *Dashboard) getServices(w http.ResponseWriter, r *http.Request) {
	d.logger.Services()
}
