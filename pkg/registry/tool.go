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
)

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
