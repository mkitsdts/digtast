package agent

import (
	"context"
	"digital-labor/pkg/ctxmanager"
	mmodel "digital-labor/pkg/model"
	"digital-labor/pkg/task"
	"digital-labor/pkg/workspace"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

func (dga *DigitalAgent) run(ctx context.Context, req mmodel.ChatRequest) (chan string, error) {
	session, err := dga.memory.GetOrCreate()
	if err != nil {
		return nil, err
	}

	if err := session.Append(&schema.Message{
		Role:    schema.User,
		Content: req.Content,
	}); err != nil {
		return nil, err
	}

	msgs := buildMessages(session.GetMessages())

	t, err := task.CreateTask(dga.ID, task.Config{
		AgentID: dga.ID,
	})
	if err != nil {
		return nil, err
	}

	runCtx, cancel := context.WithCancel(ctxmanager.GetOrCreate(t.TaskID))
	runCtx = context.WithValue(runCtx, "task_id", t.TaskID)
	runCtx = context.WithValue(runCtx, "agent_id", dga.ID)
	dga.runStops = cancel

	dga.runMu.Lock()
	events := dga.agent.Run(runCtx, &adk.AgentInput{
		Messages:        msgs,
		EnableStreaming: req.IsStream,
	})

	ch := make(chan string, 30)
	go func() {
		defer func() {
			dga.runStops = nil
			dga.runMu.Unlock()
		}()
		defer close(ch)
		defer func() {
			if err := session.CompleteTurn(); err != nil {
				slog.Error("failed to complete memory turn", "agent_id", dga.ID, "err", err)
			}
		}()

		if events == nil {
			slog.Error("events stream is nil")
			return
		}

		runResult := ""

		for {
			if runCtx.Err() != nil {
				slog.Warn("agent execution cancelled or timeout", "agent_id", dga.ID, "err", runCtx.Err())
				return
			}

			select {
			case <-runCtx.Done():
				cancel()
			default:
			}

			event, ok := events.Next()

			if !ok {
				if runResult != "" {
					session.Append(&schema.Message{
						Role:    schema.Assistant,
						Content: runResult,
					})
				}

				if !req.IsStream && runResult != "" {
					ch <- runResult
				}
				return
			}

			if event.Err != nil {
				slog.Error("execute agent failed", "agent_id", dga.ID, "err", event.Err)
				return
			}

			mv := event.Output.MessageOutput
			if mv == nil {
				continue
			}

			if mv.Role == schema.Tool {
				session.Append(mv.Message)
				slog.Info("llm use tool", "tool", mv.ToolName)
			}

			if req.IsStream {
				if mv.MessageStream != nil {
					mv.MessageStream.SetAutomaticClose()
					for {
						frame, err := mv.MessageStream.Recv()
						if errors.Is(err, io.EOF) {
							break
						}
						if err != nil {
							slog.Error("stream recv err", "err", err)
							return
						}
						if frame != nil && frame.Content != "" {
							runResult += frame.Content
							ch <- frame.Content
						}
					}
				}
				continue
			}

			if mv.Message != nil && mv.Message.Content != "" {
				runResult += mv.Message.Content
			}
		}
	}()

	return ch, nil
}

func (dga *DigitalAgent) stop() error {
	if dga.runStops != nil {
		dga.runStops()
		dga.runStops = nil
		return nil
	}
	return errors.New("task not exist")
}

func buildMessages(messages []*schema.Message) []*schema.Message {
	promptCreators := workspace.GetPromptCreators()
	var systemPrompts []string

	for _, creator := range promptCreators {
		content, err := creator.GetPromptImpl()
		if err != nil {
			slog.Error("failed to get prompt content", "name", creator.GetPromptName(), "error", err)
			continue
		}
		if content != "" {
			systemPrompts = append(systemPrompts, fmt.Sprintf("### %s\n%s", creator.GetPromptName(), content))
		}
	}

	if len(systemPrompts) == 0 {
		return messages
	}

	fullSystemPrompt := strings.Join(systemPrompts, "\n\n")

	if len(messages) > 0 && messages[0].Role == schema.System {
		return messages
	}

	sysMsg := &schema.Message{
		Role:    schema.System,
		Content: fullSystemPrompt,
	}

	return append([]*schema.Message{sysMsg}, messages...)
}
