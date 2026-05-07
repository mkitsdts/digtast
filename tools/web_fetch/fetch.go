package webfetch

import (
	"digital-labor/pkg/registry"
	"net/http"
	"sync"
)

var fetchtool *FetchTool = &FetchTool{
	client: &http.Client{},
	mux:    &sync.Mutex{},
}

func init() {
	registry.RegisterTool(fetchtool)
}
