package main

import "fmt"

// ListNode — структура узла связного списка
type ListNode struct {
	Val  int
	Next *ListNode
}

// hasCycle определяет, есть ли цикл в связном списке
// Сложность: O(N) по времени, O(1) по памяти
func hasCycle(head *ListNode) bool {
	if head == nil || head.Next == nil {
		return false
	}

	slow := head
	fast := head

	for fast != nil && fast.Next != nil {
		slow = slow.Next      // 1 шаг
		fast = fast.Next.Next // 2 шага

		if slow == fast {
			return true // Встретились → цикл есть
		}
	}

	return false // fast дошёл до null → цикла нет
}

// createListWithCycle создаёт список с циклом для теста
func createListWithCycle(vals []int, pos int) *ListNode {
	if len(vals) == 0 {
		return nil
	}

	head := &ListNode{Val: vals[0]}
	current := head
	var cycleNode *ListNode

	for i := 1; i < len(vals); i++ {
		current.Next = &ListNode{Val: vals[i]}
		current = current.Next
		if i == pos {
			cycleNode = current
		}
	}

	// Создаём цикл
	if cycleNode != nil {
		current.Next = cycleNode
	}

	return head
}

func main() {
	// Тест 1: цикл есть (pos=1, цикл к узлу со значением 2)
	list1 := createListWithCycle([]int{3, 2, 0, -4}, 1)
	fmt.Printf("Test 1: [3,2,0,-4], pos=1 → %v (ожидаем true)\n", hasCycle(list1))

	// Тест 2: цикл есть (pos=0, цикл к голове)
	list2 := createListWithCycle([]int{1, 2}, 0)
	fmt.Printf("Test 2: [1,2], pos=0 → %v (ожидаем true)\n", hasCycle(list2))

	// Тест 3: цикла нет (pos=-1)
	list3 := createListWithCycle([]int{1}, -1)
	fmt.Printf("Test 3: [1], pos=-1 → %v (ожидаем false)\n", hasCycle(list3))

	// Тест 4: пустой список
	fmt.Printf("Test 4: [] → %v (ожидаем false)\n", hasCycle(nil))
}
