package workspace

import (
	"bufio"
	"digital-labor/pkg/conf"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
)

const defaultAgentID = "default"

// DefaultAgentID is used by callers that have not yet been wired to an agent identity.
func DefaultAgentID() string {
	return defaultAgentID
}

// MaxMemoryChunkSize returns the configured chunk rotation threshold.
func MaxMemoryChunkSize() int64 {
	return conf.Conf.Memory.MaxMessagesSize
}

// MemoryStore owns the workspace memory index and the session chunk files.
type MemoryStore struct {
	root  string
	index *MemoryIndex

	// mu protects the in-memory index. File appends are intentionally done
	// outside this lock so long message writes do not block metadata reads.
	mu sync.RWMutex

	// flushMu serializes memory.json writes. Index mutations can request a
	// flush concurrently, but only one atomic replace should run at a time.
	flushMu sync.Mutex

	// flushReq is a coalescing signal: many metadata updates can collapse into
	// one periodic flush, which keeps message writes from syncing memory.json
	// on every append.
	flushReq chan struct{}

	// stop is reserved for a future explicit shutdown path that can force a
	// final index flush before the process exits.
	stop chan struct{}
}

// MemoryIndex is persisted to memory.json.
// Keyed by agentID — each agent owns exactly one session.
type MemoryIndex struct {
	Sessions map[string]*SessionRecord `json:"sessions"`
}

// SessionRecord maps one agent to the chunk files that contain its conversation.
type SessionRecord struct {
	AgentID string        `json:"agent_id"`
	Chunks  []ChunkRecord `json:"chunks"`
	Created time.Time     `json:"created"`
	Updated time.Time     `json:"updated"`
}

// ChunkRecord describes one JSONL file for a session.
type ChunkRecord struct {
	ID      string    `json:"id"`
	Path    string    `json:"path"`
	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

var (
	defaultMemoryOnce  sync.Once
	defaultMemoryStore *MemoryStore
	defaultMemoryErr   error
)

// NewMemoryStore loads or creates the memory store rooted at root.
func NewMemoryStore(root string) (*MemoryStore, error) {
	if root == "" {
		return nil, errors.New("memory root is empty")
	}
	if err := os.MkdirAll(filepath.Join(root, "sessions"), 0o755); err != nil {
		return nil, err
	}

	store := &MemoryStore{
		root: root,
		index: &MemoryIndex{
			Sessions: make(map[string]*SessionRecord),
		},
		flushReq: make(chan struct{}, 1),
		stop:     make(chan struct{}),
	}
	if err := store.loadIndex(); err != nil {
		return nil, err
	}
	go store.flushLoop()
	return store, nil
}

// DefaultMemoryStore returns the process-wide workspace memory store.
func DefaultMemoryStore() (*MemoryStore, error) {
	defaultMemoryOnce.Do(func() {
		defaultMemoryStore, defaultMemoryErr = NewMemoryStore(filepath.Join(GetWorkspacePath(), "memory"))
	})
	return defaultMemoryStore, defaultMemoryErr
}

// NewSessionFile allocates a new chunk for an agent without creating its JSONL file.
func (s *MemoryStore) NewSessionFile(agentID string) (string, error) {
	if err := validateID(agentID, "agent_id"); err != nil {
		return "", err
	}

	now := time.Now().UTC()
	chunkID := uuid.New().String()
	relPath := filepath.Join("sessions", agentID, fmt.Sprintf("%s.jsonl", chunkID))

	s.mu.Lock()
	defer s.mu.Unlock()

	record := s.index.Sessions[agentID]
	if record == nil {
		record = &SessionRecord{
			AgentID: agentID,
			Created: now,
		}
		s.index.Sessions[agentID] = record
	}

	record.Chunks = append(record.Chunks, ChunkRecord{
		ID:      chunkID,
		Path:    filepath.ToSlash(relPath),
		Created: now,
		Updated: now,
	})
	record.Updated = now
	s.requestFlush()
	return chunkID, nil
}

// EnsureSession returns the current chunk id, creating only the index entry when needed.
func (s *MemoryStore) EnsureSession(agentID string) (string, error) {
	s.mu.RLock()
	record := s.index.Sessions[agentID]
	if record != nil && len(record.Chunks) > 0 {
		chunkID := record.Chunks[len(record.Chunks)-1].ID
		s.mu.RUnlock()
		return chunkID, nil
	}
	s.mu.RUnlock()
	return s.NewSessionFile(agentID)
}

// CurrentChunkSize returns the current chunk file size. Missing files are empty.
func (s *MemoryStore) CurrentChunkSize(agentID string) (int64, error) {
	s.mu.RLock()
	record := s.index.Sessions[agentID]
	if record == nil || len(record.Chunks) == 0 {
		s.mu.RUnlock()
		return 0, nil
	}
	chunk := record.Chunks[len(record.Chunks)-1]
	path := s.chunkAbsPath(chunk)
	s.mu.RUnlock()

	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}
	return info.Size(), nil
}

