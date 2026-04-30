package gateway

import (
	"digital-labor/pkg/conf"
	"digital-labor/pkg/ctxmanager"
	"log/slog"
)

var register = make(map[string]func(kind string) MessageChannel)

func NewChannel(kind string) MessageChannel {
	if register == nil {
		slog.Error("register is nil")
		return nil
	}
	if f, ok := register[kind]; ok {
		slog.Info("found channel", "kind", kind)
		return f(kind)
	}
	slog.Warn("channel not found", "kind", kind)
	return nil
}

// 提供通道实现注册
func RegisterChannel(kind string, f func(kind string) MessageChannel) {
	register[kind] = f
	slog.Info("register channel", "kind", kind)
}

func LoadChannels() {
	if conf.Conf.Channels == nil {
		return
	}
	for kind, params := range conf.Conf.Channels {
		if _, ok := register[kind]; !ok {
			continue
		}
		c := register[kind](kind)
		ctx := ctxmanager.GetOrCreate(kind)
		c.Init(params)
		c.Serve(ctx)
	}
}
