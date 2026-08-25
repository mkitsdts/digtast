package registry

import (
	"context"
	"log/slog"
	"sort"
	"sync"

	"digital-labor/pkg/agentloop"
	"digital-labor/pkg/tool"
)

var (
	toolMu      sync.RWMutex
	toolsByName = make(map[string]tool.Tool)
	baseNames   = make(map[string]struct{})
	hookMu      sync.RWMutex
	hooks       []agentloop.Hook
)

func RegisterBaseTool(t tool.Tool) {
	name, ok := toolName(t)
	if !ok {
		return
	}
	toolMu.Lock()
	toolsByName[name] = t
	baseNames[name] = struct{}{}
	toolMu.Unlock()
}

func RegisterTool(t tool.Tool) {
	name, ok := toolName(t)
	if !ok {
		return
	}
	toolMu.Lock()
	toolsByName[name] = t
	toolMu.Unlock()
}

func toolName(t tool.Tool) (string, bool) {
	if t == nil {
		return "", false
	}
	definition, err := t.Definition(context.Background())
	if err != nil || definition.Name == "" {
		slog.Error("failed to get tool definition", "error", err)
		return "", false
	}
	return definition.Name, true
}

func GetTools() []tool.Tool {
	toolMu.RLock()
	names := make([]string, 0, len(toolsByName))
	for name := range toolsByName {
		names = append(names, name)
	}
	sort.Strings(names)
	result := make([]tool.Tool, 0, len(names))
	for _, name := range names {
		result = append(result, toolsByName[name])
	}
	toolMu.RUnlock()
	return result
}

func LoadTools(context.Context) ([]tool.Tool, error) { return GetTools(), nil }

type Loader struct{}

func (Loader) LoadTools(ctx context.Context) ([]tool.Tool, error) { return LoadTools(ctx) }

func GetToolsByType(typ string) []tool.Tool {
	toolMu.RLock()
	defer toolMu.RUnlock()
	var result []tool.Tool
	for name, t := range toolsByName {
		switch typ {
		case "research", "researcher":
			if name == "duckduckgo_search" || name == "browser_use" || name == "web_fetch" {
				result = append(result, t)
			}
		case "execution", "executer":
			if name != "fork_subagent" {
				result = append(result, t)
			}
		}
	}
	return result
}

func RegisterHandler(hook agentloop.Hook) {
	if hook == nil {
		return
	}
	hookMu.Lock()
	hooks = append(hooks, hook)
	hookMu.Unlock()
}

func GetHandlers() []agentloop.Hook {
	hookMu.RLock()
	defer hookMu.RUnlock()
	result := make([]agentloop.Hook, len(hooks))
	copy(result, hooks)
	return result
}
