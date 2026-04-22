package agent

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/eino/components/tool"
)

const rootPromptFile = "agent.md"

type PromptContext struct {
	UserInstruction string
	UserPreference  string
	Skills          []string
	Tools           []tool.BaseTool
}

type PromptBuilder struct {
	workspaceDir string
}

func NewPromptBuilder(workspaceDir string) *PromptBuilder {
	return &PromptBuilder{workspaceDir: workspaceDir}
}

func (b *PromptBuilder) Build(ctx PromptContext) string {
	parts := make([]string, 0, 5)

	if root := strings.TrimSpace(b.loadRootPrompt()); root != "" {
		parts = append(parts, root)
	}
	if instruction := strings.TrimSpace(ctx.UserInstruction); instruction != "" {
		parts = append(parts, "## User Setup\n"+instruction)
	}
	if preference := strings.TrimSpace(ctx.UserPreference); preference != "" {
		parts = append(parts, "## User Preference\n"+preference)
	}
	if toolSection := strings.TrimSpace(renderTools(ctx.Tools)); toolSection != "" {
		parts = append(parts, toolSection)
	}
	if skillSection := strings.TrimSpace(renderSkills(ctx.Skills)); skillSection != "" {
		parts = append(parts, skillSection)
	}

	return strings.Join(parts, "\n\n")
}

func (b *PromptBuilder) loadRootPrompt() string {
	if strings.TrimSpace(b.workspaceDir) == "" {
		return ""
	}
	data, err := os.ReadFile(filepath.Join(b.workspaceDir, rootPromptFile))
	if err != nil {
		return ""
	}
	return string(data)
}

func renderTools(tools []tool.BaseTool) string {
	if len(tools) == 0 {
		return ""
	}

	lines := []string{"## Tools"}
	for _, t := range tools {
		if info, err := t.Info(context.Background()); err == nil {
			lines = append(lines, fmt.Sprintf("- %s: %s", info.Name, info.Desc))
		}
	}
	if len(lines) == 1 {
		return ""
	}
	return strings.Join(lines, "\n")
}

func renderSkills(skills []string) string {
	if len(skills) == 0 {
		return ""
	}

	lines := []string{"## Skills"}
	for _, skill := range skills {
		skill = strings.TrimSpace(skill)
		if skill == "" {
			continue
		}
		lines = append(lines, "- "+skill)
	}
	if len(lines) == 1 {
		return ""
	}
	return strings.Join(lines, "\n")
}
