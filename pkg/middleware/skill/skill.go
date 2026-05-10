package registry

import (
	"digital-labor/pkg/ctxmanager"
	local "digital-labor/pkg/middleware/lbackend"
	skillmgr "digital-labor/pkg/skill"
	"digital-labor/pkg/workspace"
	"log/slog"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/middlewares/skill"
)

type Skill struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

func GetSkillMiddleware() adk.ChatModelAgentMiddleware {
	return GetSkillMiddlewareForAgent(workspace.DefaultAgentID())
}

func GetSkillMiddlewareForAgent(agentID string) adk.ChatModelAgentMiddleware {
	ctx := ctxmanager.GetOrCreate("skill-middleware")
	skillsDir := skillmgr.NewManagerForAgent(agentID).BaseDir()
	slog.Debug("skillsDir", "path", skillsDir)
	skillBackend, err := skill.NewBackendFromFilesystem(ctx, &skill.BackendFromFilesystemConfig{
		Backend: local.GetBackend(),
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
