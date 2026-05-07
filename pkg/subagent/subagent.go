package subagent

import (
	"context"
	local "digital-labor/pkg/middleware/lbackend"
	"digital-labor/pkg/model"
	"digital-labor/pkg/registry"
	"digital-labor/pkg/task"
	"fmt"
	"log/slog"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/filesystem"
	"github.com/cloudwego/eino/adk/prebuilt/deep"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type SubAgent struct {
	agent    adk.ResumableAgent
	TaskID   string
	msgs     []*schema.Message
	kind     string
	parentid string
}

func ForkSubAgent(ctx context.Context, taskID, kind string, cfg *Config) chan Result {
	if cfg.TaskDescription == "" {
		resultChan := make(chan Result)
		go func() {
			resultChan <- Result{Error: fmt.Errorf("task description is empty")}
		}()
		return resultChan
	}

	a := SubAgent{}
	if len(cfg.Messages) > 0 {
		a.msgs = make([]*schema.Message, len(cfg.Messages))
		copy(a.msgs, cfg.Messages)
	}
	a.TaskID = taskID
	a.kind = kind
	a.parentid = cfg.ParentID
	a.msgs = append(a.msgs, &schema.Message{
		Role:    schema.User,
		Content: cfg.TaskDescription,
	})

	tools := registry.GetToolsByType(kind)
	var err error

	var backend filesystem.Backend
	var streamingShell filesystem.StreamingShell

	switch kind {
	case "research", "researcher":
		backend = local.GetSecureBackend()
		streamingShell = local.GetSecureBackend()
	case "execution", "executer":
		backend = local.GetBackend()
		streamingShell = local.GetBackend()
	default:
		backend = local.GetSecureBackend()
		streamingShell = local.GetSecureBackend()
	}

	a.agent, err = deep.New(ctx, &deep.Config{
		Name:      fmt.Sprintf("subagent-%s", taskID),
		ChatModel: cfg.ChatModel,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: tools,
			},
		},
		Backend:        backend,
		StreamingShell: streamingShell,
		Handlers:       registry.GetHandlers(),
		ModelRetryConfig: &adk.ModelRetryConfig{
			MaxRetries: 5,
		},
		WithoutWriteTodos: true,
	})

	if err != nil {
		slog.Error("failed to create subagent", "err", err)
		return nil
	}
	return a.Run(ctx)
}

func (a *SubAgent) Run(ctx context.Context) chan Result {
	result := make(chan Result, 1)

	go func() {
		defer close(result)

		var cancel context.CancelFunc
		runCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		events := a.agent.Run(runCtx, &adk.AgentInput{
			Messages:        a.msgs,
			EnableStreaming: false,
		})

		content := ""
		if events == nil {
			slog.Error("events stream is nil", "agent_id", a.parentid, "task_id", a.TaskID)
			task.UpdateStatus(a.parentid, a.TaskID, model.TaskStatusFailed)
			result <- Result{Error: fmt.Errorf("events stream is nil"), Success: false}
			return
		}

		for {
			event, ok := events.Next()

			// event stop
			if !ok {
				task.UpdateStatus(a.parentid, a.TaskID, model.TaskStatusCompleted)
				result <- Result{Summary: "", Content: content, MultiContent: nil, Error: nil, Success: true}
				return
			}

			// event error
			if event.Err != nil {
				slog.Error("execute subagent failed", "agent_id", a.parentid, "task_id", a.TaskID, "err", event.Err)
				task.UpdateTask(a.parentid, a.TaskID, task.UpdateConfig{
					Status: model.TaskStatusFailed,
					Error:  event.Err.Error(),
				})
				result <- Result{Summary: "", Content: content, MultiContent: nil, Error: event.Err, Success: false}
				return
			}

			// event no output
			if event.Output == nil {
				continue
			}

			mv := event.Output.MessageOutput
			if mv == nil {
				continue
			}

			if mv.Role == schema.Tool {
				slog.Info("subagent use tool", "tool", mv.ToolName, "task_id", a.TaskID)
			}

			if mv.Message != nil && mv.Message.Content != "" {
				content += mv.Message.Content
			}
		}
	}()

	return result
}
