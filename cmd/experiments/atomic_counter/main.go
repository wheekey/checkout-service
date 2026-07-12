package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// Счётчик на Mutex
type MutexCounter struct {
	mu    sync.Mutex
	value int
}

func (c *MutexCounter) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value++
}

func (c *MutexCounter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

// Счётчик на atomic
type AtomicCounter struct {
	value atomic.Int64
}

func (c *AtomicCounter) Inc() {
	c.value.Add(1)
}

func (c *AtomicCounter) Value() int64 {
	return c.value.Load()
}

func main() {
	const goroutines = 1000
	const iterations = 1000

	// Тест Mutex
	mutexCounter := &MutexCounter{}
	var wg sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				mutexCounter.Inc()
			}
		}()
	}
	wg.Wait()
	fmt.Printf("Mutex counter: %d (ожидаем %d)\n", mutexCounter.Value(), goroutines*iterations)

	// Тест atomic
	atomicCounter := &AtomicCounter{}

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				atomicCounter.Inc()
			}
		}()
	}
	wg.Wait()
	fmt.Printf("Atomic counter: %d (ожидаем %d)\n", atomicCounter.Value(), goroutines*iterations)
}
