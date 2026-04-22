package registry

import (
	"digital-labor/pkg/ctxmanager"

	localbk "github.com/cloudwego/eino-ext/adk/backend/local"
)

var backend *localbk.Local

func GetRegistry() (*localbk.Local, error) {
	ctx := ctxmanager.GetOrCreate("backand")

	var err error
	backend, err = localbk.NewBackend(ctx, &localbk.Config{})
	if err != nil {
		return nil, err
	}
	return backend, nil
}
