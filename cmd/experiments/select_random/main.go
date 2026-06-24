package main

import (
	"fmt"
)

func main() {
	ch1 := make(chan int, 1)
	ch2 := make(chan int, 1)

	// Заполняем оба канала — они оба готовы
	ch1 <- 1
	ch2 <- 2

	// Считаем, сколько раз выберется каждый канал
	counts := make(map[string]int)

	// Запускаем 1000 итераций
	for i := 0; i < 1000; i++ {
		select {
		case <-ch1:
			counts["ch1"]++
			ch1 <- 1 // возвращаем значение обратно
		case <-ch2:
			counts["ch2"]++
			ch2 <- 2 // возвращаем значение обратно
		}
	}

	fmt.Printf("ch1 выбран: %d раз\n", counts["ch1"])
	fmt.Printf("ch2 выбран: %d раз\n", counts["ch2"])
	fmt.Printf("\n")

	// Проверяем, примерно ли равно распределение
	ratio := float64(counts["ch1"]) / float64(counts["ch2"])
	fmt.Printf("Соотношение ch1/ch2: %.2f\n", ratio)

	if ratio > 0.8 && ratio < 1.2 {
		fmt.Println("✅ Подтверждено: select выбирает случайный case!")
	} else {
		fmt.Println("❌ Что-то не так с экспериментом")
	}
}
