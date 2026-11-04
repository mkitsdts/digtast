package mem

import (
	"digital-labor/internal/agent/workspace"
	"digital-labor/schema"
	"fmt"
	"sync"
	"time"
)

// Store manages the singleton Agent's conversation session.
type Store struct {
	mu      sync.Mutex
	session *Session
}

func NewStore() (*Store, error) {
	return &Store{}, nil
}

// GetOrCreate returns the agent's session, creating it if it does not exist.
func (s *Store) GetOrCreate() (*Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// if exist, use cache
	if s.session != nil {
		return s.session, nil
	}

	res, err := workspace.Load(workspace.MemTypeChunk)
	if err != nil {
		return nil, err
	}
	messages, ok := res.([]*schema.Message)
	if !ok {
		return nil, fmt.Errorf("unexpected type from load: %T", res)
	}

	sess := &Session{
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

	if err := workspace.DeleteSession(); err != nil {
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
