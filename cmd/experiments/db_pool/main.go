package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	dsn := "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Production-настройки пула
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(1 * time.Minute)

	// 🆕 Проверяем подключение при старте
	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		log.Fatalf("❌ Cannot connect to DB: %v", err)
	}
	fmt.Println("✅ Connected to PostgreSQL")

	printStats("After ping", db)

	// 🆕 Запускаем 5 горутин с РЕАЛЬНЫМИ запросами
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// РЕАЛЬНЫЙ запрос к БД!
			queryCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			// pg_sleep имитирует долгий запрос (1 секунда)
			_, err := db.ExecContext(queryCtx, "SELECT pg_sleep(1)")
			if err != nil {
				log.Printf("Goroutine %d error: %v", id, err)
				return
			}

			// Смотрим статистику ПРЯМО ВО ВРЕМЯ выполнения
			stats := db.Stats()
			fmt.Printf("Goroutine %d | Open: %d | InUse: %d | Idle: %d | Wait: %d\n",
				id, stats.OpenConnections, stats.InUse, stats.Idle, stats.WaitCount)
		}(i)
	}

	// Ждём завершения всех горутин
	wg.Wait()

	printStats("After all queries", db)

	// 🆕 Проверяем, что соединения вернулись в idle
	time.Sleep(100 * time.Millisecond)
	printStats("After idle timeout", db)
}

func printStats(label string, db *sql.DB) {
	stats := db.Stats()
	fmt.Printf("\n=== %s ===\n", label)
	fmt.Printf("OpenConnections: %d\n", stats.OpenConnections)
	fmt.Printf("InUse:           %d\n", stats.InUse)
	fmt.Printf("Idle:            %d\n", stats.Idle)
	fmt.Printf("WaitCount:       %d\n", stats.WaitCount)
	fmt.Printf("MaxIdleClosed:   %d\n", stats.MaxIdleClosed)
}
