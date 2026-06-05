package counter_test

import (
	"testing"

	"checkout-service/internal/counter"
)

// [5.3] Бенчмарки для сравнения производительности подходов.
// Запуск: go test -bench=. -benchmem ./internal/counter/

func BenchmarkUnsafeCounter(b *testing.B) {
	// [5.3] UnsafeCounter быстрый, но НЕБЕЗОПАСНЫЙ.
	// В бенчмарке гонки не детектируются по умолчанию, но результат может быть некорректным.
	// Используй только для понимания «цены» синхронизации.
	c := &counter.UnsafeCounter{}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		c.Increment()
	}
}

func BenchmarkSafeCounter_Atomic(b *testing.B) {
	// [5.1][5.3] Atomic: ~10–30 ns/op, минимальные аллокации.
	// Идеален для счётчиков, флагов, простых инкрементов.
	// 📖 Изучить: почему atomic быстрее Mutex? (нет переключения контекста, нет блокировок)
	c := &counter.SafeCounter{}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		c.Increment()
	}
}

func BenchmarkSafeCounter_Mutex(b *testing.B) {
	// [5.1][5.3] Mutex: ~50–150 ns/op, зависит от конкуренции.
	// Медленнее, но необходим для сложных инвариантов (несколько полей, слайсы, мапы).
	// 📖 Изучить: https://go.dev/blog/mutex (когда Mutex оправдан)
	c := &counter.MutexCounter{}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		c.Increment()
	}
}

// [5.3] Параллельный бенчмарк: имитирует реальную нагрузку с конкуренцией.
// Запуск: go test -bench=. -benchmem -parallel=4 ./internal/counter/
func BenchmarkSafeCounter_Atomic_Parallel(b *testing.B) {
	c := &counter.SafeCounter{}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Increment()
		}
	})
}

func BenchmarkSafeCounter_Mutex_Parallel(b *testing.B) {
	c := &counter.MutexCounter{}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			c.Increment()
		}
	})
}
