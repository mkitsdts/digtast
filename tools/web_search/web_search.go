package websearch

import (
	"context"
	"digital-labor/pkg/registry"
	"log"
	"log/slog"

	"github.com/cloudwego/eino-ext/components/tool/duckduckgo/ddgsearch"
	"github.com/cloudwego/eino-ext/components/tool/duckduckgo/v2"
	"github.com/cloudwego/eino/components/tool"
)

func NewWebSearchTool() tool.InvokableTool {

	searchTool, err := duckduckgo.NewTextSearchTool(context.Background(), &duckduckgo.Config{ // 下面所有这些参数都是默认值，仅作用法展示
		ToolName:   "duckduckgo_search",
		ToolDesc:   "search web for information by duckduckgo",
		MaxResults: 10,
		Region:     duckduckgo.Region(ddgsearch.RegionCN),
	})
	if err != nil {
		log.Fatalf("NewTool of duckduckgo failed, err=%v", err)
	}
	return searchTool
}

func init() {
	registry.RegisterTool(NewWebSearchTool())
	slog.Info("web_search tool registered")
}
