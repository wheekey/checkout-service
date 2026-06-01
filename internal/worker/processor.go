package worker

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"golang.org/x/sync/errgroup" // [5.1] Стандартный паттерн для конкурентных задач с общим контекстом и пробросом ошибок
)

// [5.4] Единица работы. Передаётся по значению, чтобы избежать race conditions в замыкании цикла.
type Task struct {
	ID     string
	Amount int
}

// [5.1][5.5] sync.Pool: хранилище временных объектов.
// 🔍 Internals:
//   - Объекты хранятся per-P (логический процессор) → доступ без блокировок.
//   - Автоматически очищается при каждом GC → НИКОГДА не храните здесь состояние.
//   - Снижает allocs/op → реже запускается GC → стабильнее p95 latency.
//
// 📖 Изучить:
//   - https://go.dev/blog/pools
//   - https://go.dev/doc/gc-guide
var resultPool = sync.Pool{
	New: func() any {
		// Преаллоцируем capacity, чтобы append не вызывал реаллокацию буфера
		b := make([]byte, 0, 256)
		return &b
	},
}

// ProcessBatch обрабатывает пачку задач параллельно, контролируя параллелизм и отменяя работу при ошибке.
// [5.1] context.Context: явная передача отмены/дедлайнов. Не храним в структурах.
// [5.4] Worker Pool: защищает БД, внешние API и память от истощения при всплесках трафика.
func ProcessBatch(ctx context.Context, tasks []Task) error {
	// [5.1] errgroup.WithContext создаёт дочерний контекст.
	// Если ЛЮБОЙ g.Go() вернёт ошибку → контекст автоматически отменится.
	// Все остальные горутины, проверяющие ctx.Done(), завершатся gracefully.
	g, ctx := errgroup.WithContext(ctx)

	// [5.1][5.5] SetLimit работает как семафор (Go 1.21+).
	// Без лимита: 10K задач → 10K горутин → троттлинг CPU, исчерпание FD, падение пула соединений БД.
	g.SetLimit(5)

	for _, t := range tasks {
		t := t // [5.1] Захват переменной цикла. В Go <1.22 обязательно, в 1.22+ опционально, но оставляем для совместимости.

		g.Go(func() error {
			// [5.1] Fast-fail: если другой воркер уже упал, контекст отменён.
			// Пропускаем работу, экономим CPU и не долбим БД бесполезными запросами.
			if ctx.Err() != nil {
				return ctx.Err()
			}

			// Симуляция бизнес-валидации / внешнего вызова
			if t.Amount > 10_000 {
				// [5.4] Возврат ошибки → errgroup отменяет контекст → остальные горутины выходят.
				return fmt.Errorf("task %s: amount exceeds limit", t.ID)
			}

			// [5.1] Использование sync.Pool в hot-path:
			// 1. Get() → берём буфер из пула (типизация безопасна, контролируется pool.New)
			// 2. Reset длины → сохраняем capacity, очищаем данные
			// 3. Always Put() обратно → иначе пул истощится, GC будет аллоцировать заново
			// 📖 Изучить: https://go.dev/doc/effective_go#sync_pool
			buf := resultPool.Get().(*[]byte)
			*buf = (*buf)[:0]
			*buf = append(*buf, "proc-"...)
			*buf = append(*buf, t.ID...)

			slog.Info("Task processed", "result", string(*buf))

			// Возвращаем в пул для переиспользования другими горутинами
			resultPool.Put(buf)

			return nil
		})
	}

	// [5.1] Блокирует выполнение, пока все горутины не завершатся или контекст не будет отменён.
	// Возвращает ПЕРВУЮ возникшую ошибку или nil, если все задачи успешны.
	return g.Wait()
}
