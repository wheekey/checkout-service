package counter

import "sync" // [5.1] Примитивы синхронизации

// MutexCounter использует Mutex для защиты сложных инвариантов.
// ✅ Подходит, когда нужно атомарно изменить НЕСКОЛЬКО полей или выполнить сложную логику.
// 📖 Изучить: https://go.dev/blog/mutex
type MutexCounter struct {
	mu      sync.Mutex
	count   int64
	history []int64 // [5.1] Пример: нужно атомарно обновить счётчик + добавить в историю
}

// Increment атомарно увеличивает счётчик и сохраняет историю.
// [5.1] Mutex блокирует другие горутины на время выполнения критической секции.
// Это медленнее, чем atomic, но безопасно для сложных операций.
func (c *MutexCounter) Increment() {
	c.mu.Lock()
	defer c.mu.Unlock() // [5.1] defer гарантирует Unlock даже при панике

	c.count++
	c.history = append(c.history, c.count) // [5.1] Безопасно: внутри критической секции
}

// Value возвращает копию счётчика и истории.
func (c *MutexCounter) Value() (int64, []int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Возвращаем копию слайса, чтобы избежать утечки мутации наружу
	historyCopy := make([]int64, len(c.history))
	copy(historyCopy, c.history)
	return c.count, historyCopy
}
