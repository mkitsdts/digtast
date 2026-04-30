package qq

import (
	"context"
	"digital-labor/internal/center"
	"log"
	"time"

	"github.com/tencent-connect/botgo/dto"
	"github.com/tencent-connect/botgo/event"
)

// C2CMessageEventHandler 实现处理 at 消息的回调
func (c *QQChannel) C2CMessageEventHandler() event.C2CMessageEventHandler {
	return func(event *dto.WSPayload, data *dto.WSC2CMessageData) error {
		ag, err := center.AgentManager.GetAgent(c.Bots[data.Author.ID])
		if err != nil {
			return err
		}
		content := data.Content
		sm, err := ag.Run(context.Background(), content, false)
		if err != nil {
			return err
		}
		result := <-sm
		msg := dto.Message(*data)
		c.SendMessage(content, data.ID, generateMessage(result, msg))
		return nil
	}
}

func (c *QQChannel) SendMessage(content, user_id string, toCreate dto.APIMessage) error {
	if _, err := c.api.PostC2CMessage(context.Background(), user_id, toCreate); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func generateMessage(content string, data dto.Message) *dto.MessageToCreate {
	return &dto.MessageToCreate{
		Timestamp: time.Now().UnixMilli(),
		Content:   content,
		MessageReference: &dto.MessageReference{
			MessageID:             data.ID,
			IgnoreGetMessageError: true,
		},
		MsgID: data.ID,
	}
}
