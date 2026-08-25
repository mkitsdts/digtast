package mem

import (
	"testing"
	"time"

	"digital-labor/pkg/conf"
	"digital-labor/pkg/llm"
	"digital-labor/pkg/workspace"
)

func newTestSessionWithFile(t *testing.T) *Session {
	t.Helper()
	conf.Conf.WorkSpaceDir = t.TempDir()
	store := &Store{
		agentID: "test-agent",
	}

	sess := &Session{
		AgentID:   "test-agent",
		CreatedAt: time.Now(),
		messages:  make([]*llm.Message, 0),
	}
	store.session = sess
	return sess
}

func TestAppend(t *testing.T) {
	sess := newTestSessionWithFile(t)
	workspace.DeleteSession(sess.AgentID)

	msg := &llm.Message{
		Role:    llm.RoleUser,
		Content: "hello",
	}
	if err := sess.Append(msg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msgs := sess.GetMessages()
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if msgs[0].Content != "hello" {
		t.Fatalf("expected 'hello', got '%s'", msgs[0].Content)
	}
}

func TestAppend_Multiple(t *testing.T) {
	sess := newTestSessionWithFile(t)
	workspace.DeleteSession(sess.AgentID)

	sess.Append(&llm.Message{Role: llm.RoleUser, Content: "msg1"})
	sess.Append(&llm.Message{Role: llm.RoleAssistant, Content: "reply1"})
	sess.Append(&llm.Message{Role: llm.RoleUser, Content: "msg2"})

	msgs := sess.GetMessages()
	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(msgs))
	}
}

func TestGetMessages_ReturnsSnapshot(t *testing.T) {
	sess := newTestSessionWithFile(t)
	sess.Append(&llm.Message{Role: llm.RoleUser, Content: "original"})

	msgs1 := sess.GetMessages()
	msgs1[0].Content = "modified"

	msgs2 := sess.GetMessages()
	if msgs2[0].Content != "original" {
		t.Fatal("GetMessages should return a copy, not the internal slice")
	}
}

func TestTitle(t *testing.T) {
	sess := newTestSessionWithFile(t)
	if title := sess.Title(); title != "New Session" {
		t.Fatalf("expected 'New Session', got '%s'", title)
	}

	sess.Append(&llm.Message{Role: llm.RoleAssistant, Content: "ignored"})
	sess.Append(&llm.Message{Role: llm.RoleUser, Content: "first user msg"})

	if title := sess.Title(); title != "first user msg" {
		t.Fatalf("expected 'first user msg', got '%s'", title)
	}
}

func TestTitle_Truncation(t *testing.T) {
	sess := newTestSessionWithFile(t)
	longContent := ""
	for i := 0; i < 100; i++ {
		longContent += "中"
	}
	sess.Append(&llm.Message{Role: llm.RoleUser, Content: longContent})

	title := sess.Title()
	if len([]rune(title)) > 63 {
		t.Fatalf("expected truncated title, got length %d", len([]rune(title)))
	}
}

func TestSetAndGetPendingInterruptID(t *testing.T) {
	sess := newTestSessionWithFile(t)

	if id := sess.GetPendingInterruptID(); id != "" {
		t.Fatalf("expected empty, got '%s'", id)
	}

	sess.SetPendingInterruptID("interrupt-123")
	if id := sess.GetPendingInterruptID(); id != "interrupt-123" {
		t.Fatalf("expected 'interrupt-123', got '%s'", id)
	}
}

func TestSetAndGetMsgIdx(t *testing.T) {
	sess := newTestSessionWithFile(t)

	if idx := sess.GetMsgIdx(); idx != 0 {
		t.Fatalf("expected 0, got %d", idx)
	}

	sess.SetMsgIdx(42)
	if idx := sess.GetMsgIdx(); idx != 42 {
		t.Fatalf("expected 42, got %d", idx)
	}
}

func TestAppend_PersistsToDisk(t *testing.T) {
	sess := newTestSessionWithFile(t)
	workspace.DeleteSession(sess.AgentID)

	sess.Append(&llm.Message{Role: llm.RoleUser, Content: "persisted"})

	res, err := workspace.Load(sess.AgentID, workspace.MemTypeChunk)
	if err != nil {
		t.Fatal(err)
	}
	msgs := res.([]*llm.Message)
	if len(msgs) != 1 {
		t.Fatalf("expected 1 persisted message, got %d", len(msgs))
	}
	if msgs[0].Content != "persisted" {
		t.Fatalf("expected persisted message, got %q", msgs[0].Content)
	}
}
