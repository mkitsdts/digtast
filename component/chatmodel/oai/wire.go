package oai

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"digital-labor/schema"
)

type StreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type Message struct {
	Role       string     `json:"role"`
	Content    any        `json:"content,omitempty"`
	Name       string     `json:"name,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
	// Reasoning 是推理模型（如 deepseek-reasoner）回传的思维链。只在响应
	// 解码时填充；回传历史消息时必须丢弃，否则部分 Provider 会拒绝请求。
	Reasoning string `json:"reasoning_content,omitempty"`
}

type ContentPart struct {
	Type       string      `json:"type"`
	Text       string      `json:"text,omitempty"`
	ImageURL   *ImageURL   `json:"image_url,omitempty"`
	InputAudio *InputAudio `json:"input_audio,omitempty"`
	File       *File       `json:"file,omitempty"`
}

type ImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

type InputAudio struct {
	Data   string `json:"data"`
	Format string `json:"format"`
}

type File struct {
	FileData string `json:"file_data,omitempty"`
	FileID   string `json:"file_id,omitempty"`
	Filename string `json:"filename,omitempty"`
}

type Tool struct {
	Type     string             `json:"type"`
	Function FunctionDefinition `json:"function"`
}

type FunctionDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters"`
}

type ToolCall struct {
	Index    int          `json:"index,omitempty"`
	ID       string       `json:"id,omitempty"`
	Type     string       `json:"type,omitempty"`
	Function FunctionCall `json:"function"`
}

type FunctionCall struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

type Completion struct {
	ID      string   `json:"id"`
	Model   string   `json:"model"`
	Created int64    `json:"created"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	Delta        Delta   `json:"delta"`
	FinishReason string  `json:"finish_reason"`
}

type Delta struct {
	Role      string     `json:"role,omitempty"`
	Content   string     `json:"content,omitempty"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

type Chunk struct {
	ID      string   `json:"id"`
	Model   string   `json:"model"`
	Created int64    `json:"created"`
	Choices []Choice `json:"choices"`
	Usage   *Usage   `json:"usage,omitempty"`
}

// Usage 兼容各家的用量字段：OpenAI 的 prompt_tokens_details.cached_tokens 与
// DeepSeek 的 prompt_cache_hit_tokens 都映射到 schema.Usage.CachedTokens。
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
	PromptCacheHit   int `json:"prompt_cache_hit_tokens"`
	PromptCacheMiss  int `json:"prompt_cache_miss_tokens"`
	PromptDetails    struct {
		CachedTokens int `json:"cached_tokens"`
	} `json:"prompt_tokens_details"`
	CompletionDetails struct {
		ReasoningTokens int `json:"reasoning_tokens"`
	} `json:"completion_tokens_details"`
}

func (u Usage) ToSchema() schema.Usage {
	cached := u.PromptDetails.CachedTokens
	if cached == 0 {
		cached = u.PromptCacheHit
	}
	return schema.Usage{
		InputTokens: u.PromptTokens, OutputTokens: u.CompletionTokens, TotalTokens: u.TotalTokens,
		CachedTokens: cached, ReasoningTokens: u.CompletionDetails.ReasoningTokens,
	}
}

func (c Completion) ToResponse() (*schema.ChatResponse, error) {
	if len(c.Choices) == 0 {
		return nil, errors.New("response contains no choices")
	}
	message, err := FromWireMessage(c.Choices[0].Message)
	if err != nil {
		return nil, err
	}
	return &schema.ChatResponse{
		ID: c.ID, Model: c.Model, Message: message, FinishReason: schema.FinishReason(c.Choices[0].FinishReason),
		Usage: c.Usage.ToSchema(), CreatedAt: time.Unix(c.Created, 0).UTC(),
	}, nil
}

// ToDeltaChunk 把标准 OpenAI chunk 归一为装配器输入；带思维链等扩展字段的
// Provider 自行实现 decode。
func (c Chunk) ToDeltaChunk() DeltaChunk {
	chunk := DeltaChunk{ID: c.ID, Model: c.Model, Created: c.Created, FinishReason: finishReason(c.Choices)}
	if c.Usage != nil {
		usage := c.Usage.ToSchema()
		chunk.Usage = &usage
	}
	for _, choice := range c.Choices {
		chunk.Content += choice.Delta.Content
		chunk.ToolCalls = append(chunk.ToolCalls, choice.Delta.ToolCalls...)
	}
	return chunk
}

