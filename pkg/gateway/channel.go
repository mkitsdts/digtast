package gateway

import (
	"context"
	"digital-labor/pkg/model"
)

// 定义消息通知接口
type MessageChannel interface {
	Send(result model.Result, params map[string]any) error // 发送消息给用户
	IsActive() bool                                        // 判断通道是否活跃
	Register() error                                       // 注册通道
	Serve(ctx context.Context) error                       // 监听获取用户发送的消息
	GetConfig() map[string]any                             // 获取通道配置
	Init(params map[string]any) error
}
