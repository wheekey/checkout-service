package main

import (
	"fmt"
	"sync"
	"testing"
)

// Вариант 1: Канал
func benchmarkChannel(b *testing.B) {
	ch := make(chan int, 1)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			ch <- 1
			<-ch
		}
	})
}

// Вариант 2: Mutex + переменная
type MutexCounter struct {
	mu    sync.Mutex
	value int
}

func (mc *MutexCounter) Increment() {
	mc.mu.Lock()
	mc.value++
	mc.mu.Unlock()
}

func (mc *MutexCounter) Decrement() {
	mc.mu.Lock()
	mc.value--
	mc.mu.Unlock()
}

func benchmarkMutex(b *testing.B) {
	mc := &MutexCounter{}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mc.Increment()
			mc.Decrement()
		}
	})
}

func main() {
	fmt.Println("Benchmarking channels vs mutex...")
	fmt.Println()

	result1 := testing.Benchmark(benchmarkChannel)
	fmt.Printf("Channel:    %v\n", result1)

	result2 := testing.Benchmark(benchmarkMutex)
	fmt.Printf("Mutex:      %v\n", result2)

	fmt.Println()
	fmt.Println("Вывод:")
	fmt.Println("- Каналы: удобные, но медленнее (~2x)")
	fmt.Println("- Mutex: самый быстрый для простых операций")
	fmt.Println()
	fmt.Println("sync.Cond не бенчмаркается через RunParallel,")
	fmt.Println("потому что он для producer-consumer паттернов.")
}
