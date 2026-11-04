// Package oai 提供 OpenAI 兼容协议的公共底座：鉴权、HTTP、SSE 流装配、
// 消息编解码与用量统计。各 Provider 适配器（openai、deepseek 等）在此之上
// 组合自己的请求/响应差异，并用 Config 替换默认基础配置（地址、模型）。
package oai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"digital-labor/component/chatmodel"
	"digital-labor/schema"
)

// Config 是 OpenAI 兼容后端共享的基础配置，Provider 在 New 时补充自己的默认值。
type Config struct {
	APIKey     string
	BaseURL    string
	Model      string
	HTTPClient *http.Client
	Headers    http.Header
}

// Base 承载一个 OpenAI 兼容服务的公共运行时。
type Base struct {
	apiKey  string
	baseURL string
	model   string
	http    *http.Client
	headers http.Header

	usageMu sync.RWMutex
	usage   schema.Usage
}

// NewBase 校验配置并应用默认基础地址；cfg 中的非空字段替换默认值。
func NewBase(cfg Config, defaultBaseURL string) (*Base, error) {
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil, errors.New("api key is required")
	}
	if strings.TrimSpace(cfg.Model) == "" {
		return nil, errors.New("model is required")
	}
	baseURL, err := ResolveBaseURL(cfg.BaseURL, defaultBaseURL)
	if err != nil {
		return nil, err
	}
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 2 * time.Minute}
	}
	return &Base{
		apiKey:  cfg.APIKey,
		baseURL: baseURL,
		model:   cfg.Model,
		http:    httpClient,
		headers: cfg.Headers.Clone(),
	}, nil
}

// Model 返回客户端绑定的模型名，作为请求缺省值。
func (b *Base) Model() string { return b.model }

// ResolveBaseURL 规范化基础地址：为空时使用默认地址，并容忍误填的
// /chat/completions 后缀。
func ResolveBaseURL(baseURL, defaultBaseURL string) (string, error) {
	baseURL = strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	parsed, err := url.ParseRequestURI(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid base url: %w", err)
	}
	if strings.HasSuffix(parsed.Path, "/chat/completions") {
		parsed.Path = strings.TrimSuffix(parsed.Path, "/chat/completions")
		baseURL = strings.TrimRight(parsed.String(), "/")
	}
	return baseURL, nil
}

func (b *Base) endpoint(resource string) string {
	return endpoint(b.baseURL, resource)
}

func endpoint(baseURL, resource string) string {
	base, _ := url.Parse(baseURL)
	base.Path = path.Join(base.Path, strings.TrimPrefix(resource, "/"))
	return base.String()
}

// ListModels 拉取 /models 返回的模型 ID 列表（去重、按字典序排序），用于
// 配置厂商后选择要启用的模型。cfg.BaseURL 为空时使用默认地址。
func ListModels(ctx context.Context, cfg Config, defaultBaseURL string) ([]string, error) {
	baseURL, err := ResolveBaseURL(cfg.BaseURL, defaultBaseURL)
	if err != nil {
		return nil, err
	}
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint(baseURL, "models"), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	req.Header.Set("Accept", "application/json")
	for key, values := range cfg.Headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	var payload struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode model list: %w", err)
	}
	seen := make(map[string]struct{}, len(payload.Data))
	names := make([]string, 0, len(payload.Data))
	for _, model := range payload.Data {
		id := strings.TrimSpace(model.ID)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		names = append(names, id)
	}
	sort.Strings(names)
	if len(names) == 0 {
		return nil, errors.New("provider returned no models")
	}
	return names, nil
}

func (b *Base) do(ctx context.Context, method, resource string, body []byte) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, b.endpoint(resource), reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+b.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	for key, values := range b.headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	return b.http.Do(req)
}

// AddUsage 累计一次调用的用量。
func (b *Base) AddUsage(usage schema.Usage) {
	b.usageMu.Lock()
	b.usage.Add(usage)
	b.usageMu.Unlock()
}

// GetUsage 返回累计用量。
func (b *Base) GetUsage() schema.Usage {
	b.usageMu.RLock()
	defer b.usageMu.RUnlock()
	return b.usage
}

type apiError struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    any    `json:"code"`
	} `json:"error"`
}

func checkResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	var apiErr apiError
	if json.Unmarshal(data, &apiErr) == nil && apiErr.Error.Message != "" {
		return fmt.Errorf("api returned %s: %s", resp.Status, apiErr.Error.Message)
	}
	return fmt.Errorf("api returned %s: %s", resp.Status, strings.TrimSpace(string(data)))
}

// CheckStatus 探测模型可用性：先查询 models/{model}，失败再回退到模型列表。
func CheckStatus(ctx context.Context, b *Base) (*chatmodel.Status, error) {
	started := time.Now()
	checkedAt := time.Now().UTC()
	resp, err := b.do(ctx, http.MethodGet, "models/"+url.PathEscape(b.model), nil)
	if err == nil {
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			resp.Body.Close()
			return &chatmodel.Status{Available: true, Model: b.model, Latency: time.Since(started), CheckedAt: checkedAt}, nil
		}
		resp.Body.Close()
	}

	listResp, listErr := b.do(ctx, http.MethodGet, "models", nil)
	if listErr != nil {
		if err != nil {
			return nil, fmt.Errorf("check model status: %w", err)
		}
		return nil, fmt.Errorf("check models status: %w", listErr)
	}
	defer listResp.Body.Close()
	if statusErr := checkResponse(listResp); statusErr != nil {
		return nil, statusErr
	}
	return &chatmodel.Status{Available: true, Model: b.model, Latency: time.Since(started), CheckedAt: checkedAt}, nil
}
