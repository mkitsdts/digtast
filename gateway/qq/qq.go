package qq

import (
	"context"
	"digital-labor/internal/gateway"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/tencent-connect/botgo/openapi"
	"github.com/tencent-connect/botgo/websocket"
)

type QQChannel struct {
	AppID     string
	AppSecret string
	Port      int
	Bots      map[string]string // 存储机器人与AgentID的映射关系
	api       openapi.OpenAPI
	conn      *websocket.WebSocket

	mu        sync.Mutex
	sessionID string
	lastSeq   int
	handlers  map[string]func(json.RawMessage) error
}

func init() {
	gateway.RegisterChannel("qq", NewQQChannel)
}

var qq *QQChannel = &QQChannel{}

func NewQQChannel(kind string) gateway.MessageChannel {
	return qq
}

func (c *QQChannel) Init(params map[string]any) error {
	if appID, ok := params["appid"]; !ok {
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
		"bots":      c.Bots,
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
	url, err := c.getWebSocketUrl()
	if err != nil {
		slog.Error("get web socket url fatal", "error", err)
		return err
	}
	c.Handler(ctx, url)
	return nil
}

func (c *QQChannel) IsActive() bool {
	return c.AppID != "" && c.AppSecret != ""
}
