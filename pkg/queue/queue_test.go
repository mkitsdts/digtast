package queue

import (
	"sync"
	"testing"

	"github.com/cloudwego/eino/schema"
)

func TestNewQueue(t *testing.T) {
	q := NewQueue[int]()
	if q == nil {
		t.Fatal("expected non-nil queue")
	}
	if !q.Empty() {
		t.Fatal("expected empty queue")
	}
	if q.Size() != 0 {
		t.Fatal("expected size 0")
	}
}

func TestPushPop(t *testing.T) {
	q := NewQueue[string]()
	q.Push("a")
	q.Push("b")
	q.Push("c")

	if q.Size() != 3 {
		t.Fatalf("expected size 3, got %d", q.Size())
	}

	item, ok := q.Pop()
	if !ok || item != "a" {
		t.Fatalf("expected 'a', got '%s' ok=%v", item, ok)
	}

	item, ok = q.Pop()
	if !ok || item != "b" {
		t.Fatalf("expected 'b', got '%s' ok=%v", item, ok)
	}

	item, ok = q.Pop()
	if !ok || item != "c" {
		t.Fatalf("expected 'c', got '%s' ok=%v", item, ok)
	}

	item, ok = q.Pop()
	if ok {
		t.Fatalf("expected empty, got '%s'", item)
	}
}

func TestFront(t *testing.T) {
	q := NewQueue[int]()
	val, ok := q.Front()
	if ok {
		t.Fatalf("expected no front on empty queue, got %d", val)
	}

	q.Push(42)
	val, ok = q.Front()
	if !ok || val != 42 {
		t.Fatalf("expected front 42, got %d ok=%v", val, ok)
	}
	if q.Size() != 1 {
		t.Fatal("front should not remove item")
	}
}

func TestEmpty(t *testing.T) {
	q := NewQueue[int]()
	if !q.Empty() {
		t.Fatal("expected empty")
	}
	q.Push(1)
	if q.Empty() {
		t.Fatal("expected not empty")
	}
	q.Pop()
	if !q.Empty() {
		t.Fatal("expected empty after pop")
	}
}

func TestMessageQueue(t *testing.T) {
	mq := NewMessageQueue()

	// Push to nonexistent key should be a no-op, not panic
	mq.Push("nonexistent", &schema.Message{Content: "1"})

	// Pop from nonexistent key should return nil, false
	msg, ok := mq.Pop("nonexistent")
	if ok || msg != nil {
		t.Fatal("expected nil, false for Pop on nonexistent key")
	}

	// Front on nonexistent key should return nil, false
	msg, ok = mq.Front("nonexistent")
	if ok || msg != nil {
		t.Fatal("expected nil, false for Front on nonexistent key")
	}
}

func TestConcurrency_Queue(t *testing.T) {
	q := NewQueue[int]()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			q.Push(n)
		}(i)
	}
	wg.Wait()

	if q.Size() != 100 {
		t.Fatalf("expected size 100, got %d", q.Size())
	}
}
