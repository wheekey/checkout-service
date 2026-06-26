package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// or принимает каналы struct{} — как ctx.Done()
func or(channels ...<-chan struct{}) <-chan struct{} {
	switch len(channels) {
	case 0:
		return nil
	case 1:
		return channels[0]
	}

	orDone := make(chan struct{})
	go func() {
		defer close(orDone)
		switch len(channels) {
		case 2:
			select {
			case <-channels[0]:
			case <-channels[1]:
			}
		default:
			mid := len(channels) / 2
			select {
			case <-or(channels[:mid]...):
			case <-or(channels[mid:]...):
			}
		}
	}()
	return orDone
}

func main() {
	// 1. Контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 2. Канал для сигналов ОС (struct{}, как ctx.Done())
	sigChan := make(chan struct{})
	go func() {
		osSignal := make(chan os.Signal, 1)
		signal.Notify(osSignal, syscall.SIGINT, syscall.SIGTERM)
		<-osSignal
		close(sigChan)
	}()

	// 3. Ручная отмена (struct{})
	manualCancel := make(chan struct{})
	go func() {
		time.Sleep(2 * time.Second)
		fmt.Println("Ручная отмена через 2 секунды")
		close(manualCancel)
	}()

	fmt.Println("Запущен. Жду отмены от любого из 3 источников:")
	fmt.Println("  1. context.WithTimeout (5 сек)")
	fmt.Println("  2. Ctrl+C (SIGINT)")
	fmt.Println("  3. Ручная отмена (2 сек)")

	// ✅ Теперь работает! Все каналы <-chan struct{}
	<-or(ctx.Done(), sigChan, manualCancel)

	fmt.Println("Получен сигнал отмены — завершаю работу")
}