// PersistMessage appends msg to the current chunk, creating the JSONL file on first write.
func (s *MemoryStore) PersistMessage(agentID string, msg *schema.Message) error {
	if msg == nil {
		return nil
	}
	if _, err := s.EnsureSession(agentID); err != nil {
		return err
	}

	s.mu.Lock()
	record := s.index.Sessions[agentID]
	if record == nil || len(record.Chunks) == 0 {
		s.mu.Unlock()
		return errors.New("session has no chunk")
	}
	chunkIdx := len(record.Chunks) - 1
	chunk := record.Chunks[chunkIdx]
	path := s.chunkAbsPath(chunk)
	s.mu.Unlock()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	if _, err := file.Write(append(data, '\n')); err != nil {
		return err
	}

	now := time.Now().UTC()
	s.mu.Lock()
	if record := s.index.Sessions[agentID]; record != nil && len(record.Chunks) > chunkIdx {
		record.Chunks[chunkIdx].Updated = now
		record.Updated = now
	}
	s.mu.Unlock()
	s.requestFlush()
	return nil
}

// LoadSession reads all chunk files for an agent, skipping malformed message lines.
func (s *MemoryStore) LoadSession(agentID string) ([]*schema.Message, error) {
	s.mu.RLock()
	record := s.index.Sessions[agentID]
	if record == nil {
		s.mu.RUnlock()
		return nil, nil
	}
	chunks := append([]ChunkRecord(nil), record.Chunks...)
	s.mu.RUnlock()

	var messages []*schema.Message
	for _, chunk := range chunks {
		chunkMessages, err := s.loadChunk(chunk)
		if err != nil {
			return nil, err
		}
		messages = append(messages, chunkMessages...)
	}
	return messages, nil
}

// DeleteSession removes all chunk files and the index entry for an agent.
func (s *MemoryStore) DeleteSession(agentID string) error {
	s.mu.Lock()
	record := s.index.Sessions[agentID]
	if record == nil {
		s.mu.Unlock()
		return nil
	}
	chunks := append([]ChunkRecord(nil), record.Chunks...)
	delete(s.index.Sessions, agentID)
	s.mu.Unlock()

	for _, chunk := range chunks {
		if err := os.Remove(s.chunkAbsPath(chunk)); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	s.requestFlush()
	return nil
}

// DeleteSessionChunks removes all chunk files for an agent but keeps the index entry.
// Used after compression to clear original history while preserving the session record.
func (s *MemoryStore) DeleteSessionChunks(agentID string) error {
	s.mu.Lock()
	record := s.index.Sessions[agentID]
	if record == nil {
		s.mu.Unlock()
		return nil
	}
	chunks := append([]ChunkRecord(nil), record.Chunks...)
	record.Chunks = nil
	record.Updated = time.Now().UTC()
	s.mu.Unlock()

	for _, chunk := range chunks {
		if err := os.Remove(s.chunkAbsPath(chunk)); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	s.requestFlush()
	return nil
}

// ListSessions returns session records for one agent (at most one in the 1:1 model).
func (s *MemoryStore) ListSessions(agentID string) []SessionRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	record := s.index.Sessions[agentID]
	if record == nil {
		return nil
	}
	copyRecord := *record
	copyRecord.Chunks = append([]ChunkRecord(nil), record.Chunks...)
	return []SessionRecord{copyRecord}
}

// Flush writes pending index changes to memory.json.
func (s *MemoryStore) Flush() error {
	s.mu.RLock()
	data, err := json.MarshalIndent(s.index, "", "  ")
	s.mu.RUnlock()
	if err != nil {
		return err
	}

	s.flushMu.Lock()
	defer s.flushMu.Unlock()

	if err := os.MkdirAll(s.root, 0o755); err != nil {
		return err
	}
	tmp := s.indexPath() + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.indexPath())
}

