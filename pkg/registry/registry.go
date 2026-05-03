package registry

import (
	"context"
	"sync"

	localbk "digital-labor/pkg/lbackend"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/middlewares/filesystem"
)

var (
	backend     *localbk.Local
	backendOnce sync.Once
)

func GetBackendMiddleware() adk.ChatModelAgentMiddleware {
	var err error
	backend, err = localbk.NewBackend(context.TODO(), &localbk.Config{})

	middleware, err := filesystem.New(context.Background(), &filesystem.MiddlewareConfig{
		Backend:        backend,
		StreamingShell: backend,
	})
	if err != nil {
		return nil
	}

	return middleware
}
