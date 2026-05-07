package subagenttool

import (
	"context"
	"digital-labor/internal/center"
	"digital-labor/pkg/registry"
	"digital-labor/pkg/subagent"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/eino-contrib/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type SubAgentTool struct {
}

func (t *SubAgentTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	properties := orderedmap.New[string, *jsonschema.Schema]()
	properties.Set("task_description", &jsonschema.Schema{
		Type:        "string",
		Description: "对子代理要执行的任务的详细描述。如果是 research 类型，应包含调研的目标和范围；如果是 execution 类型，应包含具体执行的指令。",
	})
	properties.Set("kind", &jsonschema.Schema{
		Type:        "string",
		Description: "子代理的类型：research (调研类，可以使用搜索和浏览器) 或 execution (执行类，可以使用终端和文件操作)。",
		Enum:        []interface{}{"research", "execution"},
	})
	properties.Set("async", &jsonschema.Schema{
		Type:        "boolean",
		Description: "是否异步执行子代理任务。如果为 true，子代理将独立运行并返回结果；如果为 false，子代理将同步执行并阻塞调用者。",
	})

	jsonschema := &jsonschema.Schema{
		Version:    "v0.1",
		ID:         "fork_subagent",
		Type:       "object",
		Properties: properties,
		Required:   []string{"task_description", "kind"},
	}

	return &schema.ToolInfo{
		Name:        "fork_subagent",
		Desc:        "创建一个子代理处理一个复杂的子任务。子代理会独立运行并返回结果。",
		ParamsOneOf: schema.NewParamsOneOfByJSONSchema(jsonschema),
	}, nil
}

type SubAgentArguments struct {
	TaskDescription string `json:"task_description"`
	Kind            string `json:"kind"`
	Async           bool   `json:"async"`
}

func (t *SubAgentTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	args := SubAgentArguments{}
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", err
	}

	agentId, _ := ctx.Value("agent_id").(string)
	if agentId == "" {
		return "", errors.New("agent_id not found in context")
	}

	da, err := center.AgentManager.GetAgent(agentId)
	if err != nil {
		// If we can't find the agent via manager, maybe it's not a digital agent but we can still try to get model if it was injected
		slog.Error("failed to get agent from manager", "agent_id", agentId, "err", err)
		return "", fmt.Errorf("failed to get agent: %w", err)
	}

	cm := da.GetModel()

	// Initial message for the subagent is the task description
	messages := []*schema.Message{
		schema.UserMessage(args.TaskDescription),
	}

	taskId, _ := ctx.Value("task_id").(string)
	if taskId == "" {
		taskId = fmt.Sprintf("sub-%s", agentId[:8])
	}

	slog.Info("forking subagent", "parent_id", agentId, "task_id", taskId, "kind", args.Kind)

	resultChan := subagent.ForkSubAgent(ctx, taskId, args.Kind, &subagent.Config{
		ParentID:        agentId,
		ChatModel:       cm,
		Messages:        messages,
		TaskDescription: args.TaskDescription,
	})

	if resultChan == nil {
		return "", errors.New("failed to fork subagent")
	}

	if args.Async {
		return "doing async subagent task", nil
	}

	result := <-resultChan
	if !result.Success {
		return "", fmt.Errorf("subagent execution failed: %v", result.Error)
	}

	return result.Content, nil
}

func init() {
	registry.RegisterTool(&SubAgentTool{})
}
