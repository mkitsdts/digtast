package main

import (
	"context"
	"fmt"
	"os"

	"digital-labor/config"
	_ "digital-labor/gateway"
	"digital-labor/internal/agent/workspace"
	_ "digital-labor/prompt"
	_ "digital-labor/tools"

	"github.com/spf13/cobra"
)

func main() {
	workspace.InitWorkspace()
	if err := config.LoadConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "digtast: load config: %v\n", err)
		os.Exit(1)
	}

	if err := GetRootCmd().ExecuteContext(context.Background()); err != nil {
		os.Exit(1)
	}
}

func GetRootCmd() *cobra.Command {
	var browser bool
	cmd := &cobra.Command{
		Use:   "digtast",
		Short: "Single-instance digital labor agent",
		Long:  "Digtast 是单实例数字劳动智能体。直接运行会打开终端对话框。",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if browser {
				config.Conf.Browser.Enabled = true
			}
			return RunTUI(cmd.Context())
		},
		SilenceUsage: true,
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}
	cmd.Flags().BoolVar(&browser, "browser", false, "enable browser mode")
	return cmd
}
