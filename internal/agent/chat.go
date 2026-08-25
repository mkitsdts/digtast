package agent

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"digital-labor/pkg/agentloop"
	"digital-labor/pkg/conf"
	"digital-labor/pkg/llm"
	mmodel "digital-labor/pkg/model"
	"digital-labor/pkg/state"
	"digital-labor/pkg/task"
)

func (dga *DigitalAgent) run(ctx context.Context, req mmodel.ChatRequest) (chan string, error) {
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

	createdTask, err := task.CreateTask(dga.ID, task.Config{AgentID: dga.ID})
	if err != nil {
		dga.runMu.Unlock()
		return nil, err
	}
	flowID := createdTask.TaskID
	task.UpdateStatus(dga.ID, flowID, mmodel.TaskStatusRunning)
	if ctx == nil {
		ctx = context.Background()
	}
	runCtx, cancel := context.WithCancel(ctx)
	runCtx = context.WithValue(runCtx, "task_id", flowID)
	runCtx = context.WithValue(runCtx, "agent_id", dga.ID)
	dga.setStop(cancel)

	messages := session.GetCompressMessages()
	inputMessages := make([]llm.Message, 0, len(messages))
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
				if event.Type == agentloop.EventTextDelta && event.Delta != "" {
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
			task.UpdateTask(dga.ID, flowID, task.UpdateConfig{Status: mmodel.TaskStatusFailed, Error: runErr.Error()})
			if !errors.Is(runErr, context.Canceled) {
				output <- runErr.Error()
			}
			return
		}
		for index := range result.Generated {
			msg := result.Generated[index]
			if err := session.Append(&msg); err != nil {
				slog.Error("failed to persist generated message", "agent_id", dga.ID, "error", err)
			}
		}
		task.UpdateStatus(dga.ID, flowID, mmodel.TaskStatusCompleted)
		if !req.IsStream && result.Response != nil && result.Response.Message.Content != "" {
			output <- result.Response.Message.Content
		}
		tokenLimit := conf.Conf.Memory.TokenLimit
		if tokenLimit <= 0 {
			tokenLimit = 32000
		}
		if session.Size() > tokenLimit {
			go func() {
				if err := state.Compress(context.Background(), dga.cm, session, dga.state); err != nil {
					slog.Error("memory compression failed", "agent_id", dga.ID, "error", err)
				}
			}()
		}
	}()
	return output, nil
}

func buildUserMessage(req mmodel.ChatRequest) (llm.Message, error) {
	message := llm.Message{Role: llm.RoleUser, Content: req.Content}
	for _, resource := range req.MultiModalResources {
		part, err := buildContentPart(resource)
		if err != nil {
			return llm.Message{}, err
		}
		message.Parts = append(message.Parts, part)
	}
	return message, nil
}

func buildContentPart(resource mmodel.MultiModalResource) (llm.ContentPart, error) {
	part := llm.ContentPart{Text: resource.Text, URL: resource.URL, Data: resource.Base64Data, MIMEType: resource.MIMEType, Name: resource.Name, Detail: string(resource.Detail)}
	switch resource.Type {
	case mmodel.MultiModalResourceTypeText:
		part.Type = llm.ContentText
	case mmodel.MultiModalResourceTypeImage:
		part.Type = llm.ContentImageURL
	case mmodel.MultiModalResourceTypeAudio:
		part.Type = llm.ContentInputAudio
	case mmodel.MultiModalResourceTypeFile:
		part.Type = llm.ContentFile
	case mmodel.MultiModalResourceTypeVideo:
		return llm.ContentPart{}, fmt.Errorf("chat completions does not support video input")
	default:
		return llm.ContentPart{}, fmt.Errorf("unsupported multimodal resource type %q", resource.Type)
	}
	return part, nil
}
