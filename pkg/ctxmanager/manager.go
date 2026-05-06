package ctxmanager

import (
	"context"
	"sync"
)

type Manager struct {
	ctxs sync.Map
}

var globalManager = New()

func New() *Manager {
	return &Manager{
		ctxs: sync.Map{},
	}
}

func SetOrCreate(key string, ctx context.Context) {
	globalManager.ctxs.Store(key, ctx)
}

func GetOrCreate(key string) context.Context {
	if ctx, ok := globalManager.ctxs.Load(key); ok {
		return ctx.(context.Context)
	}
	ctx := context.Background()
	globalManager.ctxs.Store(key, ctx)
	return ctx
}

func Delete(key string) {
	globalManager.ctxs.Delete(key)
}
