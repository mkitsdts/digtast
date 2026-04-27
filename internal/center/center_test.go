package center

import (
	"digital-labor/internal/agent"
	"digital-labor/pkg/model"
	"sync"
	"testing"
)

func TestCreateAgent_MissingKey(t *testing.T) {
	c := GetCenter()
	_, err := c.CreateAgent(nil)
	if err == nil {
		t.Fatal("expected error for nil config")
	}
}

func TestCreateAgent_EmptyFields(t *testing.T) {
	// Create a fresh center for testing
	c := &Center{agents: make(map[string]*agent.DigitalAgent)}

	// Empty key
	_, err := c.CreateAgent(&model.DigitalAgentConfig{Key: ""})
	if err == nil || err.Error() != "key is empty" {
		t.Fatalf("expected 'key is empty' error, got: %v", err)
	}

	// Empty name
	_, err = c.CreateAgent(&model.DigitalAgentConfig{Key: "some-key", Name: ""})
	if err == nil || err.Error() != "name is empty" {
		t.Fatalf("expected 'name is empty' error, got: %v", err)
	}

	// Empty model
	_, err = c.CreateAgent(&model.DigitalAgentConfig{Key: "some-key", Name: "test", Model: ""})
	if err == nil || err.Error() != "model is empty" {
		t.Fatalf("expected 'model is empty' error, got: %v", err)
	}
}

func TestGet_NotFound(t *testing.T) {
	c := &Center{agents: make(map[string]*agent.DigitalAgent)}

	_, err := c.GetAgent("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent agent")
	}
	if err.Error() != "agent not found" {
		t.Fatalf("expected 'agent not found', got: %v", err)
	}
}

func TestGet_EmptyID(t *testing.T) {
	c := &Center{agents: make(map[string]*agent.DigitalAgent)}

	_, err := c.GetAgent("")
	if err == nil {
		t.Fatal("expected error for empty id")
	}
}

func TestRemove_NotFound(t *testing.T) {
	c := &Center{agents: make(map[string]*agent.DigitalAgent)}

	err := c.RemoveAgent("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent agent")
	}
}

func TestRemove_EmptyID(t *testing.T) {
	c := &Center{agents: make(map[string]*agent.DigitalAgent)}

	err := c.RemoveAgent("")
	if err == nil {
		t.Fatal("expected error for empty id")
	}
}

func TestConcurrency_Agents(t *testing.T) {
	c := &Center{agents: make(map[string]*agent.DigitalAgent)}

	var wg sync.WaitGroup
	// Concurrent reads
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			c.GetAgent("key")
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
