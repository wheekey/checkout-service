package main

import (
	"context"
	"fmt"
	"time"
)

// worker имитирует работу с проверкой контекста
func worker(ctx context.Context, name string, delay time.Duration) {
	fmt.Printf("[%s] запущен\n", name)
	defer fmt.Printf("[%s] з	авершён\n", name)

	for {
		select {
		case <-ctx.Done():
			fmt.Printf("[%s] получил отмену: %v\n", name, ctx.Err())
			return
		case <-time.After(delay):
			fmt.Printf("[%s] выполнил работу\n", name)
		}
	}
}

func main() {
	// Корневой контекст
	rootCtx, rootCancel := context.WithCancel(context.Background())

	// Первый уровень: два дочерних контекста
	child1Ctx, child1Cancel := context.WithCancel(rootCtx)
	defer child1Cancel() // На случай, если забудем отменить

	child2Ctx, child2Cancel := context.WithCancel(rootCtx)
	defer child2Cancel()

	// Второй уровень: внук от child1
	grandchildCtx, grandchildCancel := context.WithCancel(child1Ctx)
	defer grandchildCancel()

	// Запускаем горутины
	go worker(grandchildCtx, "grandchild", 300*time.Millisecond)
	go worker(child1Ctx, "child1", 500*time.Millisecond)
	go worker(child2Ctx, "child2", 700*time.Millisecond)

	// Даём поработать
	time.Sleep(1500 * time.Millisecond)

	// Отменяем КОРНЕВОЙ контекст → ВСЕ дочерние отменятся автоматически
	fmt.Println("\n=== Отменяем rootCtx ===\n")
	rootCancel()

	// Ждём завершения
	time.Sleep(500 * time.Millisecond)
}
