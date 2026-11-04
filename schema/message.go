package schema

import (
	"encoding/json"
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
	Reasoning  string        `json:"reasoning,omitempty"` // 推理模型的思维链，仅供本地展示，回传请求时由适配器丢弃
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

// ChatRequest 是非流式对话请求。
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

// ChatResponse 是非流式对话响应。
type ChatResponse struct {
	ID           string
	Model        string
	Message      Message
	FinishReason FinishReason
	Usage        Usage
	CreatedAt    time.Time
}
