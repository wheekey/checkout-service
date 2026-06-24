package main

import "fmt"

func maxSlidingWindow(nums []int, k int) []int {
	if len(nums) == 0 || k == 0 {
		return []int{}
	}

	// Дек хранит ИНДЕКСЫ в убывающем порядке значений
	// Голова дека (deque[0]) — индекс текущего максимума
	deque := []int{}
	result := []int{}

	for i, num := range nums {
		// 1. Удаляем индексы, которые вышли из окна
		// Окно: [i-k+1 ... i]
		for len(deque) > 0 && deque[0] < i-k+1 {
			deque = deque[1:]
		}

		// 2. Удаляем индексы элементов, которые МЕНЬШЕ текущего
		// Они никогда не станут максимумом
		for len(deque) > 0 && nums[deque[len(deque)-1]] < num {
			deque = deque[:len(deque)-1]
		}

		// 3. Добавляем текущий индекс в конец дека
		deque = append(deque, i)

		// 4. Если окно сформировано — добавляем максимум (голова дека)
		if i >= k-1 {
			result = append(result, nums[deque[0]])
		}
	}

	return result
}

func main() {
	// Тест 1: классический пример
	nums1 := []int{1, 3, -1, -3, 5, 3, 6, 7}
	fmt.Printf("Test 1:\n")
	fmt.Printf("  Input:  nums=%v, k=3\n", nums1)
	fmt.Printf("  Output: %v\n", maxSlidingWindow(nums1, 3))
	fmt.Printf("  Ожидаем: [3 3 5 5 6 7]\n\n")

	// Тест 2: все одинаковые
	nums2 := []int{1, 1, 1, 1, 1}
	fmt.Printf("Test 2:\n")
	fmt.Printf("  Input:  nums=%v, k=2\n", nums2)
	fmt.Printf("  Output: %v\n", maxSlidingWindow(nums2, 2))
	fmt.Printf("  Ожидаем: [1 1 1 1]\n\n")

	// Тест 3: убывающий массив
	nums3 := []int{5, 4, 3, 2, 1}
	fmt.Printf("Test 3:\n")
	fmt.Printf("  Input:  nums=%v, k=3\n", nums3)
	fmt.Printf("  Output: %v\n", maxSlidingWindow(nums3, 3))
	fmt.Printf("  Ожидаем: [5 4 3]\n\n")

	// Тест 4: возрастающий массив
	nums4 := []int{1, 2, 3, 4, 5}
	fmt.Printf("Test 4:\n")
	fmt.Printf("  Input:  nums=%v, k=3\n", nums4)
	fmt.Printf("  Output: %v\n", maxSlidingWindow(nums4, 3))
	fmt.Printf("  Ожидаем: [3 4 5]\n\n")

	// Тест 5: k = 1 (каждый элемент — окно)
	nums5 := []int{1, 3, -1, -3, 5}
	fmt.Printf("Test 5:\n")
	fmt.Printf("  Input:  nums=%v, k=1\n", nums5)
	fmt.Printf("  Output: %v\n", maxSlidingWindow(nums5, 1))
	fmt.Printf("  Ожидаем: [1 3 -1 -3 5]\n")
}
