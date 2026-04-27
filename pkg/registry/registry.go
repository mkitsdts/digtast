package registry

import (
	"context"
	"sync"

	localbk "github.com/cloudwego/eino-ext/adk/backend/local"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/middlewares/filesystem"
)

var (
	backend     *localbk.Local
	backendOnce sync.Once
)

func GetBackendMiddleware() adk.ChatModelAgentMiddleware {

	middleware, err := filesystem.New(context.Background(), &filesystem.MiddlewareConfig{
		Backend: backend,
	})
	if err != nil {
		return nil
	}

	return middleware
}
