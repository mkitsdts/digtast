package subagent

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"digital-labor/pkg/agentloop"
	"digital-labor/pkg/filesystem"
	"digital-labor/pkg/llm"
	local "digital-labor/pkg/middleware/lbackend"
	"digital-labor/pkg/model"
	"digital-labor/pkg/registry"
	"digital-labor/pkg/task"
	"digital-labor/pkg/tool"
)

type SubAgent struct {
	loop     *agentloop.Loop
	TaskID   string
	msgs     []llm.Message
	parentID string
}

func ForkSubAgent(ctx context.Context, taskID, kind string, cfg *Config) chan Result {
	result := make(chan Result, 1)
	if cfg == nil || cfg.ChatModel == nil {
		result <- Result{Error: errors.New("subagent config and model are required")}
		close(result)
		return result
	}
	if cfg.TaskDescription == "" {
		result <- Result{Error: errors.New("task description is empty")}
		close(result)
		return result
	}
	var backend filesystem.Backend = local.GetSecureBackend()
	var shell filesystem.StreamingShell = local.GetSecureBackend()
	if kind == "execution" || kind == "executer" {
		backend = local.GetBackend()
		shell = local.GetBackend()
	}
	files := filesystem.ToolLoader{Backend: backend, Shell: shell}
	if kind == "research" || kind == "researcher" {
		files.ReadOnly = true
	}
	loop, err := agentloop.New(agentloop.Config{
		LLM:         cfg.ChatModel,
		ToolLoaders: []tool.Loader{tool.StaticLoader(registry.GetToolsByType(kind)), files},
		Hooks:       registry.GetHandlers(), MaxSteps: 20, MaxModelRetries: 5,
	})
	if err != nil {
		result <- Result{Error: err}
		close(result)
		return result
	}
	messages := append([]llm.Message(nil), cfg.Messages...)
	if len(messages) == 0 || messages[len(messages)-1].Content != cfg.TaskDescription {
		messages = append(messages, llm.Message{Role: llm.RoleUser, Content: cfg.TaskDescription})
	}
	agent := &SubAgent{loop: loop, TaskID: taskID, msgs: messages, parentID: cfg.ParentID}
	return agent.Run(ctx)
}

func (a *SubAgent) Run(ctx context.Context) chan Result {
	result := make(chan Result, 1)
	go func() {
		defer close(result)
		runResult, err := a.loop.Run(ctx, agentloop.Input{Messages: a.msgs})
		if err != nil {
			slog.Error("execute subagent failed", "agent_id", a.parentID, "task_id", a.TaskID, "error", err)
			task.UpdateTask(a.parentID, a.TaskID, task.UpdateConfig{Status: model.TaskStatusFailed, Error: err.Error()})
			result <- Result{Error: err}
			return
		}
		if runResult.Response == nil {
			err := fmt.Errorf("subagent returned no response")
			result <- Result{Error: err}
			return
		}
		task.UpdateStatus(a.parentID, a.TaskID, model.TaskStatusCompleted)
		result <- Result{Content: runResult.Response.Message.Content, Success: true}
	}()
	return result
}
