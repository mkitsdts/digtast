package mem

import (
	"testing"

	"digital-labor/config"
	"digital-labor/internal/agent/workspace"
	"digital-labor/pkg/chatmodel"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	config.Conf.WorkSpaceDir = t.TempDir()
	store, err := NewStore()
	if err != nil {
		t.Fatal(err)
	}
	return store
}

func TestStoreCachesAndDeletesSession(t *testing.T) {
	store := newTestStore(t)
	first, err := store.GetOrCreate()
	if err != nil {
		t.Fatal(err)
	}
	second, err := store.GetOrCreate()
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatal("expected the singleton session to be cached")
	}
	if err := first.Append(&chatmodel.Message{Role: chatmodel.RoleUser, Content: "hello"}); err != nil {
		t.Fatal(err)
	}
	if err := store.Delete(); err != nil {
		t.Fatal(err)
	}
	loaded, err := workspace.Load(workspace.MemTypeChunk)
	if err != nil {
		t.Fatal(err)
	}
	if messages := loaded.([]*chatmodel.Message); len(messages) != 0 {
		t.Fatalf("expected deleted history, got %d messages", len(messages))
	}
	fresh, err := store.GetOrCreate()
	if err != nil {
		t.Fatal(err)
	}
	if fresh == first {
		t.Fatal("expected a fresh session after deletion")
	}
}

func TestStoreLoadsPersistedMessages(t *testing.T) {
	store := newTestStore(t)
	if err := workspace.Save(workspace.MemTypeChunk, &chatmodel.Message{Role: chatmodel.RoleUser, Content: "persisted"}); err != nil {
		t.Fatal(err)
	}
	session, err := store.GetOrCreate()
	if err != nil {
		t.Fatal(err)
	}
	messages := session.GetMessages()
	if len(messages) != 1 || messages[0].Content != "persisted" {
		t.Fatalf("unexpected messages: %#v", messages)
	}
}
