package browser_use

import (
	"digital-labor/internal/agent/registry"
	"digital-labor/pkg/ctxmanager"
	"log/slog"
)

// NewBrowserTool builds the browser tool without launching a browser. The
// browser starts lazily on the first browser_use action.
func NewBrowserTool() *Tool {
	ctx := ctxmanager.GetOrCreate("browser_use")
	return NewLazyBrowserUseTool(ctx, &Config{
		Headless: false, // set headless false that container will run a true browser and user can see the browser at the same time
	})
}

func init() {
	t := NewBrowserTool()
	registry.RegisterTool(t)
	registry.RegisterBaseTool(t)
	slog.Debug("register browser_use tool")
}
