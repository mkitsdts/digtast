package local

import (
	"context"
	"sync"
)

var (
	backend     *Local
	backendOnce sync.Once
)

func init() {
	backendOnce.Do(func() {
		backend, _ = NewBackend(context.Background(), &Config{
			ValidateCommand: nil,
		})
	})
}

func GetBackend() *Local {
	return backend
}
