package subagenttool

import (
	"context"
	"digital-labor/internal/center"
	"digital-labor/pkg/llm"
	"digital-labor/pkg/registry"
	"digital-labor/pkg/subagent"
	toolpkg "digital-labor/pkg/tool"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
)

type SubAgentTool struct {
}

func (t *SubAgentTool) Definition(context.Context) (llm.ToolDefinition, error) {
	return llm.ToolDefinition{
		Name: "fork_subagent", Description: "创建一个子代理处理复杂子任务。",
		Parameters: toolpkg.ObjectSchema(map[string]any{
			"task_description": map[string]any{"type": "string"},
			"kind":             map[string]any{"type": "string", "enum": []string{"research", "execution"}},
			"async":            map[string]any{"type": "boolean"},
		}, "task_description", "kind"),
	}, nil
}

type SubAgentArguments struct {
	TaskDescription string `json:"task_description"`
	Kind            string `json:"kind"`
	Async           bool   `json:"async"`
}

func (t *SubAgentTool) Execute(ctx context.Context, arguments json.RawMessage) (*toolpkg.Result, error) {
	args := SubAgentArguments{}
	if err := json.Unmarshal(arguments, &args); err != nil {
		return nil, err
	}

	agentId, _ := ctx.Value("agent_id").(string)
	if agentId == "" {
		return nil, errors.New("agent_id not found in context")
	}

	da, err := center.AgentManager.GetAgent(agentId)
	if err != nil {
		// If we can't find the agent via manager, maybe it's not a digital agent but we can still try to get model if it was injected
		slog.Error("failed to get agent from manager", "agent_id", agentId, "err", err)
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}

	cm := da.GetModel()

	// Initial message for the subagent is the task description
	messages := []llm.Message{
		{Role: llm.RoleUser, Content: args.TaskDescription},
	}

	taskId, _ := ctx.Value("task_id").(string)
	if taskId == "" {
		prefix := agentId
		if len(prefix) > 8 {
			prefix = prefix[:8]
		}
		taskId = fmt.Sprintf("sub-%s", prefix)
	}

	slog.Info("forking subagent", "parent_id", agentId, "task_id", taskId, "kind", args.Kind)

	resultChan := subagent.ForkSubAgent(ctx, taskId, args.Kind, &subagent.Config{
		ParentID:        agentId,
		ChatModel:       cm,
		Messages:        messages,
		TaskDescription: args.TaskDescription,
	})

	if resultChan == nil {
		return nil, errors.New("failed to fork subagent")
	}

	if args.Async {
		return &toolpkg.Result{Content: "doing async subagent task"}, nil
	}

	result := <-resultChan
	if !result.Success {
		return nil, fmt.Errorf("subagent execution failed: %v", result.Error)
	}

	return &toolpkg.Result{Content: result.Content}, nil
}

func init() {
	registry.RegisterTool(&SubAgentTool{})
}
