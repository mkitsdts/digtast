package chatmodel

import (
	"errors"
	"strings"

	"digital-labor/pkg/llm"
	openailLm "digital-labor/pkg/llm/openai"
)

const (
	DEEPSEEK = "deepseek"
	QWEN     = "qwen"
	DOUBAO   = "doubao"
	ARK      = "ark"
	OPENAI   = "openai"
)

func newChatModel(provider, key, url, name string) (llm.LLM, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" {
		return nil, errors.New("provider is empty")
	}
	switch provider {
	case OPENAI, DEEPSEEK, QWEN, DOUBAO, ARK:
		return openailLm.New(openailLm.Config{APIKey: key, BaseURL: url, Model: name})
	default:
		return nil, errors.New("unknown provider: " + provider)
	}
}
