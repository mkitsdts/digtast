package openai

import "strings"

// Provider describes an OpenAI-compatible service preset.
type Provider struct {
	ID      string
	Label   string
	BaseURL string
}

var providers = []Provider{
	{ID: "openai", Label: "OpenAI", BaseURL: "https://api.openai.com/v1"},
	{ID: "deepseek", Label: "DeepSeek", BaseURL: "https://api.deepseek.com"},
	{ID: "qwen", Label: "通义千问", BaseURL: "https://dashscope.aliyuncs.com/compatible-mode/v1"},
	{ID: "doubao", Label: "豆包 / 火山方舟", BaseURL: "https://ark.cn-beijing.volces.com/api/v3"},
	{ID: "custom", Label: "自定义 / 其他", BaseURL: ""},
}

// Providers returns the built-in provider presets.
func Providers() []Provider {
	return append([]Provider(nil), providers...)
}

// ProviderBaseURL returns the default base URL for a provider.
func ProviderBaseURL(id string) string {
	id = normalizeProvider(id)
	for _, provider := range providers {
		if provider.ID == id {
			return provider.BaseURL
		}
	}
	return ""
}

// ProviderLabel returns the display name for a provider.
func ProviderLabel(id string) string {
	id = normalizeProvider(id)
	for _, provider := range providers {
		if provider.ID == id {
			return provider.Label
		}
	}
	if id == "" {
		return "未知 Provider"
	}
	return id
}

// IsCustomProvider reports whether a provider needs an explicit base URL.
func IsCustomProvider(id string) bool {
	return ProviderBaseURL(id) == ""
}

func normalizeProvider(id string) string {
	id = strings.ToLower(strings.TrimSpace(id))
	if id == "ark" {
		return "doubao"
	}
	return id
}
