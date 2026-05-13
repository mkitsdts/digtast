package mem

import (
	"digital-labor/pkg/workspace"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
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
	agentID := s.AgentID
	s.mu.Unlock()

	return workspace.Save(agentID, workspace.MemTypeChunk, msg)
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

// GetCompressMessages returns the compacted view used as model context:
// compressed summaries followed by original messages not yet summarized.
func (s *Session) GetCompressMessages() []*schema.Message {
	s.mu.Lock()
	agentID := s.AgentID
	fallback := cloneSchemaMessages(s.messages)
	s.mu.Unlock()

	resMsgs, err := workspace.Load(agentID, workspace.MemTypeChunk)
	if err != nil {
		return fallback
	}
	messages := resMsgs.([]*schema.Message)

	resRecs, err := workspace.Load(agentID, workspace.MemTypeCompress)
	if err != nil {
		return messages
	}
	records := resRecs.([]workspace.CompressionRecord)

	if len(records) == 0 {
		return messages
	}

	var result []*schema.Message
	compressedThrough := 0
	for _, record := range records {
		if record.SourceEndLine <= compressedThrough {
			continue
		}
		result = append(result, cloneSchemaMessages(record.Messages)...)
		compressedThrough = record.SourceEndLine
	}
	if compressedThrough < len(messages) {
		result = append(result, cloneSchemaMessages(messages[compressedThrough:])...)
	}
	return result
}

// CompleteTurn is now a no-op as storage handles rotation internally.
func (s *Session) CompleteTurn() error {
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
		for _, part := range msg.UserInputMultiContent {
			total += len(part.Text)
			if part.Image != nil && part.Image.Base64Data != nil {
				total += len(*part.Image.Base64Data)
			}
			if part.Audio != nil && part.Audio.Base64Data != nil {
				total += len(*part.Audio.Base64Data)
			}
			if part.Video != nil && part.Video.Base64Data != nil {
				total += len(*part.Video.Base64Data)
			}
			if part.File != nil && part.File.Base64Data != nil {
				total += len(*part.File.Base64Data)
			}
		}
	}
	return total / 4
}

// DeleteChunks archives JSONL chunk files for this session to tar.gz and deletes originals.
func (s *Session) DeleteChunks() error {
	s.mu.Lock()
	agentID := s.AgentID
	s.mu.Unlock()

	return workspace.Archive(agentID, workspace.MemTypeChunk)
}

// Search retrieves short-term memories (compressed records) that match the query.
func (s *Session) Search(query string) ([]*schema.Message, error) {
	s.mu.Lock()
	agentID := s.AgentID
	s.mu.Unlock()

	res, err := workspace.Load(agentID, workspace.MemTypeCompress)
	if err != nil {
		return nil, err
	}
	records := res.([]workspace.CompressionRecord)

	var result []*schema.Message
	for _, record := range records {
		match := false
		for _, msg := range record.Messages {
			if strings.Contains(strings.ToLower(msg.Content), strings.ToLower(query)) {
				match = true
				break
			}
		}
		if match {
			result = append(result, cloneSchemaMessages(record.Messages)...)
		}
	}
	return result, nil
}

// AppendCompression persists a compressed segment without mutating full history.
func (s *Session) AppendCompression(startLine, endLine int, msgs []*schema.Message) error {
	s.mu.Lock()
	agentID := s.AgentID
	s.mu.Unlock()

	record := workspace.CompressionRecord{
		ID:              uuid.New().String(),
		SourceStartLine: startLine,
		SourceEndLine:   endLine,
		Messages:        cloneSchemaMessages(msgs),
		Created:         time.Now().UTC(),
	}
	return workspace.Save(agentID, workspace.MemTypeCompress, record)
}

// CompressedThrough returns the highest original line number covered by compression.
func (s *Session) CompressedThrough() int {
	s.mu.Lock()
	agentID := s.AgentID
	s.mu.Unlock()

	res, err := workspace.Load(agentID, workspace.MemTypeCompress)
	if err != nil {
		return 0
	}
	records := res.([]workspace.CompressionRecord)
	maxLine := 0
	for _, record := range records {
		if record.SourceEndLine > maxLine {
			maxLine = record.SourceEndLine
		}
	}
	return maxLine
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

func cloneSchemaMessages(messages []*schema.Message) []*schema.Message {
	result := make([]*schema.Message, len(messages))
	for i, msg := range messages {
		if msg == nil {
			continue
		}
		cp := *msg
		result[i] = &cp
	}
	return result
}
