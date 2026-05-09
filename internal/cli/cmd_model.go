package cli

import (
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"digital-labor/pkg/chatmodel"
	"digital-labor/pkg/conf"

	"github.com/spf13/cobra"
)

var modelCmd = &cobra.Command{
	Use:   "model",
	Short: "Manage model providers",
}

var modelLsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List configured model providers",
	Run: func(cmd *cobra.Command, args []string) {
		models := chatmodel.ModelManager.ListModels()
		if len(models) == 0 {
			infoMsg("No model providers configured.")
			return
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, colorCyan+"Config Name\tProvider\tModels"+colorReset)

		for name, cfg := range models {
			fmt.Fprintf(w, "%s\t%s\t%s\n", name, cfg.Provider, strings.Join(cfg.ModelNames, ", "))
		}
		fmt.Println(colorCyan + "Configured Providers & Models:" + colorReset)
		w.Flush()
	},
}

var modelAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new model provider interactively",
	Run: func(cmd *cobra.Command, args []string) {
		configName := activeState.promptString("Provider Config Name (e.g., my-openai)")
		if configName == "" {
			errorMsg("Config name is required.")
			return
		}

		provider := activeState.promptString("LLM Provider (ark/openai/qwen/deepseek)")
		if provider == "" {
			errorMsg("Provider is required.")
			return
		}

		key := activeState.promptString("API Key")

		modelsInput := activeState.promptString("Model Names (comma separated, e.g., gpt-4,gpt-3.5-turbo)")
		baseURL := activeState.promptOptional("Base URL")

		modelNames := strings.Split(modelsInput, ",")
		for i := range modelNames {
			modelNames[i] = strings.TrimSpace(modelNames[i])
		}

		err := chatmodel.ModelManager.CreateModel(configName, conf.ModelConfig{
			Provider:   provider,
			Key:        key,
			ModelNames: modelNames,
			URL:        baseURL,
		})

		if err != nil {
			errorMsg("Failed to save model configuration: %v", err)
			return
		}

		successMsg("Model configuration '%s' saved!", configName)
	},
}

func init() {
	modelCmd.AddCommand(modelLsCmd)
	modelCmd.AddCommand(modelAddCmd)
}
