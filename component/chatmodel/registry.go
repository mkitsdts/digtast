package chatmodel

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"sync"
)

// ClientOptions 是构建模型客户端的基础配置。BaseURL 非空时替换 Provider 的
// 默认地址，用于自定义网关或代理。
type ClientOptions struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

// Provider 描述一个通过 init 注册的模型厂商。注册模式与 prompts 一致：
// 新增厂商只需在对应适配器包的 init 中调用 RegisterProvider，TUI 与 Agent
// 直接使用注册结果；具体使用哪些模型由用户在配置厂商后从 /models 启用。
type Provider struct {
	ID    string
	Label string
	// BaseURL 是厂商默认地址，可被 ClientOptions.BaseURL 替换。
	BaseURL string
	// NewClient 为指定模型构建客户端。
	NewClient func(opts ClientOptions, model string) (ChatModel, error)
	// ListModels 拉取厂商 /models 暴露的模型 ID 列表。
	ListModels func(ctx context.Context, opts ClientOptions) ([]string, error)
}

var (
	registryMu sync.RWMutex
	registered = map[string]Provider{}
)

// RegisterProvider 注册一个厂商。ID 全局唯一，重复注册会 panic，便于启动时
// 暴露错误。
func RegisterProvider(provider Provider) {
	if provider.ID == "" || provider.Label == "" || provider.NewClient == nil || provider.ListModels == nil {
		panic(fmt.Sprintf("chatmodel: invalid provider registration %#v", provider))
	}
	registryMu.Lock()
	defer registryMu.Unlock()
	if _, exists := registered[provider.ID]; exists {
		panic("chatmodel: provider already registered: " + provider.ID)
	}
	registered[provider.ID] = provider
}

// Providers 返回全部已注册厂商，按 ID 排序。
func Providers() []Provider {
	registryMu.RLock()
	defer registryMu.RUnlock()
	providers := make([]Provider, 0, len(registered))
	for _, provider := range registered {
		providers = append(providers, provider)
	}
	sort.Slice(providers, func(i, j int) bool { return providers[i].ID < providers[j].ID })
	return providers
}

// FindProvider 按 ID 查找已注册厂商。
func FindProvider(id string) (Provider, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	provider, ok := registered[id]
	return provider, ok
}
