package main

import (
	"runtime"
	"testing"
	"time"
)

func TestSysmonPreemption(t *testing.T) {
	runtime.GOMAXPROCS(1)

	done := make(chan bool, 1)

	// Горутина 1: Бесконечный цикл
	go func() {
		for {
			// Пустой цикл
		}
	}()

	// Горутина 2: Полезная задача
	go func() {
		time.Sleep(20 * time.Millisecond)
		done <- true
	}()

	// Ждем сигнал от второй горутины
	select {
	case <-done:
		t.Log("✅ Успех! Sysmon прервал бесконечный цикл.")
	case <-time.After(1 * time.Second):
		t.Fatal("❌ Провал: sysmon не сработал, горутина зависла.")
	}
}
