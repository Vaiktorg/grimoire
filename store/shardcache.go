package store

import (
	"hash"
	"hash/fnv"
	"sync"
)

type Shard[T any] struct {
	items map[string]T
	mu    sync.RWMutex
}

func NewShard[T any]() *Shard[T] {
	return &Shard[T]{
		items: make(map[string]T),
	}
}

type ShardCache[T any] struct {
	mu     sync.Mutex
	shards []*Shard[T]
	fnv    hash.Hash32
	count  int
}

func NewShardCache[T any](shardCount int) *ShardCache[T] {
	cache := &ShardCache[T]{
		shards: make([]*Shard[T], shardCount),
		count:  shardCount,
		fnv:    fnv.New32(),
	}
	for i := 0; i < shardCount; i++ {
		cache.shards[i] = NewShard[T]()
	}

	cache.count = shardCount

	return cache
}

// Define hashing function
func (c *ShardCache[T]) getShardIndex(key string) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.fnv.Reset()

	_, _ = c.fnv.Write([]byte(key))
	return int(c.fnv.Sum32()) % c.count
}
func (s *Shard[T]) set(key string, value T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[key] = value
}
func (s *Shard[T]) get(key string) (T, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, ok := s.items[key]
	return val, ok
}
func (s *Shard[T]) delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.items, key)
}

func (c *ShardCache[T]) Set(key string, value T) {
	c.shards[c.getShardIndex(key)].set(key, value)
}
func (c *ShardCache[T]) Get(key string) (T, bool) {
	return c.shards[c.getShardIndex(key)].get(key)
}
func (c *ShardCache[T]) Delete(key string) {
	c.shards[c.getShardIndex(key)].delete(key)
}
