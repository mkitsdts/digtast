package telegram

import (
	"context"
	gw "digital-labor/internal/gateway"
	"digital-labor/pkg/gateway"
	"digital-labor/pkg/model"
	"digital-labor/pkg/workspace"
	"errors"
	"fmt"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type TelegramChannel struct {
	Token   string `json:"token"`
	Bot     *bot.Bot
	AgentID string `json:"agent_id"`
}

func init() {
	gateway.RegisterChannel("telegram", NewTelegramChannel)
}

var tg *TelegramChannel = &TelegramChannel{}

func NewTelegramChannel(kind string) gateway.MessageChannel {
	return tg
}

func (tg *TelegramChannel) Init(params map[string]any) error {
	if token, ok := params["token"]; ok {
		tg.Token = token.(string)
	} else {
		return errors.New("invaild token which register channel telegram")
	}
	if agentID, ok := params["agent_id"]; ok {
		tg.AgentID = agentID.(string)
	} else {
		tg.AgentID = workspace.DefaultAgentID()
	}
	return nil
}

func (tg *TelegramChannel) Send(result model.Result, params map[string]any) error {
	chatid, ok := params["chatid"]
	if !ok {
		return errors.New("chatid not found")
	}
	ctx := context.Background()
	tg.Bot.SendMessage(ctx, &bot.SendMessageParams{
		ChatID: chatid,
		Text:   result.Content,
	})
	return nil
}

func (tg *TelegramChannel) GetConfig() map[string]any {
	return map[string]any{
		"Token":   tg.Token,
		"AgentID": tg.AgentID,
	}
}

func (c *TelegramChannel) Register() error {
	// 引导用户输入
	fmt.Println("请输入Telegram机器人的Token:")
	fmt.Print("Token: ")
	_, err := fmt.Scan(&c.Token)
	fmt.Println("请输入Telegram机器人的AgentID:")
	fmt.Print("AgentID: ")
	_, err = fmt.Scan(&c.AgentID)

	if err != nil {
		return err
	}
	return nil
}

func (tg *TelegramChannel) Serve(ctx context.Context) error {
	opts := []bot.Option{
		// 这个默认处理器会接收到所有消息
		bot.WithDefaultHandler(func(ctx context.Context, b *bot.Bot, update *models.Update) {
			slog.Info("received message", "text", update.Message.Text)
			if update.Message != nil && update.Message.Text != "" {
				userText := update.Message.Text
				// 异步处理
				gw.PushUserMessage(&gw.UserMessage{
					Content:   userText,
					NotifyWay: "telegram",
					Params:    map[string]any{"chat_id": update.Message.Chat.ID},
				})
			}
		}),
	}

	var err error
	tg.Bot, err = bot.New(tg.Token, opts...)
	if err != nil {
		return err
	}
	tg.Bot.Start(ctx)

	return nil
}

func (tg *TelegramChannel) IsActive() bool {
	return true
}
