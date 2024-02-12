package util

import "github.com/vaiktorg/grimoire/store"

type Hook string

const (
	OnConnect    Hook = "connect"
	OnDisconnect      = "disconnect"
)

func (e *Hook) Is(hookType Hook) bool {
	return *e == hookType
}

type Hooks[T any] struct {
	hookChan chan func(T)
	onHook   store.Repo[Hook, func(T)]
}

func NewHookEvents[T any](data T) *Hooks[T] {
	h := &Hooks[T]{
		hookChan: make(chan func(T)),
		onHook:   store.Repo[Hook, func(T)]{},
	}

	go h.listenForHooks(data)

	return h
}

func (s *Hooks[T]) OnHook(hookType Hook, handler func(T)) {
	s.onHook.Add(hookType, handler)
}

func (s *Hooks[T]) listenForHooks(data T) {
	for event := range s.hookChan {
		event(data)
	}
}

func (s *Hooks[T]) EnqueueHook(hookType Hook) {
	s.hookChan <- func(t T) {
		if !s.onHook.Has(hookType) {
			return
		}

		if h := s.onHook.Get(hookType); h != nil {
			h(t)
		}
	}
}
