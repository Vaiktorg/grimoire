package store

import (
	"sync"
)

type Repo[K comparable, V any] struct {
	mu  sync.RWMutex
	col map[K]V
}

func NewRepo[KEY comparable, VAL any]() *Repo[KEY, VAL] {
	return &Repo[KEY, VAL]{
		col: make(map[KEY]VAL),
	}
}

func (r *Repo[K, V]) Add(key K, value V) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.col[key]; !ok {
		r.col[key] = value
	}
}
func (r *Repo[K, V]) Has(key K) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if _, ok := r.col[key]; ok {
		return true
	}
	return false
}
func (r *Repo[K, V]) Get(key K) V {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.col[key]
}

func (r *Repo[K, V]) All() map[K]V {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.col
}
func (r *Repo[K, V]) Slice() []V {
	ret := make([]V, len(r.col))
	r.Iterate(func(_ K, v V) {
		ret = append(ret, v)
	})
	return ret
}

func (r *Repo[K, V]) With(k K, h func(V) error) error {
	return h(r.Get(k))
}
func (r *Repo[K, V]) Iterate(h func(K, V)) {
	r.mu.Lock()

	opQueue := make([]func(), len(r.col))
	idx := 0
	for k, v := range r.col {
		kk, vv := k, v
		opQueue[idx] = func() {
			h(kk, vv)
		}
		idx++
	}

	r.mu.Unlock()

	for _, op := range opQueue {
		op()
	}

	opQueue = nil // Clear operation queue
}

func (r *Repo[K, V]) Delete(k K) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.col, k)
}
func (r *Repo[K, V]) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.col = nil
	r.col = make(map[K]V)
}
