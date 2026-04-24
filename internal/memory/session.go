package mem

import (
	"encoding/json"
	"fmt"
	"os"
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
	CreatedAt time.Time

	filePath           string
	mu                 sync.Mutex
	messages           []*schema.Message
	pendingInterruptID string // non-empty while the agent is paused awaiting human approval
	msgIdx             int    // A2UI component slot index at the point of last interrupt
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
	s.mu.Lock()
	defer s.mu.Unlock()

	s.messages = append(s.messages, msg)

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(s.filePath, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = fmt.Fprintf(f, "%s\n", data)
	return err
}

// GetMessages returns a snapshot of all messages.
func (s *Session) GetMessages() []*schema.Message {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make([]*schema.Message, len(s.messages))
	copy(result, s.messages)
	return result
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
