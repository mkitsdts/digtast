package agent

import (
	"context"
	mmodel "digital-labor/pkg/model"
	"errors"
	"log/slog"
)

func (dga *DigitalAgent) run(ctx context.Context, req mmodel.ChatRequest) (chan string, error) {
	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = sessionIDFromContext(ctx)
	}
	if sessionID == "" {
		return nil, errors.New("invalid session_id")
	}

	msgs, err := buildMessages(req.Content, dga.prompts.Build(req.Prompt))
	if err != nil {
		return nil, err
	}

	runCtx, cancel := context.WithCancel(ctx)
	dga.bindRun(sessionID, cancel)
	events := dga.agent.Run(runCtx, msgs)

	ch := make(chan string, 30)
	go func() {
		defer close(ch)
		defer dga.unbindRun(sessionID)

		runResult := ""
		for {
			event, ok := events.Next()
			if !ok {
				if !req.IsStream && runResult != "" {
					ch <- runResult
				}
				return
			}
			if event.Err != nil {
				slog.Error("execute agent failed", "session_id", sessionID, "err", event.Err)
				return
			}
			if msg, err := event.Output.MessageOutput.GetMessage(); err == nil && msg.Content != "" {
				runResult += msg.Content
				if req.IsStream {
					ch <- msg.Content
				}
			}
		}
	}()

	return ch, nil
}
