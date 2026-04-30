package qq

import (
	"context"
	"digital-labor/internal/gateway"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/tencent-connect/botgo"
	"github.com/tencent-connect/botgo/openapi"
	"github.com/tencent-connect/botgo/token"
	"github.com/tencent-connect/botgo/websocket"
)

type QQChannel struct {
	AppID     string
	AppSecret string
	Port      int
	Bots      map[string]string // 存储机器人与AgentID的映射关系
	api       openapi.OpenAPI
	conn      *websocket.WebSocket
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
	fmt.Println("请确保9000端口已开放")
	c.Port = 9000
	return nil
}

func (c *QQChannel) Serve(ctx context.Context) error {
	//创建oauth2标准token source
	credentials := &token.QQBotCredentials{
		AppID:     c.AppID,
		AppSecret: c.AppSecret,
	}
	tokenSource := token.NewQQBotTokenSource(credentials)
	//启动自动刷新access token协程
	if err := token.StartRefreshAccessToken(ctx, tokenSource); err != nil {
		slog.Error("start refresh access token fatal", "error", err)
		return err
	}
	// 初始化 openapi，正式环境
	c.api = botgo.NewOpenAPI(c.AppID, tokenSource).WithTimeout(5 * time.Second)
	return nil
}

func (c *QQChannel) IsActive() bool {
	return c.AppID != "" && c.AppSecret != ""
}
