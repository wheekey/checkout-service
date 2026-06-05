package counter_test

import (
	"sync"
	"testing"

	"checkout-service/internal/counter"

	"github.com/stretchr/testify/assert"
)

// [5.3] Тест на гонку данных. Запускай ТОЛЬКО с флагом -race:
// go test -race -v ./internal/counter/

func TestUnsafeCounter_Race(t *testing.T) {
	// [5.3] Этот тест НАМЕРЕННО содержит гонку. Он должен падать при -race.
	// В реальном проекте такой код не должен попадать в main.
	t.Skip("⚠️  Этот тест содержит намеренную гонку. Раскомментируй для демонстрации -race")

	var wg sync.WaitGroup
	c := &counter.UnsafeCounter{}

	// 100 горутин, каждая делает 1000 инкрементов
	// Ожидаемый результат: 100_000
	// Реальный результат с гонкой: < 100_000 (потерянные обновления)
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				c.Increment()
			}
		}()
	}
	wg.Wait()

	// Этот ассерт часто будет падать из-за гонки
	// -race детектор также выдаст предупреждение в stderr
	assert.Equal(t, int64(100_000), c.Value(), "Lost updates due to race condition")
}

func TestSafeCounter_NoRace(t *testing.T) {
	// [5.3] Этот тест БЕЗОПАСЕН. -race не должен ругаться.
	var wg sync.WaitGroup
	c := &counter.SafeCounter{}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				c.Increment()
			}
		}()
	}
	wg.Wait()

	// Атомарные операции гарантируют корректный результат
	assert.Equal(t, int64(100_000), c.Value())
}

func TestMutexCounter_NoRace(t *testing.T) {
	// [5.3] Mutex также безопасен, но медленнее из-за блокировок
	var wg sync.WaitGroup
	c := &counter.MutexCounter{}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 1000; j++ {
				c.Increment()
			}
		}()
	}
	wg.Wait()

	count, _ := c.Value()
	assert.Equal(t, int64(100_000), count)
}