func (s *MemoryStore) loadIndex() error {
	data, err := os.ReadFile(s.indexPath())
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, s.index); err != nil {
		return err
	}
	if s.index.Sessions == nil {
		s.index.Sessions = make(map[string]*SessionRecord)
	}
	return nil
}

func (s *MemoryStore) loadChunk(chunk ChunkRecord) ([]*schema.Message, error) {
	file, err := os.Open(s.chunkAbsPath(chunk))
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var messages []*schema.Message
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var msg schema.Message
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue
		}
		messages = append(messages, &msg)
	}
	return messages, scanner.Err()
}

func (s *MemoryStore) flushLoop() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	dirty := false
	for {
		select {
		case <-s.flushReq:
			dirty = true
		case <-ticker.C:
			if dirty {
				if err := s.Flush(); err != nil {
					slog.Error("failed to flush memory index", "error", err)
				}
				dirty = false
			}
		case <-s.stop:
			if dirty {
				if err := s.Flush(); err != nil {
					slog.Error("failed to flush memory index", "error", err)
				}
			}
			return
		}
	}
}

func (s *MemoryStore) requestFlush() {
	select {
	case s.flushReq <- struct{}{}:
	default:
	}
}

func (s *MemoryStore) chunkAbsPath(chunk ChunkRecord) string {
	return filepath.Join(s.root, filepath.FromSlash(chunk.Path))
}

func (s *MemoryStore) indexPath() string {
	return filepath.Join(s.root, "memory.json")
}

func validateID(id, name string) error {
	if id == "" {
		return fmt.Errorf("%s is empty", name)
	}
	if strings.ContainsAny(id, `/\`) {
		return fmt.Errorf("%s contains path separator", name)
	}
	return nil
}

// Backward-compatible package helpers use the default store.
func NewSessionFile(agentID string) (string, error) {
	store, err := DefaultMemoryStore()
	if err != nil {
		return "", err
	}
	return store.NewSessionFile(agentID)
}

func PersistMessage(agentID string, msg *schema.Message) {
	store, err := DefaultMemoryStore()
	if err != nil {
		slog.Error("failed to open memory store", "error", err)
		return
	}
	if err := store.PersistMessage(agentID, msg); err != nil {
		slog.Error("failed to persist message", "agent_id", agentID, "error", err)
	}
}

// AgentMemoryPath returns the path to an agent's memory.md file.
func AgentMemoryPath(agentID string) string {
	return filepath.Join(GetWorkspacePath(), "memory", agentID, "memory.md")
}

// LoadAgentMemory reads the agent's memory.md content. Returns empty string if not found.
func LoadAgentMemory(agentID string) string {
	data, err := os.ReadFile(AgentMemoryPath(agentID))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// SaveAgentMemory appends content to the agent's memory.md file.
func SaveAgentMemory(agentID, content string) error {
	path := AgentMemoryPath(agentID)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create memory dir: %w", err)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open memory file: %w", err)
	}
	defer f.Close()
	if _, err := f.WriteString(content); err != nil {
		return fmt.Errorf("write memory: %w", err)
	}
	return nil
}

// ReplaceAgentMemory overwrites the agent's memory.md with new content.
func ReplaceAgentMemory(agentID, content string) error {
	path := AgentMemoryPath(agentID)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create memory dir: %w", err)
	}
	return os.WriteFile(path, []byte(content), 0644)
}
