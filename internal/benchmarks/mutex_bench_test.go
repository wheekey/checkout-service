package benchmarks

import (
	"sync"
	"sync/atomic"
	"testing"
)

// Mutex счётчик
type MutexCounter struct {
	mu    sync.Mutex
	value int
}

func (c *MutexCounter) Inc() {
	c.mu.Lock()
	c.value++
	c.mu.Unlock()
}

func (c *MutexCounter) Value() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.value
}

// Atomic счётчик
type AtomicCounter struct {
	value atomic.Int64
}

func (c *AtomicCounter) Inc() {
	c.value.Add(1)
}

func (c *AtomicCounter) Value() int64 {
	return c.value.Load()
}

// Бенчмарк: 100 горутин, много операций
func BenchmarkMutex_Counter(b *testing.B) {
	c := &MutexCounter{}
	const goroutines = 100
	opsPerGoroutine := b.N / goroutines

	var wg sync.WaitGroup
	wg.Add(goroutines)

	b.ResetTimer()
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				c.Inc()
			}
		}()
	}
	wg.Wait()
}

func BenchmarkAtomic_Counter(b *testing.B) {
	c := &AtomicCounter{}
	const goroutines = 100
	opsPerGoroutine := b.N / goroutines

	var wg sync.WaitGroup
	wg.Add(goroutines)

	b.ResetTimer()
	for i := 0; i < goroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				c.Inc()
			}
		}()
	}
	wg.Wait()
}

// Бенчмарк: только чтение
func BenchmarkMutex_Read(b *testing.B) {
	c := &MutexCounter{}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = c.Value()
		}
	})
}

func BenchmarkAtomic_Read(b *testing.B) {
	c := &AtomicCounter{}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = c.Value()
		}
	})
}
