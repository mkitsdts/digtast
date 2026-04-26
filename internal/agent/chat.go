package agent

import (
	"context"
	mmodel "digital-labor/pkg/model"
	"errors"
	"io"
	"log/slog"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

func (dga *DigitalAgent) run(ctx context.Context, req mmodel.ChatRequest) (chan string, error) {
	sessionID := req.SessionID

	session, err := dga.memory.GetOrCreate(sessionID)
	if err != nil {
		return nil, err
	}
	msgs, err := buildMessages(req.Content, session.GetMessages())
	if err != nil {
		return nil, err
	}

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

func buildMessages(content string, messages []*schema.Message) ([]*schema.Message, error) {
	// TODO:
	msg := &schema.Message{
		Role:    schema.User,
		Content: content,
	}
	messages = append(messages, msg)
	return messages, nil
}
