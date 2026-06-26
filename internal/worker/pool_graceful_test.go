package worker_test

import (
	"checkout-service/internal/worker"
	"context"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPool_GracefulShutdown(t *testing.T) {
	pool := worker.NewPool(2, 10) // 2 воркера, буфер 10
	ctx := context.Background()
	pool.Start(ctx)

	// Считаем обработанные задачи
	var processed int32

	// Читаем результаты
	go func() {
		for range pool.Results() {
			atomic.AddInt32(&processed, 1)
		}
	}()

	// Отправляем 5 задач
	for i := 0; i < 5; i++ {
		pool.Submit(worker.Task{ID: string(rune('A' + i)), Amount: 100})
	}

	// Graceful shutdown
	pool.StopGraceful()

	// Все 5 задач должны быть обработаны
	assert.Equal(t, int32(5), atomic.LoadInt32(&processed),
		"Все задачи должны быть обработаны при graceful shutdown")
}

func TestPool_GracefulShutdown_WithSlowTasks(t *testing.T) {
	pool := worker.NewPool(1, 5)
	ctx := context.Background()
	pool.Start(ctx)

	var processed int32

	go func() {
		for range pool.Results() {
			atomic.AddInt32(&processed, 1)
		}
	}()

	// Отправляем 3 задачи
	for i := 0; i < 3; i++ {
		pool.Submit(worker.Task{ID: string(rune('A' + i)), Amount: 100})
	}

	// Graceful shutdown должен дождаться обработки всех задач
	pool.StopGraceful()

	assert.Equal(t, int32(3), atomic.LoadInt32(&processed))
}
