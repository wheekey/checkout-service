package worker

import (
	"checkout-service/internal/metrics"
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
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

func (p *Pool) worker(ctx context.Context, id int) {
	defer p.wg.Done()

	// Увеличиваем счётчик активных воркеров
	metrics.ActiveWorkers.Inc()
	defer metrics.ActiveWorkers.Dec()

	for {
		select {
		case <-ctx.Done():
			slog.Info("Worker stopping", "worker_id", id)
			return
		case task, ok := <-p.tasks:
			if !ok {
				// Канал закрыт — дорабатываем оставшиеся задачи в буфере
				slog.Info("Tasks channel closed, draining remaining tasks", "worker_id", id)
				p.drainRemainingTasks(id)
				return
			}
			result := p.processTask(task)
			p.results <- result
		}
	}
}

// drainRemainingTasks дорабатывает все оставшиеся задачи в буфере
func (p *Pool) drainRemainingTasks(workerID int) {
	for {
		select {
		case task, ok := <-p.tasks:
			if !ok {
				// Буфер пуст
				slog.Info("No more tasks to drain", "worker_id", workerID)
				return
			}
			slog.Debug("Draining task", "task_id", task.ID, "worker_id", workerID)
			result := p.processTask(task)
			p.results <- result
		default:
			// Буфер пуст на данный момент
			slog.Info("Buffer drained", "worker_id", workerID)
			return
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

// SubmitWithTimeout отправляет задачу с таймаутом
// [Concurrency] Если буфер полон, ждём максимум timeout
func (p *Pool) SubmitWithTimeout(task Task, timeout time.Duration) error {
	select {
	case p.tasks <- task:
		return nil // успешно отправили
	case <-time.After(timeout):
		return fmt.Errorf("timeout: could not submit task %s", task.ID)
	}
}

// StopGraceful корректно останавливает пул без потерь задач.
// [Concurrency] 3-шаговый graceful shutdown:
// 1. close(tasks) — воркеры перестают брать НОВЫЕ задачи
// 2. wg.Wait() — ждём, пока воркеры доработают ТЕКУЩИЕ задачи
// 3. close(results) — закрываем канал результатов
func (p *Pool) StopGraceful() {
	p.stopOnce.Do(func() {
		slog.Info("Graceful shutdown started")

		// 1. Закрываем канал задач
		close(p.tasks)
		slog.Info("Tasks channel closed, workers will finish current tasks")

		// 2. Ждём завершения всех воркеров
		p.wg.Wait()
		slog.Info("All workers finished")

		// 3. Закрываем канал результатов
		close(p.results)
		slog.Info("Graceful shutdown completed")
	})
}

// StopWithTimeout останавливает пул с таймаутом.
// Если воркеры не успели завершиться за timeout — возвращает ошибку.
func (p *Pool) StopWithTimeout(timeout time.Duration) error {
	done := make(chan struct{})

	go func() {
		p.StopGraceful()
		close(done)
	}()

	select {
	case <-done:
		return nil // Успешно завершили
	case <-time.After(timeout):
		return fmt.Errorf("shutdown timeout after %v", timeout)
	}
}
