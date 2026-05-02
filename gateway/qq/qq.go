package qq

import (
	"context"
	"digital-labor/internal/gateway"
	"digital-labor/pkg/conf"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"
)

type QQChannel struct {
	AppID     string
	AppSecret string
	Port      int
	AgentID   string // 存储机器人与AgentID的映射关系

	mu           sync.Mutex
	sessionID    string
	lastSeq      int
	refreshToken string
	refreshMux   sync.Mutex
	handlers     map[string]func(json.RawMessage) error
}

func init() {
	gateway.RegisterChannel("qq", NewQQChannel)
}

var qq *QQChannel = &QQChannel{}

func NewQQChannel(kind string) gateway.MessageChannel {
	return qq
}

func (c *QQChannel) Init(params map[string]any) error {
	if appID, ok := params["appid"]; ok {
		c.AppID = appID.(string)
	} else {
		return errors.New("invaild app id which register channel qq")
	}
	if secret, ok := params["appsecret"]; ok {
		c.AppSecret = secret.(string)
	} else {
		return errors.New("invaild app secret which register channel qq")
	}
	if port, ok := params["port"]; ok {
		c.Port = int(port.(float64))
	} else {
		return errors.New("invaild port which register channel qq")
	}
	if agentID, ok := params["agentid"]; ok {
		c.AgentID = agentID.(string)
	} else {
		c.AgentID = conf.Conf.State.LastUsedAgent
	}
	return nil
}

func (c *QQChannel) Send(content string) error {
	return nil
}

func (c *QQChannel) GetConfig() map[string]any {
	return map[string]any{
		"appid":     c.AppID,
		"appsecret": c.AppSecret,
		"port":      c.Port,
		"agentid":   c.AgentID,
	}
}

func (c *QQChannel) Register() error {
	// 引导用户输入
	fmt.Println("请输入QQ机器人的AppID和AppSecret:")
	fmt.Print("AppID: ")
	_, err := fmt.Scan(&c.AppID)
	if err != nil {
		return err
	}
	fmt.Print("AppSecret: ")
	_, err = fmt.Scan(&c.AppSecret)
	if err != nil {
		return err
	}
	return nil
}

func (c *QQChannel) Serve(ctx context.Context) error {
	fmt.Println("serve qq channel")
	go c.refreshTokenLoop(ctx)
	fmt.Println("before getWebSocketUrl")
	url, err := c.getWebSocketUrl()
	if err != nil {
		slog.Error("get web socket url fatal", "error", err)
		return err
	}

	c.RegisterHandler("C2C_MESSAGE_CREATE", c.C2CMessageEventHandler())

	slog.Debug("get web socket url success", "url", url)
	c.Handler(ctx, url)
	return nil
}

func (c *QQChannel) IsActive() bool {
	return c.AppID != "" && c.AppSecret != ""
}
