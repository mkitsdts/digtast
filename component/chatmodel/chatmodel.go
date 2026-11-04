package chatmodel

import (
	"context"
	"digital-labor/schema"
)

// ChatModel 是非流式模型后端。
type ChatModel interface {
	Chat(ctx context.Context, req *schema.ChatRequest) (*schema.ChatResponse, error)
	GetUsage() schema.Usage
	CheckStatus(ctx context.Context) (*Status, error)
}

// StreamChatModel 补充流式能力；流式调用返回事件迭代器。
type StreamChatModel interface {
	ChatModel
	ChatStream(ctx context.Context, req *schema.ChatRequest) schema.Stream
}
