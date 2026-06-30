package main

import (
	"sync"
	"testing"
)

func TestUnsafeCounter_Race(t *testing.T) {
	counter := &UnsafeCounter{}
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Inc()
		}()
	}
	wg.Wait()

	if counter.Value() != 1000 {
		t.Errorf("Expected 1000, got %d", counter.Value())
	}
}

func TestSafeCounter_NoRace(t *testing.T) {
	counter := &SafeCounter{}
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Inc()
		}()
	}
	wg.Wait()

	if counter.Value() != 1000 {
		t.Errorf("Expected 1000, got %d", counter.Value())
	}
}
