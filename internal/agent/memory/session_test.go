package mem

import (
	"testing"
	"time"

	"digital-labor/config"
	"digital-labor/internal/agent/workspace"
	"digital-labor/schema"
)

func newTestSession(t *testing.T) *Session {
	t.Helper()
	config.Conf.WorkSpaceDir = t.TempDir()
	return &Session{CreatedAt: time.Now(), messages: make([]*schema.Message, 0)}
}

func TestSessionAppendPersistsAndSnapshots(t *testing.T) {
	session := newTestSession(t)
	if err := session.Append(&schema.Message{Role: schema.RoleUser, Content: "hello"}); err != nil {
		t.Fatal(err)
	}
	messages := session.GetMessages()
	messages[0].Content = "changed"
	if got := session.GetMessages()[0].Content; got != "hello" {
		t.Fatalf("snapshot mutated session: %q", got)
	}
	loaded, err := workspace.Load(workspace.MemTypeChunk)
	if err != nil {
		t.Fatal(err)
	}
	if got := loaded.([]*schema.Message)[0].Content; got != "hello" {
		t.Fatalf("unexpected persisted content: %q", got)
	}
}

func TestSessionTitle(t *testing.T) {
	session := newTestSession(t)
	if got := session.Title(); got != "New Session" {
		t.Fatalf("unexpected empty title: %q", got)
	}
	if err := session.Append(&schema.Message{Role: schema.RoleUser, Content: "first message"}); err != nil {
		t.Fatal(err)
	}
	if got := session.Title(); got != "first message" {
		t.Fatalf("unexpected title: %q", got)
	}
}
