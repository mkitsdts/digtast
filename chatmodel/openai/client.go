package openai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"digital-labor/component/chatmodel"
	"digital-labor/component/chatmodel/oai"
	"digital-labor/schema"
)

// DefaultBaseURL 是 OpenAI 官方 API 地址，可被 Config.BaseURL 替换。
const DefaultBaseURL = "https://api.openai.com/v1"

// Config 与 oai.Config 一致，New 时补充 OpenAI 默认基础配置。
type Config = oai.Config

type Client struct {
	base *oai.Base
}

func New(cfg Config) (*Client, error) {
	base, err := oai.NewBase(cfg, DefaultBaseURL)
	if err != nil {
		return nil, fmt.Errorf("openai: %w", err)
	}
	return &Client{base: base}, nil
}

func (c *Client) Chat(ctx context.Context, req *schema.ChatRequest) (*schema.ChatResponse, error) {
	return oai.Chat(ctx, c.base, func() ([]byte, error) { return c.requestBody(req, false) })
}

func (c *Client) ChatStream(ctx context.Context, req *schema.ChatRequest) schema.Stream {
	return oai.ChatStream(ctx, c.base,
		func() ([]byte, error) { return c.requestBody(req, true) },
		func(data []byte) (oai.DeltaChunk, error) {
			var chunk oai.Chunk
			if err := json.Unmarshal(data, &chunk); err != nil {
				return oai.DeltaChunk{}, err
			}
			return chunk.ToDeltaChunk(), nil
		})
}

func (c *Client) GetUsage() schema.Usage { return c.base.GetUsage() }

func (c *Client) CheckStatus(ctx context.Context) (*chatmodel.Status, error) {
	return oai.CheckStatus(ctx, c.base)
}

type chatRequest struct {
	Model               string             `json:"model"`
	Messages            []oai.Message      `json:"messages"`
	Tools               []oai.Tool         `json:"tools,omitempty"`
	ToolChoice          string             `json:"tool_choice,omitempty"`
	Stream              bool               `json:"stream,omitempty"`
	StreamOptions       *oai.StreamOptions `json:"stream_options,omitempty"`
	Temperature         *float64           `json:"temperature,omitempty"`
	MaxCompletionTokens int                `json:"max_completion_tokens,omitempty"`
	ParallelToolCalls   *bool              `json:"parallel_tool_calls,omitempty"`
}

func (c *Client) requestBody(req *schema.ChatRequest, stream bool) ([]byte, error) {
	if req == nil {
		return nil, errors.New("chat request is required")
	}
	model := req.Model
	if model == "" {
		model = c.base.Model()
	}
	if len(req.Messages) == 0 {
		return nil, errors.New("chat messages are required")
	}
	messages, err := oai.ToWireMessages(req.Messages, nil)
	if err != nil {
		return nil, err
	}
	payload := chatRequest{
		Model: model, Messages: messages, Stream: stream,
		Temperature: req.Temperature, ParallelToolCalls: req.ParallelToolCalls,
	}
	if req.MaxOutputTokens > 0 {
		payload.MaxCompletionTokens = req.MaxOutputTokens
	}
	for _, definition := range req.Tools {
		payload.Tools = append(payload.Tools, oai.Tool{
			Type: "function",
			Function: oai.FunctionDefinition{
				Name: definition.Name, Description: definition.Description, Parameters: definition.Parameters,
			},
		})
	}
	if req.ToolChoice != "" {
		payload.ToolChoice = string(req.ToolChoice)
	}
	if stream {
		payload.StreamOptions = &oai.StreamOptions{IncludeUsage: true}
	}
	return json.Marshal(payload)
}