func finishReason(choices []Choice) string {
	for _, choice := range choices {
		if choice.FinishReason != "" {
			return choice.FinishReason
		}
	}
	return ""
}

// ToWireMessages 把 schema 消息转为线上格式。convert 回调作用于转换结果，
// 允许 Provider 增删改消息参数（如 DeepSeek 剔除思维链），为 nil 时原样返回。
func ToWireMessages(messages []schema.Message, convert func([]Message) []Message) ([]Message, error) {
	converted := make([]Message, 0, len(messages))
	for _, message := range messages {
		item, err := ToWireMessage(message)
		if err != nil {
			return nil, err
		}
		converted = append(converted, item)
	}
	if convert != nil {
		converted = convert(converted)
	}
	return converted, nil
}

func ToWireMessage(message schema.Message) (Message, error) {
	result := Message{
		Role: string(message.Role), Name: message.Name, ToolCallID: message.ToolCallID,
	}
	if len(message.Parts) == 0 {
		result.Content = message.Content
	} else {
		parts := make([]ContentPart, 0, len(message.Parts)+1)
		if message.Content != "" {
			parts = append(parts, ContentPart{Type: "text", Text: message.Content})
		}
		for _, part := range message.Parts {
			switch part.Type {
			case schema.ContentText:
				parts = append(parts, ContentPart{Type: "text", Text: part.Text})
			case schema.ContentImageURL:
				imageURL := part.URL
				if imageURL == "" && part.Data != "" {
					imageURL = "data:" + part.MIMEType + ";base64," + part.Data
				}
				parts = append(parts, ContentPart{Type: "image_url", ImageURL: &ImageURL{URL: imageURL, Detail: part.Detail}})
			case schema.ContentInputAudio:
				format := part.MIMEType
				if format == "audio/wav" {
					format = "wav"
				}
				if format == "audio/mpeg" || format == "audio/mp3" {
					format = "mp3"
				}
				data := part.Data
				if data == "" && part.URL != "" {
					return Message{}, errors.New("chat completions audio input requires base64 data")
				}
				if _, err := base64.StdEncoding.DecodeString(data); err != nil {
					return Message{}, fmt.Errorf("invalid audio base64 data: %w", err)
				}
				parts = append(parts, ContentPart{Type: "input_audio", InputAudio: &InputAudio{Data: data, Format: format}})
			case schema.ContentFile:
				fileData := part.URL
				if fileData == "" && part.Data != "" {
					fileData = "data:" + part.MIMEType + ";base64," + part.Data
				}
				if fileData == "" {
					return Message{}, errors.New("chat completions file input requires data or a file reference")
				}
				parts = append(parts, ContentPart{Type: "file", File: &File{FileData: fileData, Filename: part.Name}})
			default:
				return Message{}, fmt.Errorf("unsupported chat completions content part: %s", part.Type)
			}
		}
		result.Content = parts
	}
	for _, call := range message.ToolCalls {
		result.ToolCalls = append(result.ToolCalls, ToolCall{
			ID: call.ID, Type: call.Type,
			Function: FunctionCall{Name: call.Function.Name, Arguments: string(call.Function.Arguments)},
		})
	}
	return result, nil
}

func FromWireMessage(message Message) (schema.Message, error) {
	result := schema.Message{
		Role: schema.Role(message.Role), Name: message.Name, ToolCallID: message.ToolCallID,
		Reasoning: message.Reasoning,
	}
	switch content := message.Content.(type) {
	case string:
		result.Content = content
	case nil:
	case map[string]any, []any:
		encoded, err := json.Marshal(content)
		if err != nil {
			return schema.Message{}, err
		}
		result.Content = string(encoded)
	default:
		return schema.Message{}, fmt.Errorf("unsupported message content type %T", content)
	}
	for _, call := range message.ToolCalls {
		arguments := json.RawMessage(call.Function.Arguments)
		if len(arguments) == 0 {
			arguments = json.RawMessage(`{}`)
		}
		result.ToolCalls = append(result.ToolCalls, schema.ToolCall{
			ID: call.ID, Type: call.Type, Function: schema.FunctionCall{Name: call.Function.Name, Arguments: arguments},
		})
	}
	return result, nil
}
