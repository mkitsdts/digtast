package registry

import (
	"digital-labor/pkg/ctxmanager"
	"sync"

	localbk "github.com/cloudwego/eino-ext/adk/backend/local"
)

var (
	backend     *localbk.Local
	backendOnce sync.Once
)

func GetRegistry() (*localbk.Local, error) {
	var err error
	backendOnce.Do(func() {
		ctx := ctxmanager.GetOrCreate("backend")
		backend, err = localbk.NewBackend(ctx, &localbk.Config{})
	})
	if err != nil {
		return nil, err
	}
	return backend, nil
}
