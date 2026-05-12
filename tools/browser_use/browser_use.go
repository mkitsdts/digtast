package browser_use

import (
	"digital-labor/pkg/ctxmanager"
	"digital-labor/pkg/registry"
	"log/slog"
)

func NewBrowserTool() *Tool {
	ctx := ctxmanager.GetOrCreate("browser_use")

	var err error
	browserTool, err := NewBrowserUseTool(ctx, &Config{
		Headless: false, // set headless false that container will run a true browser and user can see the browser at the same time
	})
	if err != nil {
		slog.Error("create browseruse tool failed", "err", err)
	}
	return browserTool
}

func init() {
	t := NewBrowserTool()
	registry.RegisterTool(t)
	registry.RegisterBaseTool(t)
	slog.Debug("register browser_use tool")
}
