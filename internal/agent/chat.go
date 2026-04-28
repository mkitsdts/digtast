package agent

import (
	"context"
	mmodel "digital-labor/pkg/model"
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
	sessionID := req.SessionID

	session, err := dga.memory.GetOrCreate(sessionID)
	if err != nil {
		return nil, err
	}

	// Persist the user's turn through internal memory.
	if err := session.Append(&schema.Message{
		Role:    schema.User,
		Content: req.Content,
	}); err != nil {
		return nil, err
	}

	msgs := buildMessages(session.GetMessages())

	runCtx, cancel := context.WithCancel(ctx)
	dga.bindRun(sessionID, cancel)

	events := dga.agent.Run(runCtx, &adk.AgentInput{
		Messages:        msgs,
		EnableStreaming: req.IsStream,
	})

	ch := make(chan string, 30)
	go func() {
		defer close(ch)
		defer dga.unbindRun(sessionID)
		defer func() {
			// Chunk rotation is intentionally deferred until the run finishes so
			// one request/response pair is not split across two files.
			if err := session.CompleteTurn(); err != nil {
				slog.Error("failed to complete memory turn", "session_id", sessionID, "err", err)
			}
		}()

		if events == nil {
			slog.Error("events stream is nil")
			return
		}

		runResult := ""

		for {
			if runCtx.Err() != nil {
				slog.Warn("agent execution cancelled or timeout", "session_id", sessionID, "err", runCtx.Err())
				// TODO: cancel agent loop
				return
			}

			select {
			case <-runCtx.Done():
				cancel()
			default:
			}

			event, ok := events.Next()

			// loop end
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
				slog.Error("execute agent failed", "session_id", sessionID, "err", event.Err)
				return
			}

			mv := event.Output.MessageOutput
			if mv == nil {
				continue
			}

			if mv.Role == schema.Tool {
				// TODO:写入调用 tool 的日志
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

func (dga *DigitalAgent) stop(sessionId string) error {
	dga.runMu.Lock()
	cancel, ok := dga.runStops[sessionId]
	dga.runMu.Unlock()
	if !ok {
		return errors.New("session is not running")
	}

	cancel()
	return nil
}

func buildMessages(messages []*schema.Message) []*schema.Message {
	// Gather all registered prompts from workspace
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

	// If the first message is already a system message, don't inject again.
	if len(messages) > 0 && messages[0].Role == schema.System {
		return messages
	}

	sysMsg := &schema.Message{
		Role:    schema.System,
		Content: fullSystemPrompt,
	}

	// Prepend system message
	return append([]*schema.Message{sysMsg}, messages...)
}
