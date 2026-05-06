package gateway

import (
	"context"
	"digital-labor/internal/center"
	"digital-labor/pkg/gateway"
	"digital-labor/pkg/model"
	"digital-labor/pkg/queue"
	"log/slog"
)

var que *queue.Queue[*UserMessage] = &queue.Queue[*UserMessage]{}

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

		channel := gateway.NewChannel(msg.NotifyWay)
		if channel == nil {
			slog.Error("channel not exist", "notifyWay", msg.NotifyWay)
			continue
		}
		defer func() {
			// send result to user
			if err := channel.Send(result, msg.Params); err != nil {
				slog.Error("failed to send message", "error", err)
			}
		}()

		agent, err := center.AgentManager.GetDefaultAgent()
		if err != nil {
			slog.Error("failed to get agent", "error", err)
			return
		}

		sm, err := agent.Run(context.Background(), msg.Content, false)
		if err != nil {
			result.Success = false
			result.Error = err
			continue
		}

		result.Content = <-sm
		result.Success = true

	}
}
