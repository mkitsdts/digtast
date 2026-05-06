package qq

import (
	"bytes"
	"context"
	"digital-labor/internal/center"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

type C2CMessage struct {
	ID     string `json:"id"`
	Author struct {
		UserOpenID string `json:"user_openid"`
	} `json:"author"`
	Content     string `json:"content"`
	TimeStamp   string `json:"timestamp"`
	Attachments []any  `json:"attachments"` // 富消息文本
}

// C2CMessageEventHandler 实现处理 at 消息的回调
func (c *QQChannel) C2CMessageEventHandler() func(msg json.RawMessage) error {
	return func(msg json.RawMessage) error {
		data := C2CMessage{}
		if err := json.Unmarshal(msg, &data); err != nil {
			return err
		}
		slog.Debug("message received", "content", data.Content)
		ag, err := center.AgentManager.GetAgent(c.AgentID)
		if err != nil {
			ag, err = center.AgentManager.GetDefaultAgent()
			if err != nil {
				slog.Error("failed to get agent", "error", err)
				return err
			}
		}
		content := data.Content
		sm, err := ag.Run(context.Background(), content, false)
		if err != nil {
			return err
		}
		result := <-sm
		return c.SendMessage(result, data.Author.UserOpenID, data.ID)
	}
}

func (c *QQChannel) SendMessage(content, user_id, msg_id string) error {
	// /v2/users/{openid}/messages 需要调用 HTTP POST 接口发送消息，珠宝要回家了，由于需要git远程同步，先暂时提交一下
	url := fmt.Sprintf("https://api.sgroup.qq.com/v2/users/%s/messages", user_id)
	payload := map[string]any{
		"content":  content,
		"msg_type": 0,
		"event_id": "C2C_MSG_RECEIVE",
		"msg_id":   msg_id,
	}

	body := bytes.NewBuffer(marshalJSON(payload))
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.token())
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		slog.Error("send message fatal", "status", resp.StatusCode, "user_id", user_id, "msg_id", msg_id, "token", c.token())
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}
