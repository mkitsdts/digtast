package mem

import (
	"sync"
	"time"

	"github.com/cloudwego/eino/schema"
)

// SessionMeta provides summary info for the session list.
type SessionMeta struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
}

// Session holds the in-memory state for a single conversation.
type Session struct {
	ID        string
	AgentID   string
	CreatedAt time.Time

	mu                 sync.Mutex
	messages           []*schema.Message
	pendingInterruptID string // non-empty while the agent is paused awaiting human approval
	msgIdx             int    // A2UI component slot index at the point of last interrupt

	// store links the runtime session back to its owner so Append can delegate
	// persistence without exposing workspace details to callers.
	store *Store

	// rotateAfterTurn is set when the current chunk is already over the size
	// limit. The next chunk is allocated only after the active task completes.
	rotateAfterTurn bool
}

// SetPendingInterruptID saves the interrupt ID so the approve endpoint can resume it.
func (s *Session) SetPendingInterruptID(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pendingInterruptID = id
}

// GetPendingInterruptID returns the stored interrupt ID, or "" if none is pending.
func (s *Session) GetPendingInterruptID() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pendingInterruptID
}

// SetMsgIdx stores the A2UI component slot counter so a resume can continue from it.
func (s *Session) SetMsgIdx(idx int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.msgIdx = idx
}

// GetMsgIdx returns the stored component slot counter.
func (s *Session) GetMsgIdx() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.msgIdx
}

// Append adds a message to memory and persists it to disk.
func (s *Session) Append(msg *schema.Message) error {
	if msg == nil {
		return nil
	}

	s.mu.Lock()
	s.messages = append(s.messages, msg)
	// Copy the persistence handles while holding the lock, then release it
	// before doing disk I/O.
	store := s.store
	agentID := s.AgentID
	sessionID := s.ID
	s.mu.Unlock()

	if store == nil || store.persist == nil {
		return nil
	}
	return store.persist.PersistMessage(agentID, sessionID, msg)
}

// GetMessages returns a snapshot of all messages.
func (s *Session) GetMessages() []*schema.Message {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]*schema.Message, len(s.messages))
	for i, msg := range s.messages {
		if msg == nil {
			continue
		}
		// Return message copies so callers cannot mutate the session cache.
		cp := *msg
		result[i] = &cp
	}
	return result
}

// CompleteTurn rotates to a new chunk after the current task finishes when the previous chunk is too large.
func (s *Session) CompleteTurn() error {
	s.mu.Lock()
	if !s.rotateAfterTurn {
		s.mu.Unlock()
		return nil
	}
	s.rotateAfterTurn = false
	// Capture the store/session identifiers under the lock, then rotate the
	// persistent chunk without holding the session mutex.
	store := s.store
	agentID := s.AgentID
	sessionID := s.ID
	s.mu.Unlock()

	if store == nil || store.persist == nil {
		return nil
	}
	_, err := store.persist.NewSessionFile(agentID, sessionID)
	return err
}

// Title derives a display title from the first user message.
func (s *Session) Title() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, msg := range s.messages {
		if msg.Role == schema.User && msg.Content != "" {
			title := msg.Content
			if len([]rune(title)) > 60 {
				title = string([]rune(title)[:60]) + "..."
			}
			return title
		}
	}
	return "New Session"
}
