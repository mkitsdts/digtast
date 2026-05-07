package local

import (
	"context"
	"sync"
)

var (
	backend     *Local
	backendOnce sync.Once

	secureBackend     *SecureLocal
	secureBackendOnce sync.Once
)

func init() {
	backendOnce.Do(func() {
		backend, _ = NewBackend(context.Background(), &Config{
			ValidateCommand: nil,
		})
	})
	secureBackendOnce.Do(func() {
		secureBackend, _ = NewSecureBackend(context.Background(), &Config{
			ValidateCommand: nil,
		})
	})
}

func GetBackend() *Local {
	return backend
}

func GetSecureBackend() *SecureLocal {
	return secureBackend
}
