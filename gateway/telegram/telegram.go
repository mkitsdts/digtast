package telegram

import (
	"context"
	gw "digital-labor/internal/gateway"
	"digital-labor/pkg/gateway"
	"errors"
	"fmt"
	"log/slog"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type TelegramChannel struct {
	Token string `json:"token"`
	Bot   *bot.Bot
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
	return nil
}

func (tg *TelegramChannel) Send(result gateway.Result, params map[string]any) error {
	chatid, ok := params["chat_id"]
	if !ok {
		return errors.New("chat_id not found")
	}
	if tg.Bot == nil {
		return errors.New("telegram bot is not initialized")
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
		"token": tg.Token,
	}
}

func (c *TelegramChannel) Register() error {
	// 引导用户输入
	fmt.Println("请输入Telegram机器人的Token:")
	fmt.Print("Token: ")
	_, err := fmt.Scan(&c.Token)
	if err != nil {
		return err
	}
	return nil
}

func (tg *TelegramChannel) Serve(ctx context.Context) error {
	opts := []bot.Option{
		// 这个默认处理器会接收到所有消息
		bot.WithDefaultHandler(func(ctx context.Context, b *bot.Bot, update *models.Update) {
			if update.Message != nil && update.Message.Text != "" {
				slog.Info("received message", "text", update.Message.Text)
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
