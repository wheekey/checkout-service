package counter

import "sync/atomic" // [5.1] Стандартная библиотека для атомарных операций

// SafeCounter использует atomic для потокобезопасного счётчика.
// ✅ Подходит для простых операций: инкремент, декремент, загрузка/сохранение.
// 📖 Изучить: https://pkg.go.dev/sync/atomic
type SafeCounter struct {
	count int64
}

// Increment атомарно увеличивает счётчик.
// [5.1] AddInt64 гарантирует, что операция чтения-изменения-записи выполняется без прерываний.
// Нет блокировок → нет переключения контекста → быстрее, чем Mutex, для простых счётчиков.
func (c *SafeCounter) Increment() {
	atomic.AddInt64(&c.count, 1)
}

// Value атомарно читает значение.
// [5.1] LoadInt64 гарантирует, что мы читаем полностью записанное 64-битное значение.
func (c *SafeCounter) Value() int64 {
	return atomic.LoadInt64(&c.count)
}
