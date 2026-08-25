package askuser

import (
	"digital-labor/pkg/registry"
	"digital-labor/pkg/tool"
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
