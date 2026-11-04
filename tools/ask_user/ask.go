package askuser

import (
	"digital-labor/component/tool"
	"digital-labor/internal/agent/registry"
	"log/slog"
)

func NewAskUserTool() tool.Tool {
	return &AskUserTool{}
}

func init() {
	t := NewAskUserTool()
	if t == nil {
		slog.Error("ask user tool not registered due to creation failure")
		return
	}
	registry.RegisterTool(t)
	slog.Debug("ask user tool registered")
}
