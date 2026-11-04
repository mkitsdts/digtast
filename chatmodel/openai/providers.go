package openai

import (
	"context"

	"digital-labor/component/chatmodel"
	"digital-labor/component/chatmodel/oai"
)

// ListModels 拉取 /models 暴露的模型 ID 列表。
func ListModels(ctx context.Context, opts chatmodel.ClientOptions) ([]string, error) {
	return oai.ListModels(ctx, oai.Config{APIKey: opts.APIKey, BaseURL: opts.BaseURL, HTTPClient: opts.HTTPClient}, DefaultBaseURL)
}

func init() {
	chatmodel.RegisterProvider(chatmodel.Provider{
		ID:      "openai",
		Label:   "OpenAI",
		BaseURL: DefaultBaseURL,
		NewClient: func(opts chatmodel.ClientOptions, model string) (chatmodel.ChatModel, error) {
			return New(Config{APIKey: opts.APIKey, BaseURL: opts.BaseURL, Model: model, HTTPClient: opts.HTTPClient})
		},
		ListModels: ListModels,
	})
}
