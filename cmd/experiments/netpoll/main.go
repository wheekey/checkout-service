package main

import (
	"fmt"
	"io"
	"net/http"
	"runtime"
	"sync"
	"time"
)

// Медленный обработчик: имитирует долгий сетевой ответ
func slowHandler(w http.ResponseWriter, r *http.Request) {
	// "Думаем" 500мс перед ответом
	time.Sleep(500 * time.Millisecond)
	w.Write([]byte("OK"))
}

func main() {
	// Оставляем только 1 P (1 логический процессор)
	runtime.GOMAXPROCS(1)

	fmt.Printf("GOMAXPROCS: %d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("NumGoroutine at start: %d\n", runtime.NumGoroutine())

	// Запускаем HTTP-сервер
	http.HandleFunc("/slow", slowHandler)
	go func() {
		http.ListenAndServe(":8080", nil)
	}()

	// Даём серверу время на старт
	time.Sleep(100 * time.Millisecond)

	// Отправляем 100 параллельных запросов
	const numRequests = 100
	var wg sync.WaitGroup
	start := time.Now()

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			resp, err := http.Get("http://localhost:8080/slow")
			if err != nil {
				fmt.Printf("Request %d failed: %v\n", id, err)
				return
			}
			defer resp.Body.Close()
			io.ReadAll(resp.Body)
		}(i)
	}

	// Ждём завершения всех запросов
	wg.Wait()
	duration := time.Since(start)

	fmt.Printf("\n✅ Результаты:\n")
	fmt.Printf("Запросов: %d\n", numRequests)
	fmt.Printf("Время: %v\n", duration)
	fmt.Printf("RPS (примерно): %.2f\n", float64(numRequests)/duration.Seconds())
	fmt.Printf("NumGoroutine at end: %d\n", runtime.NumGoroutine())
	fmt.Printf("\n💡 Если бы не было Netpoller, при GOMAXPROCS=1\n")
	fmt.Printf("   100 запросов по 500мс выполнялись бы 50 секунд.\n")
	fmt.Printf("   Но благодаря Netpoller, они выполняются параллельно!\n")
}
