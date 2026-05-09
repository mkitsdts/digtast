package gateway

import (
	"context"
	"digital-labor/internal/center"
	"digital-labor/pkg/gateway"
	"digital-labor/pkg/model"
	"digital-labor/pkg/queue"
	"encoding/base64"
	"log/slog"
	"mime"
	"path/filepath"
	"strings"
)

var que *queue.Queue[*UserMessage] = queue.NewQueue[*UserMessage]()

func PushUserMessage(msg *UserMessage) {
	que.Push(msg)
}

func Start() {
	go ConsumeMessage()
}

func ConsumeMessage() {
	for {
		msg, ok := que.Pop()
		var result model.Result
		if !ok {
			continue
		}
		slog.Info("consume message", "notifyWay", msg.NotifyWay, "content", msg.Content)
		channel := gateway.NewChannel(msg.NotifyWay)
		if channel == nil {
			slog.Error("channel not exist", "notifyWay", msg.NotifyWay)
			continue
		}

		agent, err := center.AgentManager.GetDefaultAgent()
		if err != nil {
			slog.Error("failed to get agent", "error", err)
			result.Success = false
			result.Error = err
			if err := channel.Send(result, msg.Params); err != nil {
				slog.Error("failed to send message", "error", err)
			}
			slog.Info("agent notify success", "content", msg.Content)
			return
		}

		sm, err := agent.Run(context.Background(), msg.Content, multiContentToResources(msg.MultiContent), false)
		if err != nil {
			result.Success = false
			result.Error = err
			if err := channel.Send(result, msg.Params); err != nil {
				slog.Error("failed to send message", "error", err)
			}
			slog.Info("agent notify success", "content", msg.Content)
			slog.Error("failed to run agent", "error", err)
			continue
		}

		result.Content = <-sm
		result.Success = true
		if err := channel.Send(result, msg.Params); err != nil {
			slog.Error("failed to send message", "error", err)
		}
		slog.Info("agent notify success", "content", msg.Content)
	}
}

func multiContentToResources(content map[string][]byte) []model.MultiModalResource {
	if len(content) == 0 {
		return nil
	}

	resources := make([]model.MultiModalResource, 0, len(content))
	for name, data := range content {
		mimeType := mime.TypeByExtension(filepath.Ext(name))
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
		resources = append(resources, model.MultiModalResource{
			Type:       resourceTypeFromMIME(mimeType),
			Base64Data: base64.StdEncoding.EncodeToString(data),
			MIMEType:   mimeType,
			Name:       name,
		})
	}
	return resources
}

func resourceTypeFromMIME(mimeType string) model.MultiModalResourceType {
	switch {
	case strings.HasPrefix(mimeType, "image/"):
		return model.MultiModalResourceTypeImage
	case strings.HasPrefix(mimeType, "audio/"):
		return model.MultiModalResourceTypeAudio
	case strings.HasPrefix(mimeType, "video/"):
		return model.MultiModalResourceTypeVideo
	default:
		return model.MultiModalResourceTypeFile
	}
}
