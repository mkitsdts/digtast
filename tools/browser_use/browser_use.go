package browser_use

import (
	"digital-labor/pkg/ctxmanager"
	"digital-labor/pkg/registry"
	"log/slog"

	"github.com/cloudwego/eino-ext/components/tool/browseruse"
)

func NewBrowserTool() *browseruse.Tool {
	ctx := ctxmanager.GetOrCreate("browser_use")

	var err error
	browserTool, err := browseruse.NewBrowserUseTool(ctx, &browseruse.Config{
		Headless: false, // set headless false that container will run a true browser and user can see the browser at the same time
	})
	if err != nil {
		slog.Error("create browseruse tool failed", "err", err)
	}
	return browserTool
}

func init() {
	registry.RegisterTool(NewBrowserTool())
	slog.Info("register browser_use tool")
}
