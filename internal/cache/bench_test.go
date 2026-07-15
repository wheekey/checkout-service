package cache

import (
	"fmt"
	"sync"
	"testing"
)

// ThreadSafeLRU — обычный LRU с одним мьютексом (для сравнения)
type ThreadSafeLRU struct {
	mu  sync.Mutex
	lru *LRU
}

func NewThreadSafeLRU(capacity int) *ThreadSafeLRU {
	return &ThreadSafeLRU{lru: NewLRU(capacity)}
}

func (c *ThreadSafeLRU) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.lru.Get(key)
}

func (c *ThreadSafeLRU) Put(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.lru.Put(key, value)
}

// Бенчмарк: обычный thread-safe LRU
func BenchmarkThreadSafeLRU(b *testing.B) {
	cache := NewThreadSafeLRU(1000)

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key_%d", i%1000)
			if i%10 == 0 {
				cache.Put(key, i)
			} else {
				cache.Get(key)
			}
			i++
		}
	})
}

// Бенчмарк: шардированный LRU
func BenchmarkShardedLRU(b *testing.B) {
	cache := NewSharded(1000, 16)

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key_%d", i%1000)
			if i%10 == 0 {
				cache.Put(key, i)
			} else {
				cache.Get(key)
			}
			i++
		}
	})
}

// Бенчмарк с "hot keys" (маленький набор популярных ключей)
func BenchmarkShardedLRU_HotKeys(b *testing.B) {
	cache := NewSharded(1000, 16)

	// Предзаполняем "популярные" ключи
	for i := 0; i < 100; i++ {
		cache.Put(fmt.Sprintf("hot_%d", i), i)
	}

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("hot_%d", i%100)
			cache.Get(key)
			i++
		}
	})
}
