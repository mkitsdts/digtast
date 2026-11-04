package gateway

import (
	"log/slog"
	"sync"
)

var (
	registerMu sync.RWMutex
	register   = make(map[string]func(kind string) MessageChannel)
)

// RegisterChannel 供网关实现在 init 中注册自己。
func RegisterChannel(kind string, f func(kind string) MessageChannel) {
	registerMu.Lock()
	register[kind] = f
	registerMu.Unlock()
	slog.Debug("register channel", "kind", kind)
}

// NewChannel 按 kind 创建已注册的网关通道实例。
func NewChannel(kind string) MessageChannel {
	registerMu.RLock()
	f, ok := register[kind]
	registerMu.RUnlock()
	if !ok {
		slog.Warn("channel not found", "kind", kind)
		return nil
	}
	return f(kind)
}

// GetAllChannels 返回所有已注册的网关 kind。
func GetAllChannels() []string {
	registerMu.RLock()
	defer registerMu.RUnlock()
	kinds := make([]string, 0, len(register))
	for kind := range register {
		kinds = append(kinds, kind)
	}
	return kinds
}
