package cli

import (
	"fmt"

	"digital-labor/pkg/conf"
	"digital-labor/pkg/ctxmanager"
	"digital-labor/pkg/gateway"
)

func (s *replState) cmdChannel(args []string) {
	if len(args) < 1 {
		errorMsg("Usage: /channel <kind>")
		fmt.Println("Available: qq, telegram")
		return
	}
	kind := args[0]

	c := gateway.NewChannel(kind)
	if c == nil {
		errorMsg("Failed to create channel: %s", kind)
		return
	}

	c.Register()
	ctx := ctxmanager.GetOrCreate(kind)
	go c.Serve(ctx)

	if conf.Conf.Channels == nil {
		conf.Conf.Channels = make(map[string]map[string]any)
	}
	conf.Conf.Channels[kind] = c.GetConfig()
	conf.SaveConfig()

	successMsg("Channel '%s' started.", kind)
}
