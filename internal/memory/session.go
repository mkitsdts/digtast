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
// Each agent owns exactly one session; Session.AgentID == owning agent's ID.
type Session struct {
	AgentID   string
	CreatedAt time.Time

	mu                 sync.Mutex
	messages           []*schema.Message
	pendingInterruptID string
	msgIdx             int

	store *Store

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
	store := s.store
	agentID := s.AgentID
	s.mu.Unlock()

	if store == nil || store.persist == nil {
		return nil
	}
	return store.persist.PersistMessage(agentID, msg)
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
	store := s.store
	agentID := s.AgentID
	s.mu.Unlock()

	if store == nil || store.persist == nil {
		return nil
	}
	_, err := store.persist.NewSessionFile(agentID)
	return err
}

// ResetWithMessages replaces all messages in the session with the given ones,
// clearing old JSONL data and persisting the new messages to a fresh chunk.
func (s *Session) ResetWithMessages(msgs []*schema.Message) error {
	s.mu.Lock()
	s.messages = msgs
	store := s.store
	agentID := s.AgentID
	s.mu.Unlock()

	if store == nil || store.persist == nil {
		return nil
	}

	// Delete old chunks and start fresh
	if err := store.persist.DeleteSession(agentID); err != nil {
		return err
	}
	if _, err := store.persist.EnsureSession(agentID); err != nil {
		return err
	}

	// Persist the new messages
	for _, msg := range msgs {
		if msg != nil {
			if err := store.persist.PersistMessage(agentID, msg); err != nil {
				return err
			}
		}
	}
	return nil
}

// Size returns the estimated token count of all messages (characters / 4).
func (s *Session) Size() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	total := 0
	for _, msg := range s.messages {
		if msg == nil {
			continue
		}
		total += len(msg.Content)
		for _, part := range msg.MultiContent {
			total += len(part.Text)
		}
	}
	return total / 4
}

// DeleteChunks removes all JSONL chunk files for this session (used after compression).
func (s *Session) DeleteChunks() error {
	s.mu.Lock()
	store := s.store
	agentID := s.AgentID
	s.mu.Unlock()

	if store == nil || store.persist == nil {
		return nil
	}
	return store.persist.DeleteSessionChunks(agentID)
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
