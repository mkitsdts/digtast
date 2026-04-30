package telegram

import (
	"context"
	"digital-labor/internal/center"
	"digital-labor/internal/gateway"
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

func (tg *TelegramChannel) Send(content string) error {
	return nil
}

func (tg *TelegramChannel) GetConfig() map[string]any {
	return map[string]any{
		"Token": tg.Token,
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
			slog.Info("received message", "text", update.Message.Text)
			if update.Message != nil && update.Message.Text != "" {
				userText := update.Message.Text
				// TODO: 暂时没有处理好这里，徐娅获取对应的智能体
				agent, err := center.AgentManager.GetAgent("test")
				if err != nil {
					slog.Error("get agent error", "error", err)
					return
				}
				sm, err := agent.Run(ctx, userText, false)
				if err != nil {
					return
				}

				// 在这里做你想做的任何事，比如：
				b.SendMessage(ctx, &bot.SendMessageParams{
					ChatID: update.Message.Chat.ID,
					Text:   <-sm,
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
