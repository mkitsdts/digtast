package cli

import (
	"fmt"
	"strings"

	"digital-labor/pkg/conf"
)

func (s *replState) cmdModel(args []string) {
	if len(args) < 1 {
		errorMsg("Usage: /model <ls|add>")
		return
	}
	switch args[0] {
	case "ls":
		s.cmdModelList()
	case "add":
		s.cmdModelAdd()
	default:
		errorMsg("Unknown subcommand: %s. Use 'ls' or 'add'.", args[0])
	}
}

func (s *replState) cmdModelList() {
	if len(conf.Conf.Models) == 0 {
		fmt.Println("No model providers configured.")
		return
	}
	fmt.Println("Configured Providers & Models:")
	for name, cfg := range conf.Conf.Models {
		fmt.Printf("- Provider: %s\n", name)
		fmt.Printf("  Type: %s\n", cfg.Provider)
		fmt.Printf("  Models: %s\n", strings.Join(cfg.ModelNames, ", "))
	}
}

func (s *replState) cmdModelAdd() {
	configName := s.promptString("Provider Config Name (e.g., my-openai)")
	if configName == "" {
		errorMsg("Config name is required.")
		return
	}

	provider := s.promptString("LLM Provider (ark/openai/qwen/deepseek)")
	if provider == "" {
		errorMsg("Provider is required.")
		return
	}

	key := s.promptString("API Key")
	if key == "" {
		errorMsg("API Key is required.")
		return
	}

	modelsInput := s.promptString("Model Names (comma separated, e.g., gpt-4,gpt-3.5-turbo)")
	baseURL := s.promptOptional("Base URL")

	modelNames := strings.Split(modelsInput, ",")
	for i := range modelNames {
		modelNames[i] = strings.TrimSpace(modelNames[i])
	}

	if conf.Conf.Models == nil {
		conf.Conf.Models = make(map[string]conf.ModelConfig)
	}
	conf.Conf.Models[configName] = conf.ModelConfig{
		Provider:   provider,
		Key:        key,
		ModelNames: modelNames,
		URL:        baseURL,
	}
	conf.SaveConfig()
	successMsg("Model configuration '%s' saved!", configName)
}
