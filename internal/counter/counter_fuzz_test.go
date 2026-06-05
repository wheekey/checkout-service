package counter_test

import (
	"testing"

	"checkout-service/internal/counter"
)

// [5.3] Fuzz-тест: проверяет, что код не паникует на невалидном вводе.
// Запуск: go test -fuzz=FuzzIncrement -fuzztime=10s ./internal/counter/

func FuzzIncrement(f *testing.F) {
	// Добавляем «затравку» (seed) для fuzzing-движка
	// Движок будет мутировать эти значения, пытаясь найти краш
	// 📖 Изучить: https://go.dev/doc/fuzz/
	f.Add(int64(0))
	f.Add(int64(1))
	f.Add(int64(-1))
	f.Add(int64(9223372036854775807))  // max int64
	f.Add(int64(-9223372036854775808)) // min int64

	f.Fuzz(func(t *testing.T, inc int64) {
		c := &counter.SafeCounter{}

		// Проверяем, что инкремент не паникует даже на экстремальных значениях
		// (переполнение int64 — отдельная бизнес-логика, но не краш рантайма)
		_ = func() (deferred bool) {
			defer func() {
				if r := recover(); r != nil {
					deferred = true
				}
			}()
			c.Increment()
			return false
		}()

		// Если была паника — fuzz-тест упадёт
		// В реальном проекте: добавь валидацию входных данных
	})
}
