package grimoire

import (
	"errors"
	"time"
)

// Layer ======================================================
// ============================================================
type Layer struct {
	layerName string
	keys      map[time.Duration]func()
}

func (l *Layer) NewLayer(name string) *Layer {
	return &Layer{layerName: name, keys: make(map[time.Duration]func())}
}
func (l *Layer) AddKeyframe(timestamp time.Duration, action func()) {
	if _, ok := l.keys[timestamp]; !ok {
		l.keys[timestamp] = action
	}
}
func (l *Layer) ClearKeyframe(timestamp time.Duration) {
	if _, ok := l.keys[timestamp]; ok {
		delete(l.keys, timestamp)
	}
}

// Timeline ===================================================
// ============================================================
type Timeline struct {
	layers    []Layer
	finihed   chan struct{}
	err       chan error
	length    time.Duration
	accumTime time.Duration
	currIdx   int
}

// NewTimeline Creates a new ...
func NewTimeline(duration time.Duration) *Timeline {
	t := &Timeline{
		finihed: make(chan struct{}),
		err:     make(chan error),
		length:  duration,
	}
	return t
}

func (t *Timeline) Play() {
	go t.startTimeline()
}

func (t *Timeline) Wait() <-chan struct{} {
	return t.finihed
}

func (t *Timeline) ClearLayer(idx int) error {
	if idx < len(t.layers) || idx > len(t.layers) {
		return errors.New("layer idx does not exist")
	}

	t.layers = append(t.layers[:idx], t.layers[idx+1:]...)
	return nil
}

func (t *Timeline) startTimeline() {
	tick := time.NewTicker(time.Millisecond)
	defer tick.Stop()

	if len(t.layers) == 0 {
		t.err <- errors.New("no keyframes found")
	}

	for range tick.C {
		for _, layer := range t.layers {
			if action, ok := layer.keys[t.accumTime]; ok {
				action()
			}
		}
		if t.accumTime == t.length {
			t.finihed <- struct{}{}
			break
		}

		t.accumTime += time.Millisecond
	}
}
