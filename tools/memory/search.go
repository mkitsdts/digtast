package memory

import (
	"digital-labor/pkg/registry"
	"log/slog"

	"github.com/cloudwego/eino/components/tool"
)

func NewMemorySearchTool() tool.InvokableTool {
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
