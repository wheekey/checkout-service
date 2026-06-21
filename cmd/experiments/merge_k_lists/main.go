package main

import (
	"container/heap"
	"fmt"
)

// Связный список (стандартная структура LeetCode)
type ListNode struct {
	Val  int
	Next *ListNode
}

// Обёртка для узла, чтобы heap мог сравнивать
type HeapNode struct {
	node *ListNode
}

type MinHeap []*HeapNode

// 1. Len — размер heap
func (h MinHeap) Len() int {
	return len(h)
}

// 2. Less — сравнение (кто меньше, тот выше в куче)
// ⚠️ КЛЮЧЕВОЙ метод! Именно здесь мы говорим, что сравниваем по Val
func (h MinHeap) Less(i, j int) bool {
	return h[i].node.Val < h[j].node.Val
}

// 3. Swap — перестановка элементов
func (h MinHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

// 4. Push — добавление элемента
// ⚠️ Важный момент: принимает interface{}, нужно привести к типу
func (h *MinHeap) Push(x interface{}) {
	*h = append(*h, x.(*HeapNode))
}

// 5. Pop — удаление элемента (всегда последнего после перестройки)
func (h *MinHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]     // Берём последний элемент
	*h = old[0 : n-1] // Уменьшаем срез
	return x
}

func mergeKLists(lists []*ListNode) *ListNode {
	// 1. Создаём heap и инициализируем его
	h := &MinHeap{}
	heap.Init(h)

	// 2. Добавляем головы всех списков в heap
	for _, list := range lists {
		if list != nil { // ⚠️ Edge case: список может быть пустым
			heap.Push(h, &HeapNode{node: list})
		}
	}

	// 3. Фиктивный head — чтобы не обрабатывать первый элемент отдельно
	dummy := &ListNode{}
	current := dummy

	// 4. Основной цикл: пока heap не пуст
	for h.Len() > 0 {
		// 4.1. Достаём минимальный элемент
		min := heap.Pop(h).(*HeapNode) // ⚠️ Приводим тип обратно

		// 4.2. Добавляем его в результат
		current.Next = min.node
		current = current.Next

		// 4.3. Если у минимума есть следующий — кладём его в heap
		if min.node.Next != nil {
			heap.Push(h, &HeapNode{node: min.node.Next})
		}
	}

	// 5. Возвращаем результат (пропуская dummy)
	return dummy.Next
}

// createList — helper для создания списка из массива
func createList(vals []int) *ListNode {
	if len(vals) == 0 {
		return nil
	}
	head := &ListNode{Val: vals[0]}
	current := head
	for _, v := range vals[1:] {
		current.Next = &ListNode{Val: v}
		current = current.Next
	}
	return head
}

// printList — helper для вывода списка
func printList(head *ListNode) {
	for head != nil {
		fmt.Printf("%d ", head.Val)
		head = head.Next
	}
	fmt.Println()
}

func main() {
	// Тест 1: классический пример
	lists := []*ListNode{
		createList([]int{1, 4, 5}),
		createList([]int{1, 3, 4}),
		createList([]int{2, 6}),
	}
	fmt.Print("Test 1: ")
	printList(mergeKLists(lists)) // Ожидаем: 1 1 2 3 4 4 5 6

	// Тест 2: пустые списки
	lists2 := []*ListNode{nil, nil, nil}
	fmt.Print("Test 2: ")
	printList(mergeKLists(lists2)) // Ожидаем: (пусто)

	// Тест 3: один список
	lists3 := []*ListNode{createList([]int{1, 2, 3})}
	fmt.Print("Test 3: ")
	printList(mergeKLists(lists3)) // Ожидаем: 1 2 3

	// Тест 4: одинаковые значения
	lists4 := []*ListNode{
		createList([]int{1, 1, 1}),
		createList([]int{1, 1, 1}),
	}
	fmt.Print("Test 4: ")
	printList(mergeKLists(lists4)) // Ожидаем: 1 1 1 1 1 1
}
