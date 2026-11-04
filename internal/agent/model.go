package agent

import (
	"errors"
	"strings"

	"digital-labor/chatmodel/openai"
	"digital-labor/component/chatmodel"
	"digital-labor/config"
)

func newChatModel(cfg config.ModelConfig, model string) (chatmodel.ChatModel, error) {
	provider := strings.ToLower(strings.TrimSpace(cfg.Provider))
	if provider == "" {
		return nil, errors.New("model provider is required")
	}
	baseURL := strings.TrimSpace(cfg.URL)
	if baseURL == "" {
		baseURL = openai.ProviderBaseURL(provider)
	}
	if baseURL == "" {
		return nil, errors.New("provider " + provider + " needs a base url")
	}
	return openai.New(openai.Config{APIKey: cfg.Key, BaseURL: baseURL, Model: model})
}
