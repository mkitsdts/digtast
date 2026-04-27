package registry

import (
	"sync"
	"testing"
)

func TestGetRegistry_Singleton(t *testing.T) {
	// backendOnce ensures only one backend is created.
	// Multiple calls should return the same backend (or same error).
	r1 := GetBackendMiddleware()
	r2 := GetBackendMiddleware()

	if r1 != r2 {
		t.Fatal("expected same backend instance")
	}
}

func TestConcurrency_GetBackendMiddleware(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			GetBackendMiddleware()
		}()
	}
	wg.Wait()
	// No assertion — just verifying no race condition
}
