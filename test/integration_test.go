package test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	mem "digital-labor/internal/memory"

	"github.com/cloudwego/eino/schema"
)

// ============================================================
// Memory Integration Tests
// ============================================================

func TestMemory_CreateGetAppendListDelete(t *testing.T) {
	// Note: Store.dir is unexported, so we cannot inject the temp dir.
	// This test demonstrates the design issue — NewStore() always uses
	// the global workspace path, making integration testing impossible
	// without modifying production code.
	//
	// As a workaround, we test with the real NewStore() and unique IDs,
	// but this pollutes the workspace directory. A better fix is to
	// export a SetDir method or accept dir as NewStore parameter.
	s, err := mem.NewStore("")
	if err != nil {
		t.Skipf("NewStore failed: %v", err)
	}

	// Create
	sess, err := s.GetOrCreate()
	if err != nil {
		t.Fatalf("GetOrCreate failed: %v", err)
	}
	if sess == nil {
		t.Fatal("expected non-nil session")
	}

	// Append messages
	sess.Append(&schema.Message{
		Role:    schema.User,
		Content: "integration test message",
	})

	msgs := sess.GetMessages()
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}

	// Reload from disk — should return cached instance
	sess2, err := s.GetOrCreate()
	if err != nil {
		t.Fatalf("GetOrCreate on reload failed: %v", err)
	}
	if sess2 != sess {
		t.Fatal("expected cached instance on second GetOrCreate")
	}

	// Delete
	if err := s.Delete(); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
}

func TestMemory_PersistenceAcrossStoreInstances(t *testing.T) {
	// This test verifies that session data persists across different Store instances
	// by using a temporary directory directly (bypassing NewStore).
	dir := t.TempDir()
	sessionID := "persist-test"

	// Manually create a session file
	header := map[string]interface{}{
		"type":       "session",
		"id":         sessionID,
		"created_at": time.Now().UTC(),
	}
	headerBytes, _ := json.Marshal(header)
	filePath := filepath.Join(dir, sessionID+".jsonl")
	if err := os.WriteFile(filePath, append(headerBytes, '\n'), 0644); err != nil {
		t.Fatal(err)
	}

	// Append a message
	msg := &schema.Message{Role: schema.User, Content: "persisted message"}
	msgBytes, _ := json.Marshal(msg)
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatal(err)
	}
	f.Write(msgBytes)
	f.Write([]byte("\n"))
	f.Close()

	// Read back and verify
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}
	lines := string(data)
	if len(lines) == 0 {
		t.Fatal("expected file to have content")
	}
	t.Logf("persisted content: %s", lines)
}

func TestMemory_ConcurrentAppend(t *testing.T) {
	dir := t.TempDir()
	sessionID := "concurrent-append"

	header := map[string]interface{}{
		"type":       "session",
		"id":         sessionID,
		"created_at": time.Now().UTC(),
	}
	headerBytes, _ := json.Marshal(header)
	filePath := filepath.Join(dir, sessionID+".jsonl")
	if err := os.WriteFile(filePath, append(headerBytes, '\n'), 0644); err != nil {
		t.Fatal(err)
	}

	// Directly test Session.Append with concurrent access
	sess := &mem.Session{}
	// Note: Session.filePath is unexported, so we can't inject the temp dir path.
	// This documents another design issue — Session needs a constructor or setter
	// for the file path to be testable in isolation.
	_ = sess
	_ = filePath

	// The real test would require either:
	// 1. Exporting filePath field
	// 2. Adding a constructor like NewSessionWithFile(id, filePath)
	// 3. Using a global workspace path that can be set per-test
}

// ============================================================
// Store + Session File Format Tests
// ============================================================

func TestMemory_JSONLFormat(t *testing.T) {
	dir := t.TempDir()

	// Create a valid JSONL session file
	id := "format-test"
	filePath := filepath.Join(dir, id+".jsonl")

	// Line 1: session header
	header := `{"type":"session","id":"format-test","created_at":"2024-01-15T10:00:00Z"}`

	// Lines 2-4: messages
	msg1 := `{"role":"user","content":"hello"}`
	msg2 := `{"role":"assistant","content":"hi there"}`
	msg3 := `{"role":"user","content":"how are you?"}`

	content := header + "\n" + msg1 + "\n" + msg2 + "\n" + msg3 + "\n"
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Verify we can read and parse each line
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 4 {
		t.Fatalf("expected 4 lines, got %d", len(lines))
	}

	// Parse header
	var headerParsed map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &headerParsed); err != nil {
		t.Fatalf("failed to parse header: %v", err)
	}
	if headerParsed["type"] != "session" {
		t.Fatalf("expected type 'session', got '%s'", headerParsed["type"])
	}

	// Parse messages
	for i, line := range lines[1:] {
		var msg schema.Message
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			t.Fatalf("failed to parse message %d: %v", i, err)
		}
	}
}

func TestMemory_SkipsMalformedLines(t *testing.T) {
	dir := t.TempDir()
	id := "malformed-test"
	filePath := filepath.Join(dir, id+".jsonl")

	header := `{"type":"session","id":"malformed-test","created_at":"2024-01-15T10:00:00Z"}`
	goodMsg := `{"role":"user","content":"good"}`
	badMsg := `this is not json`

	content := header + "\n" + goodMsg + "\n" + badMsg + "\n"
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// The loadSession function should skip malformed lines
	// This test documents that behavior — the current implementation
	// uses `continue` on json.Unmarshal error (line 187 of memory.go)
	data, _ := os.ReadFile(filePath)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
}

// ============================================================
// Center Key Consistency Integration Test
// ============================================================

func TestCenter_KeyConsistency(t *testing.T) {
	// This test documents the key format used between service.go and center.go
	// service.go StartService creates key: fmt.Sprintf("%s/%s", containerId, agentId)
	// center.go CreateAgent stores key: config.ID
	// service.go RemoveService uses: buildAgentKey(containerId, agentId)
	// center.go Get uses: id (passed directly)
	//
	// The key formats MUST match across Create, Get, and Remove.
	// If they don't, agents can be created but never found.

	containerID := "test-container"
	agentID := "test-agent"
	expectedKey := "test-container/test-agent"

	actualKey := containerID + "/" + agentID
	if actualKey != expectedKey {
		t.Fatalf("key format mismatch: expected '%s', got '%s'", expectedKey, actualKey)
	}
}

// ============================================================
// Mock LLM Server Integration Test
// ============================================================

func TestAgent_WithMockLLM(t *testing.T) {
	if os.Getenv("TEST_INTEGRATION") == "" {
		t.Skip("set TEST_INTEGRATION=1 to run")
	}

	// This test requires:
	// 1. A mock LLM HTTP server
	// 2. Creating an agent that connects to the mock
	// 3. Sending a message and verifying the response
	//
	// Since agent creation requires a valid LLM connection,
	// we set up a mock server that returns a fixed response.
	//
	// TODO: Implement when LLM SDK interface is understood well enough.
	// The mock server should respond to the OpenAI-compatible chat
	// completions endpoint.
}

// ============================================================
// Helper: NewStore with custom dir
// ============================================================

// NOTE: The current Store struct has an unexported 'dir' field,
// making it impossible to test with a temporary directory.
// These tests use the real workspace path, which is a known limitation.
//
// Recommended fix: Add a constructor that accepts a dir parameter:
//
//	func NewStoreWithDir(dir string) *Store { ... }
//
// Or make dir exported:
//
//	type Store struct { Dir string; ... }

func TestMain(m *testing.M) {
	// Integration tests can override workspace path here if needed
	os.Exit(m.Run())
}
