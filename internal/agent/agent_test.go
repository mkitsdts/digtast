package agent

import (
	"digital-labor/internal/state"
	"testing"

	"github.com/cloudwego/eino/schema"
)

func TestFilterUserAssistant(t *testing.T) {
	msgs := []*schema.Message{
		{Role: schema.User, Content: "hello"},
		{Role: schema.Assistant, Content: "hi"},
		{Role: schema.Tool, Content: "tool result", ToolName: "search"},
		{Role: schema.User, Content: "thanks"},
	}

	filtered := state.FilterUserAssistant(msgs)
	if len(filtered) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(filtered))
	}
	if filtered[0].Content != "hello" || filtered[1].Content != "hi" || filtered[2].Content != "thanks" {
		t.Fatalf("unexpected filtered messages: %v", filtered)
	}
}

func TestShardMessages(t *testing.T) {
	msgs := []*schema.Message{
		{Role: schema.User, Content: "hello world"},       // ~3 tokens
		{Role: schema.Assistant, Content: "hi there"},     // ~2 tokens
		{Role: schema.User, Content: "how are you doing"}, // ~4 tokens
	}

	shards := state.ShardMessages(msgs, 5)
	if len(shards) < 2 {
		t.Fatalf("expected at least 2 shards, got %d", len(shards))
	}
}
