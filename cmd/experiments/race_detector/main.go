package main

import (
	"fmt"
	"sync"
)

// Counter с race condition
type UnsafeCounter struct {
	value int
}

func (c *UnsafeCounter) Inc() {
	c.value++ // ← RACE!
}

func (c *UnsafeCounter) Value() int {
	return c.value // ← RACE!
}

// Безопасный счётчик с мьютексом
type SafeCounter struct {
	mu    sync.Mutex
	value int
}

func (c *SafeCounter) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value++
}

func (c *SafeCounter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

func main() {
	fmt.Println("=== Unsafe Counter ===")
	unsafe := &UnsafeCounter{}
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			unsafe.Inc()
		}()
	}
	wg.Wait()
	fmt.Printf("Unsafe counter: %d (ожидаем 1000)\n\n", unsafe.Value())

	fmt.Println("=== Safe Counter ===")
	safe := &SafeCounter{}

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			safe.Inc()
		}()
	}
	wg.Wait()
	fmt.Printf("Safe counter: %d (ожидаем 1000)\n", safe.Value())
}
