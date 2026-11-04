package openai

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"digital-labor/schema"
)

type chatRequest struct {
	Model               string         `json:"model"`
	Messages            []chatMessage  `json:"messages"`
	Tools               []chatTool     `json:"tools,omitempty"`
	ToolChoice          string         `json:"tool_choice,omitempty"`
	Stream              bool           `json:"stream,omitempty"`
	StreamOptions       *streamOptions `json:"stream_options,omitempty"`
	Temperature         *float64       `json:"temperature,omitempty"`
	MaxCompletionTokens int            `json:"max_completion_tokens,omitempty"`
	ParallelToolCalls   *bool          `json:"parallel_tool_calls,omitempty"`
}

type streamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type chatMessage struct {
	Role       string         `json:"role"`
	Content    any            `json:"content,omitempty"`
	Name       string         `json:"name,omitempty"`
	ToolCallID string         `json:"tool_call_id,omitempty"`
	ToolCalls  []chatToolCall `json:"tool_calls,omitempty"`
}

type chatContentPart struct {
	Type       string          `json:"type"`
	Text       string          `json:"text,omitempty"`
	ImageURL   *chatImageURL   `json:"image_url,omitempty"`
	InputAudio *chatInputAudio `json:"input_audio,omitempty"`
	File       *chatFile       `json:"file,omitempty"`
}

type chatImageURL struct {
	URL    string `json:"url"`
	Detail string `json:"detail,omitempty"`
}

type chatInputAudio struct {
	Data   string `json:"data"`
	Format string `json:"format"`
}

type chatFile struct {
	FileData string `json:"file_data,omitempty"`
	FileID   string `json:"file_id,omitempty"`
	Filename string `json:"filename,omitempty"`
}

type chatTool struct {
	Type     string                 `json:"type"`
	Function chatFunctionDefinition `json:"function"`
}

type chatFunctionDefinition struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters"`
}

type chatToolCall struct {
	Index    int              `json:"index,omitempty"`
	ID       string           `json:"id,omitempty"`
	Type     string           `json:"type,omitempty"`
	Function chatFunctionCall `json:"function"`
}

type chatFunctionCall struct {
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

type chatCompletion struct {
	ID      string       `json:"id"`
	Model   string       `json:"model"`
	Created int64        `json:"created"`
	Choices []chatChoice `json:"choices"`
	Usage   tokenUsage   `json:"usage"`
}

type chatChoice struct {
	Index        int         `json:"index"`
	Message      chatMessage `json:"message"`
	Delta        chatDelta   `json:"delta"`
	FinishReason string      `json:"finish_reason"`
}

type chatDelta struct {
	Role      string         `json:"role,omitempty"`
	Content   string         `json:"content,omitempty"`
	ToolCalls []chatToolCall `json:"tool_calls,omitempty"`
}

type chatChunk struct {
	ID      string       `json:"id"`
	Model   string       `json:"model"`
	Created int64        `json:"created"`
	Choices []chatChoice `json:"choices"`
	Usage   *tokenUsage  `json:"usage,omitempty"`
}

type tokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
	PromptDetails    struct {
		CachedTokens int `json:"cached_tokens"`
	} `json:"prompt_tokens_details"`
	CompletionDetails struct {
		ReasoningTokens int `json:"reasoning_tokens"`
	} `json:"completion_tokens_details"`
}

func (u tokenUsage) toUsage() schema.Usage {
	return schema.Usage{
		InputTokens: u.PromptTokens, OutputTokens: u.CompletionTokens, TotalTokens: u.TotalTokens,
		CachedTokens: u.PromptDetails.CachedTokens, ReasoningTokens: u.CompletionDetails.ReasoningTokens,
	}
}

func (c chatCompletion) toResponse() (*schema.ChatResponse, error) {
	if len(c.Choices) == 0 {
		return nil, errors.New("openai response contains no choices")
	}
	message, err := fromChatMessage(c.Choices[0].Message)
	if err != nil {
		return nil, err
	}
	return &schema.ChatResponse{
		ID: c.ID, Model: c.Model, Message: message, FinishReason: schema.FinishReason(c.Choices[0].FinishReason),
		Usage: c.Usage.toUsage(), CreatedAt: time.Unix(c.Created, 0).UTC(),
	}, nil
}

func toChatMessage(message schema.Message) (chatMessage, error) {
	result := chatMessage{
		Role: string(message.Role), Name: message.Name, ToolCallID: message.ToolCallID,
	}
	if len(message.Parts) == 0 {
		result.Content = message.Content
	} else {
		parts := make([]chatContentPart, 0, len(message.Parts)+1)
		if message.Content != "" {
			parts = append(parts, chatContentPart{Type: "text", Text: message.Content})
		}
		for _, part := range message.Parts {
			switch part.Type {
			case schema.ContentText:
				parts = append(parts, chatContentPart{Type: "text", Text: part.Text})
			case schema.ContentImageURL:
				imageURL := part.URL
				if imageURL == "" && part.Data != "" {
					imageURL = "data:" + part.MIMEType + ";base64," + part.Data
				}
				parts = append(parts, chatContentPart{Type: "image_url", ImageURL: &chatImageURL{URL: imageURL, Detail: part.Detail}})
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
					return chatMessage{}, errors.New("chat completions audio input requires base64 data")
				}
				if _, err := base64.StdEncoding.DecodeString(data); err != nil {
					return chatMessage{}, fmt.Errorf("invalid audio base64 data: %w", err)
				}
				parts = append(parts, chatContentPart{Type: "input_audio", InputAudio: &chatInputAudio{Data: data, Format: format}})
			case schema.ContentFile:
				fileData := part.URL
				if fileData == "" && part.Data != "" {
					fileData = "data:" + part.MIMEType + ";base64," + part.Data
				}
				if fileData == "" {
					return chatMessage{}, errors.New("chat completions file input requires data or a file reference")
				}
				parts = append(parts, chatContentPart{Type: "file", File: &chatFile{FileData: fileData, Filename: part.Name}})
			default:
				return chatMessage{}, fmt.Errorf("unsupported chat completions content part: %s", part.Type)
			}
		}
		result.Content = parts
	}
	for _, call := range message.ToolCalls {
		result.ToolCalls = append(result.ToolCalls, chatToolCall{
			ID: call.ID, Type: call.Type,
			Function: chatFunctionCall{Name: call.Function.Name, Arguments: string(call.Function.Arguments)},
		})
	}
	return result, nil
}

func fromChatMessage(message chatMessage) (schema.Message, error) {
	result := schema.Message{Role: schema.Role(message.Role), Name: message.Name, ToolCallID: message.ToolCallID}
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
		return schema.Message{}, fmt.Errorf("unsupported openai message content type %T", content)
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
