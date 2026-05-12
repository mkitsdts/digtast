package chat

import (
	"context"

	"github.com/cloudwego/eino/adk"
)

func NewContextCompressMiddleware(maxToken int) func(ctx context.Context, sta *adk.ChatModelAgentState) error {
	return func(ctx context.Context, sta *adk.ChatModelAgentState) error {
		tokenCount := 0
		for i := range sta.Messages {
			tokenCount += len(sta.Messages[i].Content)
		}

		if tokenCount <= maxToken {
			return nil
		}

		return nil
	}
}
