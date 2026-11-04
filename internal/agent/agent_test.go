package agent

import (
	"context"
	"testing"

	"digital-labor/component/chatmodel"
	"digital-labor/config"
	"digital-labor/internal/agent/state"
	"digital-labor/schema"
)

// stubProvider 注册一个仅供测试使用的厂商，避免单测依赖真实 Provider。
type stubModel struct{}

func (m stubModel) Chat(context.Context, *schema.ChatRequest) (*schema.ChatResponse, error) {
	return &schema.ChatResponse{Message: schema.Message{Role: schema.RoleAssistant, Content: "stub"}}, nil
}
func (m stubModel) GetUsage() schema.Usage { return schema.Usage{} }
func (m stubModel) CheckStatus(context.Context) (*chatmodel.Status, error) {
	return &chatmodel.Status{Available: true}, nil
}

func init() {
	chatmodel.RegisterProvider(chatmodel.Provider{
		ID: "stub", Label: "Stub", BaseURL: "https://stub.test",
		NewClient: func(chatmodel.ClientOptions, string) (chatmodel.ChatModel, error) { return stubModel{}, nil },
		ListModels: func(context.Context, chatmodel.ClientOptions) ([]string, error) {
			return []string{"stub-model"}, nil
		},
	})
}

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
		"stub": {Key: "test-key", Models: []string{"stub-model"}},
	}
	first, err := Initialize(&config.AgentConfig{Name: "first", Provider: "stub", Model: "stub-model"})
	if err != nil {
		t.Fatal(err)
	}
	second, err := Initialize(&config.AgentConfig{Name: "second", Provider: "stub", Model: "stub-model"})
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

func TestNewAgentRejectsUnknownProvider(t *testing.T) {
	config.Conf.WorkSpaceDir = t.TempDir()
	config.Conf.Models = map[string]config.ModelConfig{"stub": {Key: "test-key"}}
	if _, err := newAgent(&config.AgentConfig{Name: "x", Provider: "missing", Model: "m"}); err == nil {
		t.Fatal("expected an error for an unregistered provider")
	}
}

func TestNewAgentRejectsMissingCredentials(t *testing.T) {
	config.Conf.WorkSpaceDir = t.TempDir()
	config.Conf.Models = map[string]config.ModelConfig{}
	if _, err := newAgent(&config.AgentConfig{Name: "x", Provider: "stub", Model: "stub-model"}); err == nil {
		t.Fatal("expected an error when the provider has no api key")
	}
}

func TestNewAgentFallsBackToFirstEnabledModel(t *testing.T) {
	config.Conf.WorkSpaceDir = t.TempDir()
	config.Conf.Models = map[string]config.ModelConfig{
		"stub": {Key: "test-key", Models: []string{"stub-model"}},
	}
	ag, err := newAgent(&config.AgentConfig{Name: "x", Provider: "stub"})
	if err != nil {
		t.Fatal(err)
	}
	if ag.GetModel() == nil {
		t.Fatal("expected a model client")
	}
	if _, err := newAgent(&config.AgentConfig{Name: "x", Provider: "stub", Model: "nope"}); err != nil {
		t.Fatalf("explicit model should pass validation: %v", err)
	}
}

func TestFilterUserAssistant(t *testing.T) {
	msgs := []*schema.Message{
		{Role: schema.RoleUser, Content: "hello"},
		{Role: schema.RoleAssistant, Content: "hi"},
		{Role: schema.RoleTool, Content: "tool result", Name: "search"},
		{Role: schema.RoleUser, Content: "thanks"},
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
		{Role: schema.RoleUser, Content: "hello world"},       // ~3 tokens
		{Role: schema.RoleAssistant, Content: "hi there"},     // ~2 tokens
		{Role: schema.RoleUser, Content: "how are you doing"}, // ~4 tokens
	}

	shards := state.ShardMessages(msgs, 5)
	if len(shards) < 2 {
		t.Fatalf("expected at least 2 shards, got %d", len(shards))
	}
}
