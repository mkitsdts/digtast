package registry

import (
	"context"
	"digital-labor/pkg/workspace"

	localbk "github.com/cloudwego/eino-ext/adk/backend/local"

	"fmt"
)

func init() {
	path := fmt.Sprintf("%s/skills", workspace.GetWorkspacePath())
	ScanSkills(path)
	backendOnce.Do(func() {
		backend, _ = localbk.NewBackend(context.Background(), &localbk.Config{
			ValidateCommand: nil,
		})
	})
}
