package qq

import (
	"context"
	"digital-labor/internal/center"
	"encoding/json"
	"log"
	"time"

	"github.com/tencent-connect/botgo/dto"
)

type C2CMessage struct {
	ID          string `json:"id"`
	Author      any    `json:"author"`
	Content     string `json:"content"`
	TimeStamp   int64  `json:"timestamp"`
	Attachments []any  `json:"attachments"` // 富消息文本
}

// C2CMessageEventHandler 实现处理 at 消息的回调
func (c *QQChannel) C2CMessageEventHandler() func(msg json.RawMessage) error {
	return func(msg json.RawMessage) error {
		data := C2CMessage{}
		if err := json.Unmarshal(msg, &data); err != nil {
			return err
		}
		ag, err := center.AgentManager.GetAgent(c.AgentID)
		if err != nil {
			return err
		}
		content := data.Content
		sm, err := ag.Run(context.Background(), content, false)
		if err != nil {
			return err
		}
		result := <-sm
		c.SendMessage(content, data.ID, generateMessage(result, data.ID))
		return nil
	}
}

func (c *QQChannel) SendMessage(content, user_id string, toCreate dto.APIMessage) error {
	// /v2/users/{openid}/messages 需要调用 HTTP POST 接口发送消息，珠宝要回家了，由于需要git远程同步，先暂时提交一下
	if _, err := c.api.PostC2CMessage(context.Background(), user_id, toCreate); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func generateMessage(content string, id string) *dto.MessageToCreate {
	return &dto.MessageToCreate{
		Timestamp: time.Now().UnixMilli(),
		Content:   content,
		MessageReference: &dto.MessageReference{
			MessageID:             id,
			IgnoreGetMessageError: true,
		},
		MsgID: id,
	}
}
