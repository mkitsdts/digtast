package agent

import (
	"context"
	"digital-labor/pkg/registry"
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
)

// CreateSubAgent creates a new agent with tools filtered by type.
func CreateSubAgent(ctx context.Context, typ string, cm model.ToolCallingChatModel, agentId, name, description string) (*adk.ChatModelAgent, error) {
	tools := registry.GetToolsByType(typ)

	// If no specific tools found for the type, maybe fallback to all tools or return error?
	// Use different tools in different types of sub agents

	ag, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        name,
		Description: description,
		Model:       cm,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: tools,
				ToolCallMiddlewares: []compose.ToolMiddleware{
					{Invokable: registry.Invokable},
				},
			},
		},
		// We might want to pass handlers as well if needed
		Handlers: registry.GetHandlers(),
		ModelRetryConfig: &adk.ModelRetryConfig{
			MaxRetries: 3,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create sub-agent: %w", err)
	}

	return ag, nil
}
