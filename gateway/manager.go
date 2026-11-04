package gateway

import (
	"fmt"
	"log/slog"
	"sync"

	cgw "digital-labor/component/gateway"
	"digital-labor/config"
	"digital-labor/pkg/ctxmanager"
)

var (
	runningMu       sync.Mutex
	runningChannels = make(map[string]cgw.MessageChannel)
)

// GetAllChannels 返回所有已注册的网关 kind。
func GetAllChannels() []string { return cgw.GetAllChannels() }

func LoadChannels() {
	if config.Conf.Channels == nil {
		slog.Info("channels is nil")
		return
	}
	for kind := range config.Conf.Channels {
		if !config.IsChannelEnabled(kind) {
			slog.Info("channel disabled, skip", "kind", kind)
			continue
		}
		if err := EnableChannel(kind); err != nil {
			slog.Error("failed to enable channel", "kind", kind, "error", err)
		}
	}
}

func CreateChannel(kind string, params map[string]any) error {
	return config.SetChannel(kind, params)
}

func RemoveChannel(kind string) error {
	DisableChannel(kind)
	return config.RemoveChannel(kind)
}

func EnableChannel(kind string) error {
	params, ok := config.Conf.Channels[kind]
	if !ok {
		return fmt.Errorf("channel config not found: %s", kind)
	}

	runningMu.Lock()
	if _, ok := runningChannels[kind]; ok {
		runningMu.Unlock()
		return nil // Already running
	}
	runningMu.Unlock()

	channel := cgw.NewChannel(kind)
	if channel == nil {
		return fmt.Errorf("channel not registered: %s", kind)
	}

	ctx := ctxmanager.GetOrCreate(kind)
	if err := channel.Init(params); err != nil {
		return err
	}
	go channel.Serve(ctx)

	runningMu.Lock()
	runningChannels[kind] = channel
	runningMu.Unlock()
	slog.Info("channel enabled", "kind", kind)
	return nil
}

func DisableChannel(kind string) error {
	runningMu.Lock()
	defer runningMu.Unlock()
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
	runningMu.Lock()
	defer runningMu.Unlock()
	_, ok := runningChannels[kind]
	return ok
}
