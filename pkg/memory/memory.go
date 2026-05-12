package mem

import (
	"digital-labor/pkg/workspace"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/cloudwego/eino/schema"
)

// Store manages the runtime session for one agent and delegates persistence to pkg/workspace.
// Each agent owns exactly one session; the session ID equals the agent ID.
type Store struct {
	agentID string

	mu      sync.Mutex
	session *Session

	persist *workspace.MemoryStore
}

// NewStore creates an in-memory session store backed by the workspace memory store.
func NewStore(agentID string) (*Store, error) {
	if agentID == "" {
		agentID = workspace.DefaultAgentID()
	}

	persist, err := workspace.DefaultMemoryStore()
	if err != nil {
		slog.Error("failed to open memory store", "error", err)
		return nil, fmt.Errorf("failed to open memory store: %w", err)
	}

	return &Store{
		agentID: agentID,
		persist: persist,
	}, nil
}

// GetOrCreate returns the agent's session, creating it if it does not exist.
func (s *Store) GetOrCreate() (*Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// if exist, use cache
	if s.session != nil {
		s.markRotateIfNeeded(s.session)
		return s.session, nil
	}

	if _, err := s.persist.EnsureSession(s.agentID); err != nil {
		return nil, err
	}
	messages, err := s.persist.LoadSession(s.agentID)
	if err != nil {
		return nil, err
	}

	sess := &Session{
		AgentID:   s.agentID,
		CreatedAt: time.Now().UTC(),
		store:     s,
		messages:  messages,
	}
	s.session = sess
	s.markRotateIfNeeded(sess)
	return sess, nil
}

// Delete clears the agent's session history and evicts it from the cache.
func (s *Store) Delete() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.persist.DeleteSession(s.agentID); err != nil {
		return err
	}
	s.session = nil
	return nil
}

// AppendMessage adds a message to the agent's session.
func (s *Store) AppendMessage(msg *schema.Message) {
	if msg == nil {
		return
	}
	s.mu.Lock()
	sess := s.session
	s.mu.Unlock()
	if sess == nil {
		return
	}
	_ = sess.Append(msg)
}

// to mark the session for rotation after the current chunk size exceeds the limit
func (s *Store) markRotateIfNeeded(sess *Session) {
	if sess == nil {
		return
	}
	limit := maxChunkSize()
	if limit <= 0 {
		return
	}
	size, err := s.persist.CurrentChunkSize(s.agentID)
	if err != nil || size <= limit {
		return
	}
	sess.mu.Lock()
	sess.rotateAfterTurn = true
	sess.mu.Unlock()
}

func maxChunkSize() int64 {
	return workspace.MaxMemoryChunkSize()
}
