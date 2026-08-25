package agent

import (
	"digital-labor/pkg/llm"
	"digital-labor/pkg/state"
	"testing"
)

func TestFilterUserAssistant(t *testing.T) {
	msgs := []*llm.Message{
		{Role: llm.RoleUser, Content: "hello"},
		{Role: llm.RoleAssistant, Content: "hi"},
		{Role: llm.RoleTool, Content: "tool result", Name: "search"},
		{Role: llm.RoleUser, Content: "thanks"},
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
	msgs := []*llm.Message{
		{Role: llm.RoleUser, Content: "hello world"},       // ~3 tokens
		{Role: llm.RoleAssistant, Content: "hi there"},     // ~2 tokens
		{Role: llm.RoleUser, Content: "how are you doing"}, // ~4 tokens
	}

	shards := state.ShardMessages(msgs, 5)
	if len(shards) < 2 {
		t.Fatalf("expected at least 2 shards, got %d", len(shards))
	}
}
