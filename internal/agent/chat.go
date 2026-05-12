package agent

import (
	"context"
	"digital-labor/pkg/conf"
	"digital-labor/pkg/ctxmanager"
	mmodel "digital-labor/pkg/model"
	"digital-labor/pkg/state"
	"digital-labor/pkg/task"
	"errors"
	"io"
	"log/slog"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

func (dga *DigitalAgent) run(ctx context.Context, req mmodel.ChatRequest) (chan string, error) {
	session, err := dga.memory.GetOrCreate()
	if err != nil {
		return nil, err
	}

	msg, err := buildUserMessage(req)
	if err != nil {
		return nil, err
	}

	if err := session.Append(msg); err != nil {
		return nil, err
	}

	sysMsg := &schema.Message{Role: schema.System, Content: dga.state.Build()}
	msgs := append([]*schema.Message{sysMsg}, session.GetPromptMessages()...)

	t, err := task.CreateTask(dga.ID, task.Config{
		AgentID: dga.ID,
	})
	if err != nil {
		return nil, err
	}
	flowID := t.TaskID
	task.UpdateStatus(dga.ID, flowID, mmodel.TaskStatusRunning)

	runCtx, cancel := context.WithCancel(ctxmanager.GetOrCreate(flowID))
	runCtx = context.WithValue(runCtx, "task_id", flowID)
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
			task.UpdateStatus(dga.ID, flowID, mmodel.TaskStatusFailed)
			return
		}

		runResult := ""
		var runErr error

		for {
			if runCtx.Err() != nil {
				slog.Warn("agent execution cancelled or timeout", "agent_id", dga.ID, "err", runCtx.Err())
				task.UpdateStatus(dga.ID, flowID, mmodel.TaskStatusFailed)
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

				if runErr != nil {
					task.UpdateTask(dga.ID, flowID, task.UpdateConfig{
						Status: mmodel.TaskStatusFailed,
						Error:  runErr.Error(),
					})
				} else {
					task.UpdateStatus(dga.ID, flowID, mmodel.TaskStatusCompleted)
				}

				// Trigger memory compression if session exceeds token limit
				tokenLimit := conf.Conf.Memory.TokenLimit
				if tokenLimit <= 0 {
					tokenLimit = 32000
				}
				if session.Size() > tokenLimit {
					go func() {
						if err := state.Compress(runCtx, dga.cm, session, dga.state); err != nil {
							slog.Error("memory compression failed", "agent_id", dga.ID, "err", err)
						}
					}()
				}

				if !req.IsStream && runResult != "" {
					ch <- runResult
				}
				return
			}

			// maybe need to handle tool calls fail result
			if event.Err != nil {
				slog.Error("execute agent failed", "agent_id", dga.ID, "err", event.Err)
				runErr = event.Err
				ch <- event.Err.Error()
				continue
			}

			if event.Output == nil {
				slog.Error("event output is nil", "agent_id", dga.ID, "event", event)
				continue
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

func buildUserMessage(req mmodel.ChatRequest) (*schema.Message, error) {
	msg := new(schema.Message)

	msg.Role = schema.User
	msg.Content = req.Content

	arts, err := buildMessageInputParts(req.Content, req.MultiModalResources)
	if err != nil {
		return nil, err
	}
	msg.UserInputMultiContent = arts

	return msg, nil
}

func buildMessageInputParts(content string, resources []mmodel.MultiModalResource) ([]schema.MessageInputPart, error) {
	if len(resources) == 0 {
		return nil, nil
	}

	parts := make([]schema.MessageInputPart, 0, len(resources)+1)
	if content != "" {
		parts = append(parts, schema.MessageInputPart{
			Type: schema.ChatMessagePartTypeText,
			Text: content,
		})
	}

	for _, resource := range resources {
		part, err := buildMessageInputPart(resource)
		if err != nil {
			return nil, err
		}
		parts = append(parts, part)
	}

	return parts, nil
}

func buildMessageInputPart(resource mmodel.MultiModalResource) (schema.MessageInputPart, error) {
	switch resource.Type {
	case mmodel.MultiModalResourceTypeText:
		return schema.MessageInputPart{
			Type:  schema.ChatMessagePartTypeText,
			Text:  resource.Text,
			Extra: resource.Extra,
		}, nil
	case mmodel.MultiModalResourceTypeImage:
		return schema.MessageInputPart{
			Type: schema.ChatMessagePartTypeImageURL,
			Image: &schema.MessageInputImage{
				MessagePartCommon: messagePartCommon(resource),
				Detail:            schema.ImageURLDetail(resource.Detail),
			},
			Extra: resource.Extra,
		}, nil
	case mmodel.MultiModalResourceTypeAudio:
		return schema.MessageInputPart{
			Type:  schema.ChatMessagePartTypeAudioURL,
			Audio: &schema.MessageInputAudio{MessagePartCommon: messagePartCommon(resource)},
			Extra: resource.Extra,
		}, nil
	case mmodel.MultiModalResourceTypeVideo:
		return schema.MessageInputPart{
			Type:  schema.ChatMessagePartTypeVideoURL,
			Video: &schema.MessageInputVideo{MessagePartCommon: messagePartCommon(resource)},
			Extra: resource.Extra,
		}, nil
	case mmodel.MultiModalResourceTypeFile:
		return schema.MessageInputPart{
			Type: schema.ChatMessagePartTypeFileURL,
			File: &schema.MessageInputFile{
				MessagePartCommon: messagePartCommon(resource),
				Name:              resource.Name,
			},
			Extra: resource.Extra,
		}, nil
	default:
		return schema.MessageInputPart{}, errors.New("unsupported multimodal resource type")
	}
}

func messagePartCommon(resource mmodel.MultiModalResource) schema.MessagePartCommon {
	common := schema.MessagePartCommon{
		MIMEType: resource.MIMEType,
		Extra:    resource.Extra,
	}

	if resource.URL != "" {
		common.URL = &resource.URL
	}
	if resource.Base64Data != "" {
		common.Base64Data = &resource.Base64Data
	}

	return common
}

func (dga *DigitalAgent) stop() error {
	if dga.runStops != nil {
		dga.runStops()
		dga.runStops = nil
		return nil
	}
	return errors.New("task not exist")
}
