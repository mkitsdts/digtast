package gateway

import "context"

// 定义消息通知接口
type MessageChannel interface {
	Send(content string) error       // 发送消息给用户
	IsActive() bool                  // 判断通道是否活跃
	Register() error                 // 注册通道
	Serve(ctx context.Context) error // 监听获取用户发送的消息
	GetConfig() map[string]any       // 获取通道配置
}
