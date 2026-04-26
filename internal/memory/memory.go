package mem

import (
	"digital-labor/pkg/workspace"
	"log/slog"
	"sync"
	"time"

	"github.com/cloudwego/eino/schema"
)

// Store manages runtime sessions and delegates persistence to pkg/workspace.
type Store struct {
	agentID string

	// mu protects the runtime session cache. Persistent metadata is protected
	// by the workspace.MemoryStore below.
	mu    sync.Mutex
	cache map[string]*Session

	// persist is the only layer that knows about memory.json and JSONL chunk
	// files. Store keeps the chat-facing API in memory terms.
	persist *workspace.MemoryStore
}

// NewStore creates an in-memory session store backed by the workspace memory store.
func NewStore(agentIDs ...string) *Store {
	agentID := workspace.DefaultAgentID()
	if len(agentIDs) > 0 && agentIDs[0] != "" {
		agentID = agentIDs[0]
	}

	persist, err := workspace.DefaultMemoryStore()
	if err != nil {
		slog.Error("failed to open memory store", "error", err)
		return nil
	}

	return &Store{
		agentID: agentID,
		cache:   make(map[string]*Session),
		persist: persist,
	}
}

// GetOrCreate returns the session for id, creating it if it does not exist.
func (s *Store) GetOrCreate(id string) (*Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sess, ok := s.cache[id]; ok {
		s.markRotateIfNeeded(sess)
		return sess, nil
	}

	if _, err := s.persist.EnsureSession(s.agentID, id); err != nil {
		return nil, err
	}
	// Rehydrate the runtime session from all persisted chunks before returning
	// it to the agent.
	messages, err := s.persist.LoadSession(s.agentID, id)
	if err != nil {
		return nil, err
	}

	sess := &Session{
		ID:        id,
		AgentID:   s.agentID,
		CreatedAt: time.Now().UTC(),
		store:     s,
		messages:  messages,
	}
	s.cache[id] = sess
	s.markRotateIfNeeded(sess)
	return sess, nil
}

// List returns metadata for all known sessions.
func (s *Store) List() ([]SessionMeta, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	records := s.persist.ListSessions(s.agentID)
	metas := make([]SessionMeta, 0, len(records))
	for _, record := range records {
		if sess, ok := s.cache[record.SessionID]; ok {
			metas = append(metas, SessionMeta{ID: record.SessionID, Title: sess.Title(), CreatedAt: sess.CreatedAt})
		} else {
			metas = append(metas, SessionMeta{ID: record.SessionID, Title: "New Session", CreatedAt: record.Created})
		}
	}
	return metas, nil
}

// Delete removes the session file and evicts it from the cache.
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.persist.DeleteSession(s.agentID, id); err != nil {
		return err
	}
	delete(s.cache, id)
	return nil
}

// AppendMessage adds a message to a cached session.
func (s *Store) AppendMessage(id string, msg *schema.Message) {
	if msg == nil {
		return
	}
	s.mu.Lock()
	sess := s.cache[id]
	s.mu.Unlock()
	if sess == nil {
		return
	}
	_ = sess.Append(msg)
}

func (s *Store) markRotateIfNeeded(sess *Session) {
	if sess == nil {
		return
	}
	limit := maxChunkSize()
	if limit <= 0 {
		return
	}
	size, err := s.persist.CurrentChunkSize(s.agentID, sess.ID)
	if err != nil || size <= limit {
		return
	}
	// Rotation is delayed until CompleteTurn so the current user/assistant
	// exchange stays in one chunk even when the previous file is already large.
	sess.mu.Lock()
	sess.rotateAfterTurn = true
	sess.mu.Unlock()
}

func maxChunkSize() int64 {
	return workspace.MaxMemoryChunkSize()
}

func mergeMessages(msgs []*schema.Message, msg *schema.Message) {
	// TODO: inspect the full message payload and merge every relevant field.
	// Do not merge Content alone once tool calls or structured parts are used.

	msgs[len(msgs)-1].Content += "\n" + msg.Content
}
