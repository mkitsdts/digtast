package agent

import (
	"context"
	"digital-labor/pkg/model"
	"digital-labor/pkg/workspace"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/tool"
)

const defaultModelDescription = "你是一位云端数字助理。你的核心目标是成为用户高效、可靠且易于沟通的智能伙伴。你应具备卓越的理解能力、严谨的逻辑思维和强大的信息整合能力，旨在帮助用户解决问题、获取知识、激发创意并提升效率。你的回答应始终体现专业性、准确性和用户友好性。"

type PromptBuilder struct {
}

func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{}
}

func (b *PromptBuilder) Build(ctx model.PromptContext) string {
	parts := make([]string, 0, 5)

	if root := strings.TrimSpace(b.loadPrompt()); root != "" {
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

func (b *PromptBuilder) loadPrompt() string {
	// TODO:
	result := ""
	prompts := workspace.GetPromptCreators()
	for _, prompt := range prompts {
		if p, err := prompt.GetPromptImpl(); err == nil {
			result += p + "\n\n"
		}
	}
	return ""
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
