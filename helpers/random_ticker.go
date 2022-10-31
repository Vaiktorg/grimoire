package helpers

import (
	"math/rand"
	"time"
)

type RandomTicker struct {
	C   chan time.Time
	t   *time.Ticker
	min int64
	max int64

	closer chan struct{}
}

// NewRandomTicker returns a pointer to an initialized instance of the
// RandomTicker. Min and max are durations of the shortest and longest
// allowed ticks. Ticker will run in a goroutine until explicitly stopped.
func NewRandomTicker(min, max time.Duration) *RandomTicker {
	rt := &RandomTicker{
		C:   make(chan time.Time),
		min: min.Nanoseconds(),
		max: max.Nanoseconds(),
	}
	rt.t = time.NewTicker(rt.nextInterval())

	go rt.loop()
	return rt
}

func (rt *RandomTicker) loop() {
	for val := range rt.t.C {
		rt.C <- val
		rt.t.Stop()
		rt.t = time.NewTicker(rt.nextInterval())
	}
}

// Stop terminates the ticker goroutine and closes the C channel.
func (rt *RandomTicker) Stop() {
	rt.t.Stop()
}

func (rt *RandomTicker) nextInterval() time.Duration {
	interval := rand.Int63n(rt.max-rt.min) + rt.min
	return time.Duration(interval) * time.Nanosecond
}
