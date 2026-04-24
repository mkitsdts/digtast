package registry

import (
	"sync"
	"testing"
)

func TestGetRegistry_NilBackend(t *testing.T) {
	// GetRegistry creates the Eino local backend lazily.
	// Without a real workspace path configured, this may fail.
	// This test documents current behavior — it will error
	// if ctxmanager/workspace are not properly initialized.
	_, err := GetRegistry()
	if err != nil {
		t.Logf("GetRegistry returned error (expected if workspace not configured): %v", err)
	}
}

func TestGetRegistry_Singleton(t *testing.T) {
	// backendOnce ensures only one backend is created.
	// Multiple calls should return the same backend (or same error).
	r1, err1 := GetRegistry()
	r2, err2 := GetRegistry()

	if r1 != r2 {
		t.Fatal("expected same backend instance")
	}
	if (err1 == nil) != (err2 == nil) {
		t.Fatalf("expected consistent error state: err1=%v, err2=%v", err1, err2)
	}
}

func TestConcurrency_GetRegistry(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			GetRegistry()
		}()
	}
	wg.Wait()
	// No assertion — just verifying no race condition
}
