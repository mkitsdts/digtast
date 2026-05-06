package ctxmanager

import (
	"context"
	"sync"
	"testing"
)

func TestGetOrCreate_New(t *testing.T) {
	ctx := GetOrCreate("new-key")
	if ctx == nil {
		t.Fatal("expected non-nil context")
	}
}

func TestGetOrCreate_Existing(t *testing.T) {
	ctx1 := GetOrCreate("reuse-key")
	ctx2 := GetOrCreate("reuse-key")
	if ctx1 != ctx2 {
		t.Fatal("expected same context for same key")
	}
}

func TestSetOrCreate(t *testing.T) {
	customCtx := context.WithValue(context.Background(), "test-key", "test-val")
	SetOrCreate("custom", customCtx)

	got := GetOrCreate("custom")
	if got != customCtx {
		t.Fatal("expected the custom context we set")
	}
}

func TestDelete(t *testing.T) {
	GetOrCreate("to-delete")
	Delete("to-delete")

	ctx := GetOrCreate("to-delete")
	if ctx == nil {
		t.Fatal("expected new context after delete")
	}
}

func TestConcurrency(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := "concurrent"
			ctx := GetOrCreate(key)
			if ctx == nil {
				t.Errorf("goroutine %d got nil context", n)
			}
		}(i)
	}
	wg.Wait()
}

func TestConcurrency_SetAndDelete(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(3)
		go func(n int) {
			defer wg.Done()
			SetOrCreate("race-key", context.Background())
		}(i)
		go func() {
			defer wg.Done()
			GetOrCreate("race-key")
		}()
		go func() {
			defer wg.Done()
			Delete("race-key")
		}()
	}
	wg.Wait()
}
