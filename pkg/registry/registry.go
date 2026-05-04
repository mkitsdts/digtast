package registry

import (
	"github.com/cloudwego/eino/adk"
)

var (
	handlers []adk.ChatModelAgentMiddleware
)

func RegisterHandler(middleware adk.ChatModelAgentMiddleware) {
	handlers = append(handlers, middleware)
}

func GetHandlers() []adk.ChatModelAgentMiddleware {
	return handlers
}
