package registry

import (
	"context"
	"log/slog"
	"sync"

	"github.com/cloudwego/eino/components/tool"
)

var (
	tools  = make([]tool.BaseTool, 0)
	toolMu sync.RWMutex

	baseTools  = make([]tool.BaseTool, 0)
	baseToolMu sync.RWMutex
)

func RegisterBaseTool(t tool.BaseTool) {
	if t == nil {
		return
	}
	baseToolMu.Lock()
	defer baseToolMu.Unlock()
	info, err := t.Info(context.Background())
	if err != nil {
		slog.Error("failed to get tool info", "error", err)
		return
	}
	slog.Info("registering tool", "name", info.Name)
	baseTools = append(baseTools, t)
}

func RegisterTool(t tool.BaseTool) {
	if t == nil {
		return
	}
	toolMu.Lock()
	defer toolMu.Unlock()
	info, err := t.Info(context.Background())
	if err != nil {
		slog.Error("failed to get tool info", "error", err)
		return
	}
	slog.Info("registering tool", "name", info.Name)
	tools = append(tools, t)
}

func GetTools() []tool.BaseTool {
	toolMu.RLock()
	defer toolMu.RUnlock()

	result := make([]tool.BaseTool, len(tools))
	copy(result, tools)
	return result
}

func GetToolsByType(typ string) []tool.BaseTool {
	toolMu.RLock()
	defer toolMu.RUnlock()

	var result []tool.BaseTool
	var targetNames []string

	switch typ {
	case "research":
		targetNames = []string{"duckduckgo_search", "browser_use"}
	case "execution":
		targetNames = []string{"ask_user"}
	default:
		return nil
	}

	nameMap := make(map[string]bool)
	for _, n := range targetNames {
		nameMap[n] = true
	}

	for _, t := range tools {
		info, err := t.Info(context.Background())
		if err != nil {
			continue
		}
		if nameMap[info.Name] {
			result = append(result, t)
		}
	}
	return result
}
