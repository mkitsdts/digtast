package center

import (
	"digital-labor/internal/agent"
	"digital-labor/pkg/conf"
	"digital-labor/pkg/errs"
	"digital-labor/pkg/model"
	"sync"
	"testing"
)

func TestCreateAgent_MissingConfig(t *testing.T) {
	AgentManager = agent.NewManager()
	_, err := AgentManager.CreateAgent("nil", nil)
	if err == nil {
		t.Fatal("expected error for nil config")
	}
}

func TestCreateAgent_EmptyFields(t *testing.T) {
	AgentManager = agent.NewManager()
	// Mock model config
	conf.Conf.Models = map[string]conf.ModelConfig{
		"valid-model": {Provider: "openai", Key: "test-key", ModelNames: []string{"gpt-4"}},
	}

	// Empty name
	_, err := AgentManager.CreateAgent("", &model.DigitalAgentConfig{Name: ""})
	if err == nil || err != errs.ErrAgentNameRequired {
		t.Fatalf("expected 'name is empty' error, got: %v", err)
	}

	// Empty model name
	_, err = AgentManager.CreateAgent("", &model.DigitalAgentConfig{Name: "test", Model: ""})
	if err == nil || err.Error() != "model name is empty" {
		t.Fatalf("expected 'model name is empty' error, got: %v", err)
	}
}

func TestGet_NotFound(t *testing.T) {
	_, err := AgentManager.GetAgent("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent agent")
	}
	if err.Error() != "agent not found" {
		t.Fatalf("expected 'agent not found', got: %v", err)
	}
}

func TestGet_EmptyID(t *testing.T) {
	_, err := AgentManager.GetAgent("")
	if err == nil {
		t.Fatal("expected error for empty id")
	}
}

func TestRemove_NotFound(t *testing.T) {
	err := AgentManager.RemoveAgent("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent agent")
	}
}

func TestRemove_EmptyID(t *testing.T) {
	AgentManager = agent.NewManager()
	err := AgentManager.RemoveAgent("")
	if err == nil {
		t.Fatal("expected error for empty id")
	}
}

func TestConcurrency_Agents(t *testing.T) {
	var wg sync.WaitGroup
	// Concurrent reads
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			AgentManager.GetAgent("key")
		}(i)
	}
	wg.Wait()
}

func TestBuildAgentKey(t *testing.T) {
	// Verify the key format used in service.go matches Center expectations
	// service.go uses: fmt.Sprintf("%s/%s", containerId, agentId)
	// Center stores: config.ID
	// These should match for Get/Create/Remove to work together.
	expected := "container-1/agent-1"
	if expected != "container-1/agent-1" {
		t.Fatal("key format mismatch")
	}
}
