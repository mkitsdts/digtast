package llm

import (
	"context"
	"encoding/json"
	"io"
	"time"
)

type Role string

const (
	RoleDeveloper Role = "developer"
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

type ContentPartType string

const (
	ContentText       ContentPartType = "text"
	ContentImageURL   ContentPartType = "image_url"
	ContentInputAudio ContentPartType = "input_audio"
	ContentFile       ContentPartType = "file"
)

type ContentPart struct {
	Type     ContentPartType `json:"type"`
	Text     string          `json:"text,omitempty"`
	URL      string          `json:"url,omitempty"`
	Data     string          `json:"data,omitempty"`
	MIMEType string          `json:"mime_type,omitempty"`
	Name     string          `json:"name,omitempty"`
	Detail   string          `json:"detail,omitempty"`
}

type FunctionCall struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type,omitempty"`
	Function FunctionCall `json:"function"`
}

type Message struct {
	Role       Role          `json:"role"`
	Content    string        `json:"content,omitempty"`
	Parts      []ContentPart `json:"parts,omitempty"`
	Name       string        `json:"name,omitempty"`
	ToolCallID string        `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall    `json:"tool_calls,omitempty"`
}

type ToolDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters"`
}

type ToolChoice string

const (
	ToolChoiceNone     ToolChoice = "none"
	ToolChoiceAuto     ToolChoice = "auto"
	ToolChoiceRequired ToolChoice = "required"
)

type ChatRequest struct {
	Model             string
	Messages          []Message
	Tools             []ToolDefinition
	ToolChoice        ToolChoice
	Temperature       *float64
	MaxOutputTokens   int
	ParallelToolCalls *bool
	Metadata          map[string]string
}

type FinishReason string

const (
	FinishStop      FinishReason = "stop"
	FinishToolCalls FinishReason = "tool_calls"
	FinishLength    FinishReason = "length"
	FinishContent   FinishReason = "content_filter"
	FinishError     FinishReason = "error"
)

type Usage struct {
	InputTokens     int `json:"input_tokens"`
	OutputTokens    int `json:"output_tokens"`
	TotalTokens     int `json:"total_tokens"`
	CachedTokens    int `json:"cached_tokens,omitempty"`
	ReasoningTokens int `json:"reasoning_tokens,omitempty"`
}

func (u *Usage) Add(other Usage) {
	u.InputTokens += other.InputTokens
	u.OutputTokens += other.OutputTokens
	u.TotalTokens += other.TotalTokens
	u.CachedTokens += other.CachedTokens
	u.ReasoningTokens += other.ReasoningTokens
}

type ChatResponse struct {
	ID           string
	Model        string
	Message      Message
	FinishReason FinishReason
	Usage        Usage
	CreatedAt    time.Time
}

type StreamEventType string

const (
	StreamEventTextDelta     StreamEventType = "text_delta"
	StreamEventToolCallDelta StreamEventType = "tool_call_delta"
	StreamEventUsage         StreamEventType = "usage"
	StreamEventDone          StreamEventType = "done"
)

type ToolCallDelta struct {
	Index     int
	ID        string
	Type      string
	Name      string
	Arguments string
}

type StreamEvent struct {
	Type     StreamEventType
	Delta    string
	ToolCall *ToolCallDelta
	Usage    *Usage
	Response *ChatResponse
}

type Stream interface {
	Recv() (*StreamEvent, error)
	Close() error
}

type Status struct {
	Available bool
	Model     string
	Latency   time.Duration
	CheckedAt time.Time
}

type LLM interface {
	Chat(ctx context.Context, req *ChatRequest) (*ChatResponse, error)
	ChatStream(ctx context.Context, req *ChatRequest) (Stream, error)
	GetUsage() Usage
	CheckStatus(ctx context.Context) (*Status, error)
}

type SliceStream struct {
	Events []*StreamEvent
	index  int
}

func (s *SliceStream) Recv() (*StreamEvent, error) {
	if s.index >= len(s.Events) {
		return nil, io.EOF
	}
	event := s.Events[s.index]
	s.index++
	return event, nil
}

func (s *SliceStream) Close() error { return nil }
