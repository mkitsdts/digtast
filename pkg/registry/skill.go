package registry

import (
	"digital-labor/pkg/ctxmanager"
	"digital-labor/pkg/workspace"
	"fmt"
	"log/slog"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/middlewares/skill"
)

type Skill struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

func GetSkillMiddleware() adk.ChatModelAgentMiddleware {
	ctx := ctxmanager.GetOrCreate("skill-middleware")
	skillsDir := fmt.Sprintf("%s/skills", workspace.GetWorkspacePath())
	fmt.Println("skillsDir", "path", skillsDir)
	slog.Debug("skillsDir", "path", skillsDir)
	skillBackend, err := skill.NewBackendFromFilesystem(ctx, &skill.BackendFromFilesystemConfig{
		Backend: backend,
		BaseDir: skillsDir,
	})
	if err != nil {
		return nil
	}
	skillMiddleware, err := skill.NewMiddleware(ctx, &skill.Config{
		Backend: skillBackend,
	})
	if err != nil {
		return nil
	}

	return skillMiddleware
}
