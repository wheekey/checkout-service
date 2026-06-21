package worker

import (
	"context"
	"log/slog"
	"sync"
)

// Pool — воркер-пул с паттерном Fan-out/Fan-in
// [Concurrency] Fan-out: одна задача распределяется по N воркерам
// [Concurrency] Fan-in: результаты от N воркеров собираются в один канал
type Pool struct {
	workers  int
	tasks    chan Task
	results  chan Result
	wg       sync.WaitGroup
	stopOnce sync.Once // 🆕 Защита от двойного вызова
}

// Result представляет результат обработки
// [Concurrency] Передаётся по значению, чтобы избежать race conditions
type Result struct {
	TaskID string
	Output string
	Err    error
}

// NewPool создаёт новый воркер-пул
func NewPool(workers int, taskBufferSize int) *Pool {
	return &Pool{
		workers: workers,
		tasks:   make(chan Task, taskBufferSize),
		results: make(chan Result, taskBufferSize),
	}
}

// Start запускает воркеры
func (p *Pool) Start(ctx context.Context) {
	for i := 0; i < p.workers; i++ {
		p.wg.Add(1)
		go p.worker(ctx, i)
	}
	slog.Info("Worker pool started", "workers", p.workers)
}

// worker — отдельный воркер
func (p *Pool) worker(ctx context.Context, id int) {
	defer p.wg.Done()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Worker stopping", "worker_id", id)
			return
		case task, ok := <-p.tasks:
			if !ok {
				slog.Info("Tasks channel closed, worker exiting", "worker_id", id)
				return
			}
			result := p.processTask(task)
			p.results <- result
		}
	}
}

// processTask — логика обработки задачи
// [Архитектура] Здесь можно вызвать реальную бизнес-логику
func (p *Pool) processTask(task Task) Result {
	slog.Debug("Processing task", "task_id", task.ID, "amount", task.Amount)

	// Имитация работы (например, отправка колбэка, обновление статуса)
	return Result{
		TaskID: task.ID,
		Output: "processed",
	}
}

// Submit отправляет задачу в пул
func (p *Pool) Submit(task Task) {
	p.tasks <- task
}

// Results возвращает канал результатов (Fan-in)
func (p *Pool) Results() <-chan Result {
	return p.results
}

// Stop корректно останавливает пул.
// [Concurrency] sync.Once гарантирует, что каналы закроются только ОДИН раз,
// даже если Stop() вызывается из нескольких мест (например, из server.Run и main).
func (p *Pool) Stop() {
	p.stopOnce.Do(func() {
		close(p.tasks)
		p.wg.Wait()
		close(p.results)
		slog.Info("Worker pool stopped")
	})
}
