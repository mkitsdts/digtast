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
	if sessionID == "" {
		sessionID = sessionIDFromContext(ctx)
	}
	if sessionID == "" {
		return nil, errors.New("invalid session_id")
	}

	session, err := dga.memory.GetOrCreate(sessionID)
	if err != nil {
		return nil, err
	}
	msgs := session.GetMessages()
	buildMessages(req.Content, msgs)

	runCtx, cancel := context.WithCancel(ctx)
	dga.bindRun(sessionID, cancel)
	events := dga.agent.Run(runCtx, &adk.AgentInput{
		Messages:        msgs,
		EnableStreaming: req.IsStream,
	})

	ch := make(chan string, 30)
	go func() {
		defer close(ch)

		if dga == nil || session == nil {
			slog.Error("agent or session is nil")
			return
		}
		defer dga.unbindRun(sessionID)

		runResult := ""
		for {

			if events == nil {
				slog.Error("events stream is nil")
				return
			}

			event, ok := events.Next()
			if !ok {
				session.Append(&schema.Message{
					Role:    schema.Assistant,
					Content: runResult,
				})
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
			if mv.Role != schema.Assistant {
				continue
			}

			if req.IsStream {
				mv.MessageStream.SetAutomaticClose()
				for {
					frame, err := mv.MessageStream.Recv()
					if errors.Is(err, io.EOF) {
						break
					}
					if err != nil {
						return
					}
					if frame != nil && frame.Content != "" {
						ch <- frame.Content
					}
				}
				continue
			}

			if mv.Message != nil {
				ch <- mv.Message.Content
			}

			// if event.Output != nil && event.Output.MessageOutput != nil {
			// 	msg, err := event.Output.MessageOutput.GetMessage()
			// 	if err == nil && msg != nil && msg.Content != "" {
			// 		runResult += msg.Content
			// 		if req.IsStream {
			// 			ch <- msg.Content
			// 		}
			// 	}

			// 	if event.Output.MessageOutput.ToolName != "" {
			// 		session.Append(&schema.Message{
			// 			Role:    schema.Assistant,
			// 			Content: event.Output.MessageOutput.ToolName,
			// 		})
			// 	}
			// }
		}
	}()

	return ch, nil
}

func buildMessages(content string, messages []*schema.Message) error {
	// TODO:

	return nil
}

func sessionIDFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	sessionID, _ := ctx.Value("session_id").(string)
	return sessionID
}
