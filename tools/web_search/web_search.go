package websearch

import (
	"digital-labor/pkg/ctxmanager"
	"digital-labor/pkg/tool"
	"log/slog"

	"github.com/cloudwego/eino-ext/components/tool/browseruse"
)

func NewBrowserTool() *browseruse.Tool {
	ctx := ctxmanager.GetOrCreate("browser_use")

	var err error
	browserTool, err := browseruse.NewBrowserUseTool(ctx, &browseruse.Config{})
	if err != nil {
		slog.Error("create browseruse tool failed", "err", err)
	}
	return browserTool
}

func init() {
	tool.RegisterTool(NewBrowserTool())
}
