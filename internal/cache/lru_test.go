package cache

import (
	"fmt"
	"testing"
)

func TestLRU_Basic(t *testing.T) {
	c := NewLRU(2)

	// Put + Get
	c.Put("a", 1)
	c.Put("b", 2)

	if v, ok := c.Get("a"); !ok || v != 1 {
		t.Errorf("Get(a) = %v, %v; want 1, true", v, ok)
	}

	// Eviction: добавляем "c", должен удалиться "b" (LRU)
	c.Put("c", 3)

	if _, ok := c.Get("b"); ok {
		t.Error("Get(b) should return false after eviction")
	}
	if v, ok := c.Get("c"); !ok || v != 3 {
		t.Errorf("Get(c) = %v, %v; want 3, true", v, ok)
	}
}

func TestLRU_OrderAfterGet(t *testing.T) {
	c := NewLRU(2)

	c.Put("a", 1)
	c.Put("b", 2)

	// Get("a") перемещает "a" в MRU
	c.Get("a")

	// Теперь LRU = "b", добавляем "c" → "b" должен удалиться
	c.Put("c", 3)

	if _, ok := c.Get("b"); ok {
		t.Error("Get(b) should return false — it was LRU")
	}
	if v, ok := c.Get("a"); !ok || v != 1 {
		t.Errorf("Get(a) = %v, %v; want 1, true", v, ok)
	}
}

func TestLRU_UpdateExisting(t *testing.T) {
	c := NewLRU(2)

	c.Put("a", 1)
	c.Put("a", 10) // обновление

	if v, _ := c.Get("a"); v != 10 {
		t.Errorf("Get(a) = %v; want 10", v)
	}
	if c.Len() != 1 {
		t.Errorf("Len() = %d; want 1", c.Len())
	}
}

// Тест равномерности распределения по шардам
func TestSharded_Distribution(t *testing.T) {
	cache := NewSharded(1000, 16)

	counts := make([]int, 16)

	// Добавляем 10000 ключей
	for i := 0; i < 10000; i++ {
		key := fmt.Sprintf("key_%d", i)
		cache.Put(key, i)

		// Определяем шард
		sh := cache.getShard(key)
		for idx, s := range cache.shards {
			if s == sh {
				counts[idx]++
				break
			}
		}
	}

	// Ожидаем ~625 ключей на шард (10000 / 16)
	expected := 10000 / 16
	for i, count := range counts {
		// Допускаем отклонение ±30%
		if count < expected*7/10 || count > expected*13/10 {
			t.Errorf("Shard %d: неравномерное распределение (%d, ожидалось ~%d)",
				i, count, expected)
		}
	}

	t.Logf("Распределение по шардам: %v", counts)
}

func TestSharded_Concurrent(t *testing.T) {
	cache := NewSharded(1000, 16)

	done := make(chan bool)

	// 10 writers
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				cache.Put(key, j)
			}
			done <- true
		}(i)
	}

	// 10 readers
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				cache.Get(key)
			}
			done <- true
		}(i)
	}

	// Ждём завершения всех горутин
	for i := 0; i < 20; i++ {
		<-done
	}
}
