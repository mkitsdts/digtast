package memory

import (
	"digital-labor/internal/agent/registry"
	"digital-labor/pkg/tool"
	"log/slog"
)

func NewMemorySearchTool() tool.Tool {
	return &MemorySearchTool{}
}

func init() {
	t := NewMemorySearchTool()
	if t == nil {
		slog.Error("memory search tool not registered due to creation failure")
		return
	}
	registry.RegisterTool(t)
	registry.RegisterBaseTool(t)
	slog.Debug("memory search tool registered")
}
