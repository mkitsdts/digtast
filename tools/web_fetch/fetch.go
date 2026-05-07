package webfetch

import (
	"digital-labor/pkg/registry"
	"net/http"
)

var fetchtool *FetchTool = &FetchTool{
	client: &http.Client{},
}

func init() (
	registry.RegisterTool(fetchtool)
)
