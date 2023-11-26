package store

import "sync"

type Repo[T any] struct {
	mu  sync.RWMutex
	col map[string]T
}

func NewRepo[T any]() *Repo[T] {
	return &Repo[T]{
		col: make(map[string]T),
	}
}

func (r *Repo[T]) Has(key string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if _, ok := r.col[key]; ok {
		return true
	}
	return false
}

func (r *Repo[T]) Add(key string, service T) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.col[key]; !ok {
		r.col[key] = service
	}
}

func (r *Repo[T]) Get(key string) T {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.col[key]
}

func (r *Repo[T]) All() map[string]T {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.col
}

func (r *Repo[T]) Iterate(h func(k string, v T)) {
	r.mu.RLock()
	var opQueue []func()
	for k, v := range r.col {
		opQueue = append(opQueue, func() { h(k, v) })
	}
	r.mu.RUnlock()

	for _, op := range opQueue {
		op()
	}
	opQueue = nil // Clear operation queue
}

func (r *Repo[T]) Delete(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.col, key)
}

func (r *Repo[T]) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.col = nil
	r.col = make(map[string]T)
}
