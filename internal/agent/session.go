package agent

import (
	"context"
	"errors"
	"strings"

	"github.com/cloudwego/eino-examples/quickstart/chatwitheino/mem"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
)

// 删除会话记录
func (dga *DigitalAgent) RemoveSession(ctx context.Context) error {
	sessionId := ctx.Value("session_id").(string)

	if sessionId == "" {
		return errors.New("sessionID is empty")
	}
	return dga.sessions.Delete(sessionId)
}

// 获取对话记录
func (dga *DigitalAgent) GetOrCreateSession(sessionID string) (*mem.Session, error) {
	if sessionID == "" {
		sessionID = uuid.New().String()
	}
	session, err := dga.sessions.GetOrCreate(sessionID)
	if err != nil {
		return nil, err
	}
	return session, nil
}

// 追加对话记录
func (dga *DigitalAgent) AppendSessionMemory(ctx context.Context, role, content string, toolCalls []schema.ToolCall) error {
	sessionID := ctx.Value("session_id").(string)

	// 根据角色构造消息
	msg, err := buildMsg(role, content, toolCalls)
	if err != nil {
		return err
	}

	// 获取会话
	session, err := dga.sessions.GetOrCreate(sessionID)
	if err != nil {
		return err
	}

	// 将新消息追加到消息记录
	if err := session.Append(msg); err != nil {
		return err
	}
	return nil
}

func buildMsg(role, content string, toolCalls []schema.ToolCall) (*schema.Message, error) {
	role = strings.ToLower(role)

	switch role {
	case "user":
		return schema.UserMessage(content), nil
	case "system":
		return schema.SystemMessage(content), nil
	case "assistant":
		return schema.AssistantMessage(content, toolCalls), nil
	default:
		return nil, errors.New("Unknown Message")
	}
}
