package mem

import (
	"path/filepath"
	"testing"

	"digital-labor/pkg/workspace"

	"github.com/cloudwego/eino/schema"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	persist, err := workspace.NewMemoryStore(filepath.Join(dir, "memory"))
	if err != nil {
		t.Fatal(err)
	}
	return &Store{
		agentID: "test-agent",
		cache:   make(map[string]*Session),
		persist: persist,
	}
}

func TestGetOrCreate_NewSession(t *testing.T) {
	s := newTestStore(t)

	sess, err := s.GetOrCreate("new-session-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sess.ID != "new-session-1" {
		t.Fatalf("expected ID 'new-session-1', got '%s'", sess.ID)
	}

	if len(s.persist.ListSessions("test-agent")) != 1 {
		t.Fatal("expected session index entry to exist")
	}
}

func TestGetOrCreate_ReturnsCached(t *testing.T) {
	s := newTestStore(t)

	sess1, err := s.GetOrCreate("cached-session")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sess2, err := s.GetOrCreate("cached-session")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sess1 != sess2 {
		t.Fatal("expected same cached instance")
	}
}

func TestGetOrCreate_LoadExisting(t *testing.T) {
	s := newTestStore(t)

	if _, err := s.persist.NewSessionFile("test-agent", "existing"); err != nil {
		t.Fatal(err)
	}
	if err := s.persist.PersistMessage("test-agent", "existing", &schema.Message{Role: schema.User, Content: "hello"}); err != nil {
		t.Fatal(err)
	}

	sess, err := s.GetOrCreate("existing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sess.ID != "existing" {
		t.Fatalf("expected ID 'existing', got '%s'", sess.ID)
	}

	msgs := sess.GetMessages()
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if msgs[0].Content != "hello" {
		t.Fatalf("expected message content 'hello', got '%s'", msgs[0].Content)
	}
}

func TestList(t *testing.T) {
	s := newTestStore(t)

	s.GetOrCreate("sess-a")
	s.GetOrCreate("sess-b")

	metas, err := s.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(metas) != 2 {
		t.Fatalf("expected 2 sessions, got %d", len(metas))
	}
}

func TestDelete(t *testing.T) {
	s := newTestStore(t)

	s.GetOrCreate("to-delete")
	if err := s.persist.PersistMessage("test-agent", "to-delete", &schema.Message{Role: schema.User, Content: "hello"}); err != nil {
		t.Fatal(err)
	}

	if err := s.Delete("to-delete"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(s.persist.ListSessions("test-agent")) != 0 {
		t.Fatal("expected session index entry to be deleted")
	}

	// Double delete should not error
	if err := s.Delete("to-delete"); err != nil {
		t.Fatalf("expected no error on double delete, got: %v", err)
	}
}

func TestDelete_EvictsCache(t *testing.T) {
	s := newTestStore(t)

	s.GetOrCreate("cache-del")
	s.Delete("cache-del")

	// After delete, GetOrCreate should return a new instance
	sess, err := s.GetOrCreate("cache-del")
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
			sess, err := s.GetOrCreate("concurrent-session")
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
	// Documents that NewStore() uses global workspace path and returns nil on failure.
	// This test cannot run in isolation because dir is unexported in Store.
	// The returned Store may use the user's home directory — a design issue to fix later.
	s := NewStore()
	if s == nil {
		t.Skip("NewStore returned nil, workspace path may be unavailable")
	}
}
