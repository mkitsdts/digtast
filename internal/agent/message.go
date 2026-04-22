package agent

import (
	"context"
	"errors"
	"log/slog"

	"github.com/cloudwego/eino/schema"
)

func (dga *DigitalAgent) SendMessageToSession(ctx context.Context, content string, is_stream bool) (chan string, error) {
	sessionId := ctx.Value("session_id").(string)
	if sessionId == "" {
		return nil, errors.New("invalid session_id")
	}

	// 获取对话记录
	mem, err := dga.sessions.GetOrCreate(sessionId)
	if err != nil {
		return nil, err
	}

	// 添加用户消息
	msgs := mem.GetMessages()
	msg := &schema.Message{Role: "user", Content: content}
	msgs = append(msgs, msg)

	events := dga.agent.Run(ctx, msgs)

	// 添加AI回复

	ch := make(chan string, 30)

	go func(ch chan string, is_stream bool) {
		run_result := ""
		for {
			event, ok := events.Next()
			if !ok {
				mem.Append(msg)
				mem.Append(&schema.Message{Role: "assistant", Content: run_result})
				if !is_stream {
					ch <- run_result
				}
				break
			}
			if event.Err != nil {
				slog.Error("执行错误", "err", event.Err)
				break
			}
			// 打印智能体输出（计划、执行结果、最终响应等）
			if msg, err := event.Output.MessageOutput.GetMessage(); err == nil && msg.Content != "" {
				if is_stream {
					ch <- msg.Content
				}
				run_result += msg.Content
			}
		}
	}(ch, is_stream)

	return ch, nil
}

// 发起疑问
func (dga *DigitalAgent) Response(ctx context.Context, question string, is_stream bool) (chan string, error) {
	sessionId := ctx.Value("session_id").(string)
	if sessionId == "" {
		return nil, errors.New("invalid session_id")
	}

	// 获取对话记录
	mem, err := dga.sessions.GetOrCreate(sessionId)
	if err != nil {
		return nil, err
	}

	// 添加用户消息
	msgs := mem.GetMessages()
	msgs = append(msgs, &schema.Message{
		Role:    "user",
		Content: question,
	})

	events := dga.agent.Run(ctx, msgs)

	ch := make(chan string, 30)

	go func(ch chan string, is_stream bool) {
		run_result := ""
		for {
			event, ok := events.Next()
			if !ok {
				break
			}
			if event.Err != nil {
				slog.Error("执行错误", "err", event.Err)
				break
			}
			// 打印智能体输出（计划、执行结果、最终响应等）
			if msg, err := event.Output.MessageOutput.GetMessage(); err == nil && msg.Content != "" {
				if is_stream {
					ch <- msg.Content
				} else {
					run_result += msg.Content
				}
			}
		}
	}(ch, is_stream)

	return ch, nil
}
