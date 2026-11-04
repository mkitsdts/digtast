package tool

import (
	"context"
	"digital-labor/schema"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
)

type Tool interface {
	Definition(ctx context.Context) (schema.ToolDefinition, error)
	Execute(ctx context.Context, arguments json.RawMessage) (*Result, error)
}

type Loader interface {
	LoadTools(ctx context.Context) ([]Tool, error)
}

type StaticLoader []Tool

func (s StaticLoader) LoadTools(context.Context) ([]Tool, error) {
	result := make([]Tool, len(s))
	copy(result, s)
	return result, nil
}

type Registry struct {
	mu    sync.RWMutex
	tools map[string]Tool
}

func NewRegistry() *Registry {
	return &Registry{tools: make(map[string]Tool)}
}

func (r *Registry) Register(ctx context.Context, t Tool) error {
	if t == nil {
		return errors.New("tool is nil")
	}
	definition, err := t.Definition(ctx)
	if err != nil {
		return fmt.Errorf("get tool definition: %w", err)
	}
	if definition.Name == "" {
		return errors.New("tool name is required")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.tools[definition.Name]; exists {
		return fmt.Errorf("tool %q is already registered", definition.Name)
	}
	r.tools[definition.Name] = t
	return nil
}

func (r *Registry) Replace(ctx context.Context, t Tool) error {
	if t == nil {
		return errors.New("tool is nil")
	}
	definition, err := t.Definition(ctx)
	if err != nil {
		return fmt.Errorf("get tool definition: %w", err)
	}
	if definition.Name == "" {
		return errors.New("tool name is required")
	}
	r.mu.Lock()
	r.tools[definition.Name] = t
	r.mu.Unlock()
	return nil
}

func (r *Registry) Get(name string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tools[name]
	return t, ok
}

func (r *Registry) All(ctx context.Context) ([]Tool, []schema.ToolDefinition, error) {
	r.mu.RLock()
	items := make(map[string]Tool, len(r.tools))
	for name, t := range r.tools {
		items[name] = t
	}
	r.mu.RUnlock()

	names := make([]string, 0, len(items))
	for name := range items {
		names = append(names, name)
	}
	sort.Strings(names)
	tools := make([]Tool, 0, len(names))
	definitions := make([]schema.ToolDefinition, 0, len(names))
	for _, name := range names {
		definition, err := items[name].Definition(ctx)
		if err != nil {
			return nil, nil, fmt.Errorf("get definition for tool %q: %w", name, err)
		}
		tools = append(tools, items[name])
		definitions = append(definitions, definition)
	}
	return tools, definitions, nil
}

func Load(ctx context.Context, loaders ...Loader) (*Registry, error) {
	registry := NewRegistry()
	for _, loader := range loaders {
		if loader == nil {
			continue
		}
		tools, err := loader.LoadTools(ctx)
		if err != nil {
			return nil, err
		}
		for _, t := range tools {
			if err := registry.Register(ctx, t); err != nil {
				return nil, err
			}
		}
	}
	return registry, nil
}

func ObjectSchema(properties map[string]any, required ...string) map[string]any {
	schema := map[string]any{
		"type":                 "object",
		"properties":           properties,
		"additionalProperties": false,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}
