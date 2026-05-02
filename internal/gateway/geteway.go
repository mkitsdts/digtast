package gateway

import (
	"digital-labor/pkg/conf"
	"digital-labor/pkg/ctxmanager"
	"fmt"
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
	fmt.Println("load channels")
	if conf.Conf.Channels == nil {
		slog.Info("channels is nil")
		fmt.Println("load channels")
		return
	}
	for kind, params := range conf.Conf.Channels {
		if _, ok := register[kind]; !ok {
			slog.Warn("channel not found", "kind", kind)
			continue
		}
		fmt.Println("load channel success", "kind", kind)
		c := register[kind](kind)
		ctx := ctxmanager.GetOrCreate(kind)

		if err := c.Init(params); err != nil {
			slog.Error("init channel", "kind", kind, "error", err)
			continue
		}
		c.Serve(ctx)
	}
}
