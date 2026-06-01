package worker_test

import (
	"context"
	"testing"
	"time"

	"checkout-service/internal/worker"
	"github.com/stretchr/testify/assert" // [5.3] Стандарт де-факто для чистых ассертов в Go-экосистеме
)

// [5.3] Тестирование в отдельном пакете (_test):
//   • Проверяем только публичный API (контракт интерфейса)
//   • Исключаем циклические зависимости между internal-пакетами
//   • Имитируем поведение внешнего потребителя библиотеки
// 📖 Изучить: https://go.dev/doc/tutorial/add-a-test

// TestProcessBatch проверяет корректность параллельной обработки пачки задач.
// [5.1][5.3] Table-driven tests — идиома Go для компактного и масштабируемого покрытия.
// [5.5] Покрываем: happy path, error propagation, graceful cancellation, edge cases.
func TestProcessBatch(t *testing.T) {
	// [5.3] Запускаем тесты параллельно. Ускоряет CI/CD пайплайн на 30–60%.
	// Требует, чтобы код под тестом был потокобезопасным (проверяем через -race).
	// 📖 Изучить: https://pkg.go.dev/testing#T.Parallel
	t.Parallel()

	// [5.3] Структура кейсов: входные данные + ожидаемый результат.
	// Легко добавить новый сценарий без копипасты и усложнения логики.
	tests := []struct {
		name        string
		tasks       []worker.Task
		expectError bool
		errContains string // Подстрока ошибки для частичного совпадения (без brittle-проверок)
	}{
		{
			name: "success: all tasks processed",
			tasks: []worker.Task{
				{ID: "t1", Amount: 100},
				{ID: "t2", Amount: 200},
				{ID: "t3", Amount: 300},
			},
			expectError: false,
		},
		{
			// [5.1][5.4] Проверка отмены контекста: errgroup должен прервать выполнение
			// после первой ошибки и не запускать оставшиеся горутины.
			name: "error propagation & context cancellation",
			tasks: []worker.Task{
				{ID: "t1", Amount: 100},
				{ID: "t_fail", Amount: 15_000}, // Превышает лимит → trigger error
				{ID: "t3", Amount: 50},         // Должен быть отменён или пропущен
			},
			expectError: true,
			errContains: "amount exceeds limit",
		},
		{
			// [5.3] Edge case: пустой слайс. Функция должна вернуть nil без паники/блокировки.
			name:        "edge case: empty tasks slice",
			tasks:       []worker.Task{},
			expectError: false,
		},
	}

	// [5.3] Цикл по таблице. Каждый кейс изолирован.
	for _, tt := range tests {
		// t.Run создаёт подтест. t.Parallel() внутри позволяет запускать их конкурентно.
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// [5.1] Контекст с таймаутом КРИТИЧНО для асинхронных тестов.
			// Без него deadlock или вечная блокировка положат CI/CD.
			// defer cancel() гарантирует освобождение ресурсов даже при панике.
			// 📖 Изучить: https://go.dev/blog/context
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			// Вызов тестируемой функции
			err := worker.ProcessBatch(ctx, tt.tasks)

			// [5.3] Ассерты через testify: читаемо, дают детальные сообщения при падении.
			// В Go 1.21+ также доступен testing.T.Fatalf для встроенных проверок,
			// но testify остаётся стандартом для интеграционных/юнит-тестов.
			// 📖 Изучить: https://github.com/stretchr/testify
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
