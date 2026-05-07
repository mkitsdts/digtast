package cli

import (
	"digital-labor/pkg/conf"
	"digital-labor/pkg/ctxmanager"
	"digital-labor/pkg/gateway"

	"github.com/spf13/cobra"
)

var channelCmd = &cobra.Command{
	Use:   "channel <kind>",
	Short: "Start a channel (qq, telegram, etc.)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
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
	},
}
