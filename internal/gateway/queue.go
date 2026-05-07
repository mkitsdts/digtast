package gateway

import (
	"context"
	"digital-labor/internal/center"
	"digital-labor/pkg/gateway"
	"digital-labor/pkg/model"
	"digital-labor/pkg/queue"
	"log/slog"
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

		sm, err := agent.Run(context.Background(), msg.Content, false)
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
