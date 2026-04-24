package agent

import (
	"context"
	"testing"
)

func TestChat(t *testing.T) {
	agent, err := NewDigitalAgent("doubao", "***REMOVED***", "https://ark.cn-beijing.volces.com/api/v3", "doubao-seed-2-0-lite-260215", "114514")
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, "session_id", "114514")

	ch, err := agent.Run(ctx, "帮我查一下广州的天气", true)
	if err != nil {
		t.Fatal(err)
	}
	for result := range ch {
		t.Log(result)
	}
}
