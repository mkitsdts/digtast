package skill

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"digital-labor/component/tool"
	"digital-labor/schema"
)

type LoadSkillTool struct {
	manager    *Manager
	authorizer Authorizer
}

func NewLoadSkillTool(manager *Manager, authorizer Authorizer) (*LoadSkillTool, error) {
	if manager == nil {
		return nil, errors.New("skill manager is required")
	}
	return &LoadSkillTool{manager: manager, authorizer: authorizer}, nil
}

func (t *LoadSkillTool) Definition(context.Context) (schema.ToolDefinition, error) {
	return schema.ToolDefinition{
		Name:        "load_skill",
		Description: "Load the complete instructions for one enabled skill. Use the exact name from the available-skills section before applying that skill.",
		Parameters: tool.ObjectSchema(map[string]any{
			"name": map[string]any{"type": "string", "description": "Exact skill name to load."},
		}, "name"),
	}, nil
}

func (t *LoadSkillTool) Execute(ctx context.Context, arguments json.RawMessage) (*tool.Result, error) {
	var input struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(arguments, &input); err != nil {
		return nil, err
	}
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return nil, errors.New("skill name is required")
	}
	content, err := t.manager.LoadSkill(ctx, input.Name, t.authorizer)
	if err != nil {
		return nil, err
	}
	return &tool.Result{Content: content}, nil
}
