package gateway

import (
	"context"
	"digital-labor/internal/agent"
	"digital-labor/pkg/gateway"
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
		var result gateway.Result
		if !ok {
			continue
		}
		slog.Info("consume message", "notifyWay", msg.NotifyWay, "content", msg.Content)
		channel := gateway.NewChannel(msg.NotifyWay)
		if channel == nil {
			slog.Error("channel not exist", "notifyWay", msg.NotifyWay)
			continue
		}

		ag, err := agent.Instance()
		if err != nil {
			slog.Error("failed to get agent", "error", err)
			result.Success = false
			result.Error = err
			if err := channel.Send(result, msg.Params); err != nil {
				slog.Error("failed to send message", "error", err)
			}
			slog.Info("agent notify success", "content", msg.Content)
			continue
		}

		sm, err := ag.Run(context.Background(), msg.Content, multiContentToResources(msg.MultiContent), false)
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

func multiContentToResources(content map[string][]byte) []agent.MultiModalResource {
	if len(content) == 0 {
		return nil
	}

	resources := make([]agent.MultiModalResource, 0, len(content))
	for name, data := range content {
		mimeType := mime.TypeByExtension(filepath.Ext(name))
		if mimeType == "" {
			mimeType = "application/octet-stream"
		}
		resources = append(resources, agent.MultiModalResource{
			Type:       resourceTypeFromMIME(mimeType),
			Base64Data: base64.StdEncoding.EncodeToString(data),
			MIMEType:   mimeType,
			Name:       name,
		})
	}
	return resources
}

func resourceTypeFromMIME(mimeType string) agent.MultiModalResourceType {
	switch {
	case strings.HasPrefix(mimeType, "image/"):
		return agent.MultiModalResourceTypeImage
	case strings.HasPrefix(mimeType, "audio/"):
		return agent.MultiModalResourceTypeAudio
	case strings.HasPrefix(mimeType, "video/"):
		return agent.MultiModalResourceTypeVideo
	default:
		return agent.MultiModalResourceTypeFile
	}
}
