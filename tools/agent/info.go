package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"digital-labor/internal/agent"
	"digital-labor/internal/center"
	pkgagent "digital-labor/pkg/agent"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"github.com/eino-contrib/jsonschema"
	orderedmap "github.com/wk8/go-ordered-map/v2"
)

type SubAgentTool struct {
}

const (
	SubAgentTypeResearch  = "research"
	SubAgentTypeExecution = "execution"
)

func (t *SubAgentTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	properties := orderedmap.New[string, *jsonschema.Schema]()
	properties.Set("task", &jsonschema.Schema{
		Type:        "string",
		Description: "需要子智能体执行的任务描述。",
	})
	properties.Set("subagent_type", &jsonschema.Schema{
		Type:        "string",
		Description: "子智能体的类型，目前支持：research (研究/搜索型), execution (执行/任务型)。",
		Enum: []any{
			SubAgentTypeResearch,
			SubAgentTypeExecution,
		},
	})

	jsonschema := &jsonschema.Schema{
		Version:    "draft",
		ID:         "sub_agent_tool",
		Type:       "object",
		Properties: properties,
		Required:   []string{"task", "subagent_type"},
	}

	return &schema.ToolInfo{
		Name:        "sub_agent",
		Desc:        "创建并运行一个子智能体来协助处理复杂的任务。你可以根据任务性质选择 'research' 或 'execution' 类型。",
		ParamsOneOf: schema.NewParamsOneOfByJSONSchema(jsonschema),
	}, nil
}

type SubAgentArguments struct {
	Task         string `json:"task"`
	SubAgentType string `json:"subagent_type"`
}

func (t *SubAgentTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	args := SubAgentArguments{}
	if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
		return "", err
	}

	if center.AgentManager == nil {
		return "", fmt.Errorf("agent manager not initialized")
	}

	// Get the current agent to use its model
	// We need to get the agent_id from context
	agentID, ok := ctx.Value("agent_id").(string)
	if !ok {
		return "", fmt.Errorf("agent_id not found in context")
	}

	dga, err := center.AgentManager.GetAgent(agentID)
	if err != nil {
		return "", fmt.Errorf("failed to get current agent: %w", err)
	}

	// Determine agent description based on type
	description := ""
	name := ""
	switch args.SubAgentType {
	case SubAgentTypeResearch:
		name = "ResearchSubAgent"
		description = "专注于信息检索、研究和分析的子智能体。你应该利用搜索工具、浏览器等手段，为用户提供准确、详尽的研究报告和背景资料。"
	case SubAgentTypeExecution:
		name = "ExecutionSubAgent"
		description = "专注于任务执行和行动的子智能体。你应该根据任务要求，调用相关工具完成具体操作，并反馈执行结果。"
	default:
		name = "GenericSubAgent"
		description = "通用的子智能体，协助主智能体完成各项任务。"
	}

	// Use the new pkg/agent logic to create the sub-agent
	// We pass the model from the current agent
	subAgent, err := pkgagent.CreateSubAgent(ctx, args.SubAgentType, dga.GetModel(), agentID, name, description)
	if err != nil {
		return "", fmt.Errorf("failed to create sub-agent: %w", err)
	}

	// Run the agent
	// We need a way to run the adk.ChatModelAgent directly.
	// adk.ChatModelAgent.Run returns a stream of events.
	mainAgent, err := agent.GetManager().GetAgent(agentID)
	if err != nil {
		return "", fmt.Errorf("failed to get main agent: %w", err)
	}
	session, err := mainAgent.GetSession()
	if err != nil {
		return "", fmt.Errorf("failed to get session: %w", err)
	}
	msgs := make([]*schema.Message, 0, len(session.GetMessages())+1)
	copy(msgs, session.GetMessages())
	msgs = append(msgs, &schema.Message{
		Role:    schema.User,
		Content: args.Task,
	})

	events := subAgent.Run(ctx, &adk.AgentInput{
		Messages: msgs,
	})

	var result strings.Builder
	for {
		event, ok := events.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			return "", fmt.Errorf("sub-agent run error: %w", event.Err)
		}
		if event.Output != nil && event.Output.MessageOutput != nil && event.Output.MessageOutput.Message != nil {
			result.WriteString(event.Output.MessageOutput.Message.Content)
		}
	}

	return result.String(), nil
}
