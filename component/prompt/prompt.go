package prompt

import (
	"context"
	"strings"
)

// Loader 提供系统提示词的抽象，AgentLoop 在每次运行前通过它组装提示词。
type Loader interface {
	LoadPrompt(ctx context.Context) (string, error)
}

type LoaderFunc func(context.Context) (string, error)

func (f LoaderFunc) LoadPrompt(ctx context.Context) (string, error) { return f(ctx) }

type MultiLoader []Loader

func (m MultiLoader) LoadPrompt(ctx context.Context) (string, error) {
	var parts []string
	for _, loader := range m {
		if loader == nil {
			continue
		}
		prompt, err := loader.LoadPrompt(ctx)
		if err != nil {
			return "", err
		}
		if prompt != "" {
			parts = append(parts, prompt)
		}
	}
	return strings.Join(parts, "\n\n"), nil
}
