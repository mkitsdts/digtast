package ctxmanager

import (
	"context"
	"sync"
)

type ctxInfo struct {
	ctx    context.Context
	cancel context.CancelFunc
}

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
	globalManager.ctxs.Store(key, ctxInfo{ctx: ctx})
}

func GetOrCreate(key string) context.Context {
	if info, ok := globalManager.ctxs.Load(key); ok {
		return info.(ctxInfo).ctx
	}
	ctx, cancel := context.WithCancel(context.Background())
	globalManager.ctxs.Store(key, ctxInfo{ctx: ctx, cancel: cancel})
	return ctx
}

func Remove(key string) {
	if info, ok := globalManager.ctxs.Load(key); ok {
		if cancel := info.(ctxInfo).cancel; cancel != nil {
			cancel()
		}
	}
	globalManager.ctxs.Delete(key)
}

func Delete(key string) {
	Remove(key)
}
