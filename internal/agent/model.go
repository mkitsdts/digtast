package agent

import (
	"errors"
	"strings"

	"digital-labor/config"
	"digital-labor/pkg/chatmodel"
	"digital-labor/pkg/chatmodel/openai"
)

func newChatModel(cfg config.ModelConfig, model string) (chatmodel.ChatModel, error) {
	provider := strings.ToLower(strings.TrimSpace(cfg.Provider))
	switch provider {
	case "openai", "deepseek", "qwen", "doubao", "ark":
		return openai.New(openai.Config{APIKey: cfg.Key, BaseURL: cfg.URL, Model: model})
	case "":
		return nil, errors.New("model provider is required")
	default:
		return nil, errors.New("unknown model provider: " + provider)
	}
}
