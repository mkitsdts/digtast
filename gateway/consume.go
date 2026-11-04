package gateway

import (
	"context"
	"encoding/base64"
	"log/slog"
	"mime"
	"path/filepath"
	"strings"

	cgw "digital-labor/component/gateway"
	"digital-labor/internal/agent"
)

// Start 启动用户消息消费循环，把网关收到的消息交给 Agent 处理。
func Start() {
	go ConsumeMessage()
}

func ConsumeMessage() {
	for {
		msg, ok := cgw.PopUserMessage()
		if !ok {
			continue
		}
		consume(msg)
	}
}

func consume(msg *cgw.UserMessage) {
	slog.Info("consume message", "notifyWay", msg.NotifyWay, "content", msg.Content)
	channel := cgw.NewChannel(msg.NotifyWay)
	if channel == nil {
		slog.Error("channel not exist", "notifyWay", msg.NotifyWay)
		return
	}

	var result cgw.Result
	ag, err := agent.Instance()
	if err != nil {
		result.Success = false
		result.Error = err
		sendResult(channel, result, msg)
		return
	}

	sm, err := ag.Run(context.Background(), msg.Content, multiContentToResources(msg.MultiContent), false)
	if err != nil {
		result.Success = false
		result.Error = err
		sendResult(channel, result, msg)
		slog.Error("failed to run agent", "error", err)
		return
	}

	result.Content = <-sm
	result.Success = true
	sendResult(channel, result, msg)
}

func sendResult(channel cgw.MessageChannel, result cgw.Result, msg *cgw.UserMessage) {
	if err := channel.Send(result, msg.Params); err != nil {
		slog.Error("failed to send message", "error", err)
	}
	slog.Info("agent notify success", "content", msg.Content)
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
