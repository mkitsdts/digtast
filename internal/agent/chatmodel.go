package agent

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/ark"
	"github.com/cloudwego/eino-ext/components/model/deepseek"
	"github.com/cloudwego/eino-ext/components/model/qwen"
	"github.com/cloudwego/eino/components/model"
)

const (
	DEEPSEEK = "deepseek"
	QWEN     = "qwen"
	DOUBAO   = "doubao"

	DeepseekDefaultBaseURL = ""
	DeepseekDefaultModel   = ""

	QwenDefaultBaseURL = ""
	QwenDefaultModel   = ""

	DoubaoDefaultBaseURL = ""
	DoubaoDefaultModel   = ""
)

func newChatModel(ctx context.Context, provider, key, url, name string) (model.ToolCallingChatModel, error) {
	provider = strings.ToLower(provider)
	if provider == "" {
		return nil, errors.New("provider is empty")
	}
	if key == "" {
		return nil, errors.New("key is empty")
	}
	if name == "" {
		slog.Warn("name is empty, using default")
	}

	switch provider {
	case DEEPSEEK:
		if url == "" {
			slog.Warn("url is empty, using default")
			url = DeepseekDefaultBaseURL
		}
		if name == "" {
			slog.Warn("name is empty, using default")
			name = DeepseekDefaultModel
		}
		return deepseek.NewChatModel(ctx, &deepseek.ChatModelConfig{
			APIKey:  key,
			BaseURL: url,
			Model:   name,
		})
	case QWEN:
		if url == "" {
			slog.Warn("url is empty, using default")
			url = QwenDefaultBaseURL
		}
		if name == "" {
			slog.Warn("name is empty, using default")
			name = QwenDefaultModel
		}
		return qwen.NewChatModel(ctx, &qwen.ChatModelConfig{
			APIKey:  key,
			BaseURL: url,
			Model:   name,
		})
	case DOUBAO:
		if url == "" {
			slog.Warn("url is empty, using default")
			url = DoubaoDefaultBaseURL
		}
		if name == "" {
			slog.Warn("name is empty, using default")
			name = DoubaoDefaultModel
		}
		return ark.NewChatModel(ctx, &ark.ChatModelConfig{
			APIKey:  key,
			BaseURL: url,
			Model:   name,
		})
	default:
		return nil, errors.New("unknown provider")
	}
}
