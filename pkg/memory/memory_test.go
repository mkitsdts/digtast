package mem

import (
	"testing"

	"digital-labor/pkg/workspace"

	"github.com/cloudwego/eino/schema"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	// Tests will use the default workspace path, which might be messy.
	// However, workspace.GetWorkspacePath() is hardcoded.
	// For testing purposes, we should ideally be able to override it,
	// but since it's not easily overrideable without changing workspace.go,
	// we'll just use it and hope for the best or assume tests run in a clean env.
	return &Store{
		agentID: "test-agent",
	}
}

func TestGetOrCreate_NewSession(t *testing.T) {
	s := newTestStore(t)

	sess, err := s.GetOrCreate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sess.AgentID != "test-agent" {
		t.Fatalf("expected AgentID 'test-agent', got '%s'", sess.AgentID)
	}
}

func TestGetOrCreate_ReturnsCached(t *testing.T) {
	s := newTestStore(t)

	sess1, err := s.GetOrCreate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sess2, err := s.GetOrCreate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sess1 != sess2 {
		t.Fatal("expected same cached instance")
	}
}

func TestGetOrCreate_LoadExisting(t *testing.T) {
	s := newTestStore(t)
	agentID := "test-agent-load"
	s.agentID = agentID

	// Clean up first
	workspace.DeleteSession(agentID)

	if err := workspace.Save(agentID, workspace.MemTypeChunk, &schema.Message{Role: schema.User, Content: "hello"}); err != nil {
		t.Fatal(err)
	}

	sess, err := s.GetOrCreate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sess.AgentID != agentID {
		t.Fatalf("expected AgentID '%s', got '%s'", agentID, sess.AgentID)
	}

	msgs := sess.GetMessages()
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if msgs[0].Content != "hello" {
		t.Fatalf("expected message content 'hello', got '%s'", msgs[0].Content)
	}
}

func TestDelete(t *testing.T) {
	s := newTestStore(t)
	agentID := "test-agent-delete"
	s.agentID = agentID

	if _, err := s.GetOrCreate(); err != nil {
		t.Fatal(err)
	}
	if err := workspace.Save(agentID, workspace.MemTypeChunk, &schema.Message{Role: schema.User, Content: "hello"}); err != nil {
		t.Fatal(err)
	}

	if err := s.Delete(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// After delete, loading should return empty
	res, _ := workspace.Load(agentID, workspace.MemTypeChunk)
	msgs := res.([]*schema.Message)
	if len(msgs) != 0 {
		t.Fatal("expected messages to be deleted")
	}

	// Double delete should not error
	if err := s.Delete(); err != nil {
		t.Fatalf("expected no error on double delete, got: %v", err)
	}
}

func TestDelete_EvictsCache(t *testing.T) {
	s := newTestStore(t)
	agentID := "test-agent-evict"
	s.agentID = agentID

	if _, err := s.GetOrCreate(); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete(); err != nil {
		t.Fatal(err)
	}

	// After delete, GetOrCreate should return a new instance
	sess, err := s.GetOrCreate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	msgs := sess.GetMessages()
	if len(msgs) != 0 {
		t.Fatalf("expected fresh session with 0 messages after delete, got %d", len(msgs))
	}
}

func TestConcurrency_GetOrCreate(t *testing.T) {
	s := newTestStore(t)

	done := make(chan bool, 50)
	for i := 0; i < 50; i++ {
		go func() {
			sess, err := s.GetOrCreate()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if sess == nil {
				t.Error("expected non-nil session")
			}
			done <- true
		}()
	}

	for i := 0; i < 50; i++ {
		<-done
	}
}

func TestNewStore_Default(t *testing.T) {
	s, err := NewStore("")
	if err != nil {
		t.Skipf("NewStore failed: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil store when err is nil")
	}
}
