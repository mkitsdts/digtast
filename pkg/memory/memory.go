package mem

import (
	"digital-labor/pkg/workspace"
	"fmt"
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
}

// NewStore creates an in-memory session store backed by the workspace memory functions.
func NewStore(agentID string) (*Store, error) {
	if agentID == "" {
		agentID = workspace.DefaultAgentID()
	}

	return &Store{
		agentID: agentID,
	}, nil
}

// GetOrCreate returns the agent's session, creating it if it does not exist.
func (s *Store) GetOrCreate() (*Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// if exist, use cache
	if s.session != nil {
		return s.session, nil
	}

	res, err := workspace.Load(s.agentID, workspace.MemTypeChunk)
	if err != nil {
		return nil, err
	}
	messages, ok := res.([]*schema.Message)
	if !ok {
		return nil, fmt.Errorf("unexpected type from load: %T", res)
	}

	sess := &Session{
		AgentID:   s.agentID,
		CreatedAt: time.Now().UTC(),
		messages:  messages,
	}
	s.session = sess
	return sess, nil
}

// Delete clears the agent's session history and evicts it from the cache.
func (s *Store) Delete() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := workspace.DeleteSession(s.agentID); err != nil {
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
