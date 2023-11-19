package log

import "sync"

type Repo[T any] struct {
	mu       sync.Mutex
	services map[string]T
}

func NewRepo[T any]() *Repo[T] {
	return &Repo[T]{
		services: make(map[string]T),
	}
}

func (r *Repo[T]) Has(key string) bool {
	if _, ok := r.services[key]; ok {
		return true
	}
	return false
}

func (r *Repo[T]) Add(key string, service T) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.services[key]; !ok {
		r.services[key] = service
	}
}

func (r *Repo[T]) Get(key string) T {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.services[key]
}

func (r *Repo[T]) All() map[string]T {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.services
}

func (r *Repo[T]) Iterate(h func(T)) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, v := range r.services {
		h(v)
	}
}

func (r *Repo[T]) Delete(key string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.services, key)
}

func (r *Repo[T]) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.services = nil
	r.services = make(map[string]T)
}
