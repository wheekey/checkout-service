package worker_test

import (
	"testing"
	"time"

	"checkout-service/internal/worker"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPool_SubmitWithTimeout(t *testing.T) {
	// Создаём пул с буфером 2, но НЕ запускаем воркеров
	pool := worker.NewPool(1, 2)

	// НЕ вызываем pool.Start(ctx) — воркеров нет!

	// Заполняем буфер (2 задачи)
	pool.Submit(worker.Task{ID: "1", Amount: 100})
	pool.Submit(worker.Task{ID: "2", Amount: 200})

	// Третья задача должна упасть по таймауту, потому что:
	// - Буфер полон (2 задачи)
	// - Воркеров нет (никто не читает из tasks)
	// - Отправка заблокируется
	err := pool.SubmitWithTimeout(
		worker.Task{ID: "3", Amount: 300},
		50*time.Millisecond,
	)

	// Проверяем ошибку безопасно
	require.Error(t, err, "Ожидалась ошибка таймаута")
	assert.Contains(t, err.Error(), "timeout")
}

func TestPool_SubmitWithTimeout_Success(t *testing.T) {
	pool := worker.NewPool(1, 2)
	// НЕ запускаем воркеров

	// Буфер пуст — задача должна успешно отправиться
	err := pool.SubmitWithTimeout(
		worker.Task{ID: "1", Amount: 100},
		50*time.Millisecond,
	)

	require.NoError(t, err)
}
