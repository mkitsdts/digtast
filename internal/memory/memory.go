package mem

import (
	"bufio"
	"digital-labor/pkg/workspace"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/schema"
)

// Store manages persisted sessions backed by JSONL files.
//
// File format:
//
//	{"type":"session","id":"...","created_at":"..."}   ← header (line 1)
//	{"role":"user","content":"..."}                    ← message (lines 2+)
type Store struct {
	dir   string
	mu    sync.Mutex
	cache map[string]*Session
}

// NewStore creates a new Store backed by the given directory (created if absent).
func NewStore() *Store {
	dir := workspace.GetWorkspacePath() + "/sessions"

	if err := os.MkdirAll(dir, 0o755); err != nil {
		slog.Error("failed to create session dir", "error", err)
		return nil
	}

	return &Store{
		dir:   dir,
		cache: make(map[string]*Session),
	}
}

// GetOrCreate returns the session for id, creating it if it does not exist.
func (s *Store) GetOrCreate(id string) (*Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sess, ok := s.cache[id]; ok {
		return sess, nil
	}

	filePath := filepath.Join(s.dir, id+".jsonl")

	var (
		sess *Session
		err  error
	)
	if _, statErr := os.Stat(filePath); os.IsNotExist(statErr) {
		sess, err = createSession(id, filePath)
	} else {
		sess, err = loadSession(filePath)
	}
	if err != nil {
		return nil, err
	}

	s.cache[id] = sess
	return sess, nil
}

// List returns metadata for all known sessions.
func (s *Store) List() ([]SessionMeta, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return nil, err
	}

	var metas []SessionMeta
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".jsonl") {
			continue
		}
		id := strings.TrimSuffix(e.Name(), ".jsonl")

		if sess, ok := s.cache[id]; ok {
			metas = append(metas, SessionMeta{ID: id, Title: sess.Title(), CreatedAt: sess.CreatedAt})
			continue
		}

		sess, loadErr := loadSession(filepath.Join(s.dir, e.Name()))
		if loadErr != nil {
			continue
		}
		metas = append(metas, SessionMeta{ID: id, Title: sess.Title(), CreatedAt: sess.CreatedAt})
	}
	return metas, nil
}

// Delete removes the session file and evicts it from the cache.
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	filePath := filepath.Join(s.dir, id+".jsonl")
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return err
	}
	delete(s.cache, id)
	return nil
}

// AppendMessage adds a message to the session, merging it with the last message if it has the same role.
func (s *Store) AppendMessage(id string, msg *schema.Message) {
	s.cache[id].mu.Lock()
	defer s.cache[id].mu.Unlock()
	if msg == nil {
		return
	}
	appendMessage(s.cache[id], msg)
}

// sessionHeader is the first JSONL line in every session file.
type sessionHeader struct {
	Type      string    `json:"type"`
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

func createSession(id, filePath string) (*Session, error) {
	header := sessionHeader{
		Type:      "session",
		ID:        id,
		CreatedAt: time.Now().UTC(),
	}
	data, err := json.Marshal(header)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(filePath, append(data, '\n'), 0o644); err != nil {
		return nil, err
	}
	return &Session{
		ID:        id,
		CreatedAt: header.CreatedAt,
		filePath:  filePath,
		messages:  make([]*schema.Message, 0),
	}, nil
}

func loadSession(filePath string) (*Session, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	// First line: header
	if !scanner.Scan() {
		return nil, fmt.Errorf("empty session file: %s", filePath)
	}
	var header sessionHeader
	if err := json.Unmarshal(scanner.Bytes(), &header); err != nil {
		return nil, fmt.Errorf("bad session header in %s: %w", filePath, err)
	}

	sess := &Session{
		ID:        header.ID,
		CreatedAt: header.CreatedAt,
		filePath:  filePath,
		messages:  make([]*schema.Message, 0),
	}

	// Remaining lines: messages
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var msg schema.Message
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue // skip malformed lines
		}
		sess.messages = append(sess.messages, &msg)
	}

	return sess, scanner.Err()
}

func appendMessage(sess *Session, msg *schema.Message) {
	sess.mu.Lock()
	defer sess.mu.Unlock()
	if msg == nil {
		return
	}
	for sess.messages[len(sess.messages)-1].Role == msg.Role {
		mergeMessages(sess.messages, msg)
	}
	sess.messages = append(sess.messages, msg)
}

func mergeMessages(msgs []*schema.Message, msg *schema.Message) {
	// TODO:检查消息详细，并合并全部内容
	// 万万不可单独合并content

	msgs[len(msgs)-1].Content += "\n" + msg.Content
}
