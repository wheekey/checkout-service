package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"
)

// ✅ Исправленная версия — нет утечек
func fixedWorker() {
	ch := make(chan int, 1) // Буферизованный канал
	go func() {
		ch <- 42 // Пишем в канал
	}()
	<-ch // Читаем результат
}

func main() {
	go func() {
		fmt.Println("pprof available at http://localhost:6060/debug/pprof/")
		http.ListenAndServe("localhost:6060", nil)
	}()

	fmt.Printf("Начало: %d горутин\n", runtime.NumGoroutine())

	// Создаём 10 "правильных" воркеров
	for i := 0; i < 10; i++ {
		fixedWorker()
	}

	fmt.Printf("После fixed workers: %d горутин\n", runtime.NumGoroutine())
	// Ожидаем: то же число (нет утечек)

	time.Sleep(100 * time.Millisecond)

	fmt.Println("\nСервер работает. Нажми Ctrl+C для выхода.")
	fmt.Println("Открой: http://localhost:6060/debug/pprof/goroutine?debug=1")
	select {}
}
