package main

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"
)

// ❌ Функция с утечкой
func leakyWorker() {
	ch := make(chan int)
	go func() {
		// Эта горутина ждёт вечно — никто не пишет в ch
		<-ch
	}()
	// Функция завершилась, но горутина осталась
}

// ✅ Функция без утечки
func goodWorker() {
	ch := make(chan int, 1)
	go func() {
		ch <- 42
	}()
	<-ch // Читаем результат
}

func main() {
	// pprof endpoint
	go func() {
		fmt.Println("pprof available at http://localhost:6060/debug/pprof/")
		http.ListenAndServe("localhost:6060", nil)
	}()

	fmt.Printf("Начало: %d горутин\n", runtime.NumGoroutine())

	// Создаём 10 утечек
	for i := 0; i < 10; i++ {
		leakyWorker()
	}

	fmt.Printf("После утечек: %d горутин\n", runtime.NumGoroutine())
	// Ожидаем: +10 горутин (каждая leakyWorker создала 1 горутину)

	// Создаём 10 правильных воркеров
	for i := 0; i < 10; i++ {
		goodWorker()
	}

	fmt.Printf("После good workers: %d горутин\n", runtime.NumGoroutine())
	// Ожидаем: то же число (goodWorker не создаёт утечек)

	// Даём время pprof endpoint подняться
	time.Sleep(100 * time.Millisecond)

	// Выводим стек всех горутин
	fmt.Println("\n=== Стек всех горутин ===")
	buf := make([]byte, 1<<20)
	n := runtime.Stack(buf, true)
	fmt.Println(string(buf[:n]))

	// Держим программу живой для pprof
	fmt.Println("\nСервер работает. Нажми Ctrl+C для выхода.")
	fmt.Println("Открой: http://localhost:6060/debug/pprof/goroutine?debug=1")
	select {}
}
