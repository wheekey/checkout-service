package main

import (
	"fmt"
	"runtime"
	"time"
)

func main() {
	// Оставляем только 1 логический процессор
	runtime.GOMAXPROCS(1)

	fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("NumCPU: %d\n", runtime.NumCPU())

	// Горутина 1: Бесконечный цикл
	go func() {
		for {
			// Пустой цикл. В Go >= 1.14 sysmon прервет это через ~10мс.
		}
	}()

	// Горутина 2: Полезная задача
	go func() {
		time.Sleep(20 * time.Millisecond)
		fmt.Println("✅ Успех! Sysmon прервал бесконечный цикл и дал нам поработать.")
	}()

	// Ждем завершения
	time.Sleep(100 * time.Millisecond)
}
