package main

import (
	"container/list"
	"fmt"
)

// entry — элемент кэша (хранит ключ для быстрого удаления из map)
type entry struct {
	key   int
	value int
}

// LRUCache — Least Recently Used Cache
// Сложность: O(1) на Get и Put
//
// Структура: HashMap + Doubly Linked List
// - HashMap: key → *list.Element (быстрый поиск O(1))
// - List: двусвязный список (быстрое перемещение O(1))
//
// Конвенция (классическая):
//
//	head ←→ [MRU] ←→ ... ←→ [LRU] ←→ tail
//	         ↑                   ↑
//	     head.next           tail.prev
//	     (добавляем)         (удаляем)
type LRUCache struct {
	capacity int
	items    map[int]*list.Element // key → node
	order    *list.List            // front = MRU, back = LRU
}

// NewLRUCache создаёт новый LRU кэш
func NewLRUCache(capacity int) *LRUCache {
	if capacity <= 0 {
		panic("capacity must be positive")
	}
	return &LRUCache{
		capacity: capacity,
		items:    make(map[int]*list.Element),
		order:    list.New(),
	}
}

// Get возвращает значение по ключу
// Если ключ найден — перемещает его в MRU (в начало)
// Сложность: O(1)
func (c *LRUCache) Get(key int) (int, bool) {
	elem, exists := c.items[key]
	if !exists {
		return -1, false // Cache miss
	}

	// Перемещаем в начало (MRU)
	c.order.MoveToFront(elem)
	return elem.Value.(*entry).value, true
}

// Put добавляет или обновляет ключ-значение
// Если кэш переполнен — удаляет LRU (из конца)
// Сложность: O(1)
func (c *LRUCache) Put(key int, value int) {
	// Случай 1: ключ уже есть → обновляем и перемещаем в MRU
	if elem, exists := c.items[key]; exists {
		c.order.MoveToFront(elem)
		elem.Value.(*entry).value = value
		return
	}

	// Случай 2: новый ключ → добавляем в MRU
	elem := c.order.PushFront(&entry{key: key, value: value})
	c.items[key] = elem

	// Случай 3: переполнение → удаляем LRU (из конца)
	if c.order.Len() > c.capacity {
		back := c.order.Back()
		if back != nil {
			c.order.Remove(back)
			delete(c.items, back.Value.(*entry).key)
		}
	}
}

// Len возвращает текущее количество элементов
func (c *LRUCache) Len() int {
	return c.order.Len()
}

func main() {
	fmt.Println("=== Пример из LeetCode ===")
	cache := NewLRUCache(2)

	cache.Put(1, 1)
	fmt.Printf("После Put(1,1):    %s\n", cache)

	cache.Put(2, 2)
	fmt.Printf("После Put(2,2):    %s\n", cache)

	val, _ := cache.Get(1)
	fmt.Printf("Get(1) = %d         %s\n", val, cache)

	cache.Put(3, 3) // evicts key 2
	fmt.Printf("После Put(3,3):    %s (key 2 evicted)\n", cache)

	val, ok := cache.Get(2)
	fmt.Printf("Get(2) = %d (ok=%v)  %s\n", val, ok, cache)

	cache.Put(4, 4) // evicts key 1
	fmt.Printf("После Put(4,4):    %s (key 1 evicted)\n", cache)

	val, ok = cache.Get(1)
	fmt.Printf("Get(1) = %d (ok=%v)  %s\n", val, ok, cache)

	val, _ = cache.Get(3)
	fmt.Printf("Get(3) = %d         %s\n", val, cache)

	val, _ = cache.Get(4)
	fmt.Printf("Get(4) = %d         %s\n", val, cache)
}

// String — для красивой визуализации состояния кэша
func (c *LRUCache) String() string {
	result := "MRU ["
	for e := c.order.Front(); e != nil; e = e.Next() {
		ent := e.Value.(*entry)
		result += fmt.Sprintf("(%d:%d)", ent.key, ent.value)
		if e.Next() != nil {
			result += " ↔ "
		}
	}
	result += "] LRU"
	return result
}
