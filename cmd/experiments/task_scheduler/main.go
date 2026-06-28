package main

import (
	"fmt"
)

// leastInterval находит минимальное время выполнения всех задач
// Сложность: O(N) по времени, O(1) по памяти (26 букв)
func leastInterval(tasks []byte, n int) int {
	// Шаг 1: Считаем частоты каждой задачи
	freq := make(map[byte]int)
	maxCount := 0

	for _, task := range tasks {
		freq[task]++
		if freq[task] > maxCount {
			maxCount = freq[task]
		}
	}

	// Шаг 2: Считаем, сколько задач имеют максимальную частоту
	numMaxTasks := 0
	for _, count := range freq {
		if count == maxCount {
			numMaxTasks++
		}
	}

	// Шаг 3: Применяем формулу
	// Вариант 1: (max_count - 1) * (n + 1) + num_max_tasks
	// Вариант 2: len(tasks) (если задач много, кулдаун не важен)

	formula := (maxCount-1)*(n+1) + numMaxTasks

	if formula > len(tasks) {
		return formula
	}
	return len(tasks)
}

func main() {
	// Тест 1: классический пример
	tasks1 := []byte{'A', 'A', 'A', 'B', 'B', 'B'}
	fmt.Printf("Test 1: tasks=%c, n=2 → %d (ожидаем 8)\n", tasks1, leastInterval(tasks1, 2))

	// Тест 2: кулдаун 0
	tasks2 := []byte{'A', 'A', 'A', 'B', 'B', 'B'}
	fmt.Printf("Test 2: tasks=%c, n=0 → %d (ожидаем 6)\n", tasks2, leastInterval(tasks2, 0))

	// Тест 3: много разных задач
	tasks3 := []byte{'A', 'A', 'A', 'A', 'A', 'A', 'B', 'C', 'D', 'E', 'F', 'G'}
	fmt.Printf("Test 3: tasks=%c, n=2 → %d (ожидаем 16)\n", tasks3, leastInterval(tasks3, 2))

	// Тест 4: все задачи одинаковые
	tasks4 := []byte{'A', 'A', 'A', 'A'}
	fmt.Printf("Test 4: tasks=%c, n=2 → %d (ожидаем 10)\n", tasks4, leastInterval(tasks4, 2))

	// Тест 5: все задачи разные
	tasks5 := []byte{'A', 'B', 'C', 'D'}
	fmt.Printf("Test 5: tasks=%c, n=2 → %d (ожидаем 4)\n", tasks5, leastInterval(tasks5, 2))
}
