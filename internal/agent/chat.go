package agent

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"digital-labor/config"
	agentloop "digital-labor/internal/agent/loop"
	"digital-labor/internal/agent/state"
	"digital-labor/internal/agent/task"
	"digital-labor/schema"
)

func (dga *Agent) run(ctx context.Context, req ChatRequest) (chan string, error) {
	dga.runMu.Lock()
	session, err := dga.memory.GetOrCreate()
	if err != nil {
		dga.runMu.Unlock()
		return nil, err
	}
	message, err := buildUserMessage(req)
	if err != nil {
		dga.runMu.Unlock()
		return nil, err
	}
	if err := session.Append(&message); err != nil {
		dga.runMu.Unlock()
		return nil, err
	}

	createdTask := task.Create(task.Config{})
	flowID := createdTask.TaskID
	task.UpdateStatus(flowID, task.StatusRunning)
	if ctx == nil {
		ctx = context.Background()
	}
	runCtx, cancel := context.WithCancel(ctx)
	runCtx = task.WithContext(runCtx, flowID)
	dga.setStop(cancel)

	messages := session.GetCompressMessages()
	inputMessages := make([]schema.Message, 0, len(messages))
	for _, item := range messages {
		if item != nil {
			inputMessages = append(inputMessages, *item)
		}
	}
	output := make(chan string, 30)
	go func() {
		defer close(output)
		defer cancel()
		defer dga.runMu.Unlock()
		defer dga.setStop(nil)

		result, runErr := dga.loop.Run(runCtx, agentloop.Input{
			Messages: inputMessages,
			Stream:   req.IsStream,
			Emit: func(event agentloop.Event) error {
				switch event.Type {
				case agentloop.EventTextDelta, agentloop.EventReasoningDelta:
					if event.Delta == "" {
						return nil
					}
					select {
					case output <- event.Delta:
					case <-runCtx.Done():
						return runCtx.Err()
					}
				}
				return nil
			},
		})
		if runErr != nil {
			task.Update(flowID, task.UpdateConfig{Status: task.StatusFailed, Error: runErr.Error()})
			if !errors.Is(runErr, context.Canceled) {
				output <- runErr.Error()
			}
			return
		}
		for index := range result.Generated {
			msg := result.Generated[index]
			if err := session.Append(&msg); err != nil {
				slog.Error("failed to persist generated message", "error", err)
			}
		}
		task.UpdateStatus(flowID, task.StatusCompleted)
		if !req.IsStream && result.Response != nil && result.Response.Message.Content != "" {
			output <- result.Response.Message.Content
		}
		tokenLimit := config.Conf.Memory.TokenLimit
		if tokenLimit <= 0 {
			tokenLimit = 32000
		}
		if session.Size() > tokenLimit {
			go func() {
				if err := state.Compress(context.Background(), dga.cm, session, dga.state); err != nil {
					slog.Error("memory compression failed", "error", err)
				}
			}()
		}
	}()
	return output, nil
}

func buildUserMessage(req ChatRequest) (schema.Message, error) {
	message := schema.Message{Role: schema.RoleUser, Content: req.Content}
	for _, resource := range req.MultiModalResources {
		part, err := buildContentPart(resource)
		if err != nil {
			return schema.Message{}, err
		}
		message.Parts = append(message.Parts, part)
	}
	return message, nil
}

func buildContentPart(resource MultiModalResource) (schema.ContentPart, error) {
	part := schema.ContentPart{Text: resource.Text, URL: resource.URL, Data: resource.Base64Data, MIMEType: resource.MIMEType, Name: resource.Name, Detail: string(resource.Detail)}
	switch resource.Type {
	case MultiModalResourceTypeText:
		part.Type = schema.ContentText
	case MultiModalResourceTypeImage:
		part.Type = schema.ContentImageURL
	case MultiModalResourceTypeAudio:
		part.Type = schema.ContentInputAudio
	case MultiModalResourceTypeFile:
		part.Type = schema.ContentFile
	case MultiModalResourceTypeVideo:
		return schema.ContentPart{}, fmt.Errorf("chat completions does not support video input")
	default:
		return schema.ContentPart{}, fmt.Errorf("unsupported multimodal resource type %q", resource.Type)
	}
	return part, nil
}
