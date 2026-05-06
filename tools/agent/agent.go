package agent

import (
	"digital-labor/pkg/registry"
	"log/slog"

	"github.com/cloudwego/eino/components/tool"
)

func NewSubAgentTool() tool.InvokableTool {
	return &SubAgentTool{}
}

func init() {
	t := NewSubAgentTool()
	if t == nil {
		slog.Error("sub_agent tool not registered due to creation failure")
		return
	}
	registry.RegisterTool(t)
	slog.Info("sub_agent tool registered")
}
