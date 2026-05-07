package cli

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "/",
		Short: "Digital Labor REPL Commands",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}

	exitREPL bool
)

func init() {
	// Root command setup
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	// Add subcommands
	rootCmd.AddCommand(lsCmd)
	rootCmd.AddCommand(useCmd)
	rootCmd.AddCommand(newCmd)
	rootCmd.AddCommand(resetCmd)
	rootCmd.AddCommand(modelCmd)
	rootCmd.AddCommand(channelCmd)
	rootCmd.AddCommand(exitCmd)
}

var exitCmd = &cobra.Command{
	Use:     "exit",
	Aliases: []string{"quit"},
	Short:   "Exit the application",
	Run: func(cmd *cobra.Command, args []string) {
		exitREPL = true
	},
}

// Commands will be defined in other files using the logic from cmd_*.go
