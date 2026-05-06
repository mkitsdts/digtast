package gateway

import (
	"digital-labor/pkg/conf"
	"digital-labor/pkg/ctxmanager"
	"fmt"
	"log/slog"
)

var register = make(map[string]func(kind string) MessageChannel)
var runningChannels = make(map[string]MessageChannel)

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
	for kind := range conf.Conf.Channels {
		EnableChannel(kind)
	}
}

func GetAllChannels() []string {
	var kinds []string
	for kind := range register {
		kinds = append(kinds, kind)
	}
	return kinds
}

func CreateChannel(kind string, params map[string]any) error {
	if conf.Conf.Channels == nil {
		conf.Conf.Channels = make(map[string]map[string]any)
	}
	conf.Conf.Channels[kind] = params
	return conf.SaveConfig()
}

func RemoveChannel(kind string) error {
	DisableChannel(kind)
	delete(conf.Conf.Channels, kind)
	return conf.SaveConfig()
}

func EnableChannel(kind string) error {
	params, ok := conf.Conf.Channels[kind]
	if !ok {
		return fmt.Errorf("channel config not found: %s", kind)
	}

	if _, ok := runningChannels[kind]; ok {
		return nil // Already running
	}

	f, ok := register[kind]
	if !ok {
		return fmt.Errorf("channel not registered: %s", kind)
	}

	c := f(kind)
	ctx := ctxmanager.GetOrCreate(kind)
	if err := c.Init(params); err != nil {
		return err
	}
	go c.Serve(ctx)
	runningChannels[kind] = c
	slog.Info("channel enabled", "kind", kind)
	return nil
}

func DisableChannel(kind string) error {
	if _, ok := runningChannels[kind]; !ok {
		return nil // Not running
	}

	// Currently MessageChannel doesn't have a Stop/Close method.
	// We might need to cancel the context.
	ctxmanager.Remove(kind)
	delete(runningChannels, kind)
	slog.Info("channel disabled", "kind", kind)
	return nil
}

func IsChannelRunning(kind string) bool {
	_, ok := runningChannels[kind]
	return ok
}
