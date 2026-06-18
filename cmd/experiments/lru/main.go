package main

import "fmt"

type Node struct {
	key   int
	value int
	prev  *Node
	next  *Node
}

type LRUCache struct {
	capacity int
	nodes    map[int]*Node
	head     *Node
	tail     *Node
}

func Constructor(capacity int) LRUCache {
	tail := &Node{}
	head := &Node{}

	tail.prev = head
	head.next = tail

	return LRUCache{
		tail:     tail,
		head:     head,
		capacity: capacity,
		nodes:    make(map[int]*Node, capacity),
	}
}

func (c *LRUCache) Get(key int) int {
	if node, ok := c.nodes[key]; ok {
		c.moveToHead(node)
		return node.value
	}

	return -1
}

func (c *LRUCache) Put(key int, value int) {
	if node, ok := c.nodes[key]; ok {
		node.value = value
		c.moveToHead(node)
		return
	}

	newNode := &Node{key: key, value: value}
	c.nodes[key] = newNode
	c.addToHead(newNode)

	if len(c.nodes) > c.capacity {
		removed := c.removeTail()
		delete(c.nodes, removed.key)
	}
}

func (c *LRUCache) removeNode(node *Node) {
	// Познакомить соседей друг с другом (Вася уходит)
	node.prev.next = node.next
	node.next.prev = node.prev
}

func (c *LRUCache) moveToHead(node *Node) {
	c.removeNode(node)
	c.addToHead(node)
}

func (c *LRUCache) addToHead(node *Node) {
	//c.head охранник
	//c.head.next Алиса
	//node Вася

	node.next = c.head.next
	node.prev = c.head

	c.head.next.prev = node
	c.head.next = node
}

func (c *LRUCache) removeTail() *Node {
	// 1. Находим самый старый узел (перед дверью)
	node := c.tail.prev

	// 2. Вырываем из списка
	c.removeNode(node)

	// 3. Возвращаем, чтобы удалить из map
	return node
}

func main() {
	cache := Constructor(2)

	cache.Put(1, 1)
	cache.Put(2, 2)
	fmt.Println(cache.Get(1)) // 1

	cache.Put(3, 3)           // Удаляет ключ 2
	fmt.Println(cache.Get(2)) // -1 (не найден)

	cache.Put(4, 4)           // Удаляет ключ 1
	fmt.Println(cache.Get(1)) // -1
	fmt.Println(cache.Get(3)) // 3
	fmt.Println(cache.Get(4)) // 4
}
