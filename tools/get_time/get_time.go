package gettime

import (
	"context"
	"digital-labor/pkg/registry"
	"log/slog"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type GetTimeTool struct {
}

func (t *GetTimeTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: "get_time",
		Desc: "Get the current time",
	}, nil
}

func (t *GetTimeTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	now := time.Now().GoString()
	return now, nil
}

func NewGetTimeTool() tool.InvokableTool {
	return &GetTimeTool{}
}

func init() {
	registry.RegisterTool(NewGetTimeTool())
	slog.Info("get_time tool registered")
}
