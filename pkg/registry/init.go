package registry

import (
	"context"

	localbk "digital-labor/pkg/lbackend"
)

func init() {
	backendOnce.Do(func() {
		backend, _ = localbk.NewBackend(context.Background(), &localbk.Config{
			ValidateCommand: nil,
		})
	})
}
