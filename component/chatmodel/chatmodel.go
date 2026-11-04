package chatmodel

import (
	"context"
	"digital-labor/schema"
)

type ChatModel interface {
	Chat(ctx context.Context, req *schema.ChatRequest) (*schema.ChatResponse, error)
	ChatStream(ctx context.Context, req *schema.ChatRequest) (schema.Stream, error)
	GetUsage() schema.Usage
	CheckStatus(ctx context.Context) (*Status, error)
}
