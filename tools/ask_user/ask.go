package askuser

import (
	"digital-labor/pkg/registry"
	"log/slog"

	"github.com/cloudwego/eino/components/tool"
)

func NewAskUserTool() tool.InvokableTool {
	return &AskUserTool{}
}

func init() {
	t := NewAskUserTool()
	if t == nil {
		slog.Error("ask user tool not registered due to creation failure")
		return
	}
	registry.RegisterTool(t)
	slog.Info("ask user tool registered")
}
