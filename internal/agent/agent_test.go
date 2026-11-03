package agent

import (
	"testing"

	"digital-labor/config"
	"digital-labor/internal/agent/state"
	"digital-labor/pkg/chatmodel"
)

func TestInitializeReturnsSingleton(t *testing.T) {
	instanceMu.Lock()
	previous := instance
	instance = nil
	instanceMu.Unlock()
	t.Cleanup(func() {
		instanceMu.Lock()
		instance = previous
		instanceMu.Unlock()
	})

	config.Conf.WorkSpaceDir = t.TempDir()
	config.Conf.Models = map[string]config.ModelConfig{
		"test": {Provider: "openai", Key: "test-key", ModelNames: []string{"test-model"}},
	}
	first, err := Initialize(&config.AgentConfig{Name: "first", Model: "test"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := Initialize(&config.AgentConfig{Name: "second", Model: "test"})
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatal("expected Initialize to return the same Agent instance")
	}
	got, err := Instance()
	if err != nil {
		t.Fatal(err)
	}
	if got != first || got.Name != "first" {
		t.Fatalf("unexpected singleton: %#v", got)
	}
}

func TestFilterUserAssistant(t *testing.T) {
	msgs := []*chatmodel.Message{
		{Role: chatmodel.RoleUser, Content: "hello"},
		{Role: chatmodel.RoleAssistant, Content: "hi"},
		{Role: chatmodel.RoleTool, Content: "tool result", Name: "search"},
		{Role: chatmodel.RoleUser, Content: "thanks"},
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
	msgs := []*chatmodel.Message{
		{Role: chatmodel.RoleUser, Content: "hello world"},       // ~3 tokens
		{Role: chatmodel.RoleAssistant, Content: "hi there"},     // ~2 tokens
		{Role: chatmodel.RoleUser, Content: "how are you doing"}, // ~4 tokens
	}

	shards := state.ShardMessages(msgs, 5)
	if len(shards) < 2 {
		t.Fatalf("expected at least 2 shards, got %d", len(shards))
	}
}
