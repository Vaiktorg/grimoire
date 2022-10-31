package helpers

import (
	"sync"
	"time"
)

type VariableTicker struct {
	C      chan time.Time
	t      *time.Ticker
	speed  time.Duration
	nSpeed time.Duration
	sync.Mutex
}

// NewVariableTicker returns a pointer to an initialized instance of the
// RandomTicker. Min and max are durations of the shortest and longest
// allowed ticks. Ticker will run in a goroutine until explicitly stopped.
func NewVariableTicker(initSpeed time.Duration) *VariableTicker {
	rt := &VariableTicker{
		C:      make(chan time.Time),
		speed:  initSpeed,
		nSpeed: initSpeed,
	}
	rt.t = time.NewTicker(rt.speed)

	go rt.loop()
	return rt
}
func (rt *VariableTicker) SetSpeed(speed time.Duration) {
	rt.Lock()
	defer rt.Unlock()
	rt.nSpeed = speed
}
func (rt *VariableTicker) loop() {
	for c := range rt.t.C {
		if rt.nSpeed != rt.speed {
			rt.speed = rt.nSpeed
			rt.t.Stop()
			rt.t = time.NewTicker(rt.speed)
		}
		rt.C <- c
	}
}

// Stop terminates the ticker goroutine and closes the C channel.
func (rt *VariableTicker) Stop() {
	rt.t.Stop()
	close(rt.C)
}
