package cache

import (
	"container/list"
)

// entry — элемент LRU кэша
type entry struct {
	key   string
	value interface{}
}

// LRU — базовый LRU кэш (НЕ thread-safe)
// Используется внутри шарда в ShardedLRU
//
// Структура: HashMap + Doubly Linked List
// - HashMap: O(1) поиск по ключу
// - list.List: O(1) перемещение (MoveToFront, Remove)
//
// Конвенция (классическая):
//
//	Front = MRU (Most Recently Used) — самый свежий
//	Back  = LRU (Least Recently Used) — самый старый
type LRU struct {
	capacity int
	items    map[string]*list.Element
	order    *list.List
}

// NewLRU создаёт новый LRU кэш
func NewLRU(capacity int) *LRU {
	if capacity <= 0 {
		panic("cache: capacity must be positive")
	}
	return &LRU{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		order:    list.New(),
	}
}

// Get возвращает значение по ключу
// Перемещает элемент в MRU (в начало)
// Возвращает (value, true) если найден, (nil, false) если нет
func (c *LRU) Get(key string) (interface{}, bool) {
	if elem, ok := c.items[key]; ok {
		c.order.MoveToFront(elem)
		return elem.Value.(*entry).value, true
	}
	return nil, false
}

// Put добавляет или обновляет ключ-значение
// Если кэш переполнен — удаляет LRU (из конца)
func (c *LRU) Put(key string, value interface{}) {
	// Обновление существующего ключа
	if elem, ok := c.items[key]; ok {
		c.order.MoveToFront(elem)
		elem.Value.(*entry).value = value
		return
	}

	// Добавление нового ключа
	elem := c.order.PushFront(&entry{key: key, value: value})
	c.items[key] = elem

	// Eviction LRU при переполнении
	if c.order.Len() > c.capacity {
		if back := c.order.Back(); back != nil {
			c.order.Remove(back)
			delete(c.items, back.Value.(*entry).key)
		}
	}
}

// Delete удаляет ключ из кэша
func (c *LRU) Delete(key string) bool {
	if elem, ok := c.items[key]; ok {
		c.order.Remove(elem)
		delete(c.items, key)
		return true
	}
	return false
}

// Len возвращает количество элементов
func (c *LRU) Len() int {
	return c.order.Len()
}

// Keys возвращает все ключи в порядке от MRU к LRU
func (c *LRU) Keys() []string {
	keys := make([]string, 0, c.order.Len())
	for e := c.order.Front(); e != nil; e = e.Next() {
		keys = append(keys, e.Value.(*entry).key)
	}
	return keys
}
