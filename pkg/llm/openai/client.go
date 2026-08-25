package openai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"digital-labor/pkg/llm"
)

const defaultBaseURL = "https://api.openai.com/v1"

type Config struct {
	APIKey     string
	BaseURL    string
	Model      string
	HTTPClient *http.Client
	Headers    http.Header
}

type Client struct {
	apiKey  string
	baseURL string
	model   string
	http    *http.Client
	headers http.Header

	usageMu sync.RWMutex
	usage   llm.Usage
}

func New(cfg Config) (*Client, error) {
	if strings.TrimSpace(cfg.APIKey) == "" {
		return nil, errors.New("openai api key is required")
	}
	if strings.TrimSpace(cfg.Model) == "" {
		return nil, errors.New("openai model is required")
	}
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	parsedBase, err := url.ParseRequestURI(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid openai base url: %w", err)
	}
	if strings.HasSuffix(parsedBase.Path, "/chat/completions") {
		parsedBase.Path = strings.TrimSuffix(parsedBase.Path, "/chat/completions")
		baseURL = strings.TrimRight(parsedBase.String(), "/")
	}
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 2 * time.Minute}
	}
	return &Client{
		apiKey:  cfg.APIKey,
		baseURL: baseURL,
		model:   cfg.Model,
		http:    httpClient,
		headers: cfg.Headers.Clone(),
	}, nil
}

func (c *Client) Chat(ctx context.Context, req *llm.ChatRequest) (*llm.ChatResponse, error) {
	body, err := c.requestBody(req, false)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(ctx, http.MethodPost, c.endpoint("chat/completions"), body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	var raw chatCompletion
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode openai response: %w", err)
	}
	result, err := raw.toResponse()
	if err != nil {
		return nil, err
	}
	c.addUsage(result.Usage)
	return result, nil
}

func (c *Client) ChatStream(ctx context.Context, req *llm.ChatRequest) (llm.Stream, error) {
	body, err := c.requestBody(req, true)
	if err != nil {
		return nil, err
	}
	resp, err := c.do(ctx, http.MethodPost, c.endpoint("chat/completions"), body)
	if err != nil {
		return nil, err
	}
	if err := checkResponse(resp); err != nil {
		resp.Body.Close()
		return nil, err
	}
	return newStream(resp.Body, c.addUsage), nil
}

func (c *Client) GetUsage() llm.Usage {
	c.usageMu.RLock()
	defer c.usageMu.RUnlock()
	return c.usage
}

func (c *Client) CheckStatus(ctx context.Context) (*llm.Status, error) {
	started := time.Now()
	checkedAt := time.Now().UTC()
	modelURL := c.endpoint("models/" + url.PathEscape(c.model))
	resp, err := c.do(ctx, http.MethodGet, modelURL, nil)
	if err == nil {
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			resp.Body.Close()
			return &llm.Status{Available: true, Model: c.model, Latency: time.Since(started), CheckedAt: checkedAt}, nil
		}
		resp.Body.Close()
	}

	resp, listErr := c.do(ctx, http.MethodGet, c.endpoint("models"), nil)
	if listErr != nil {
		if err != nil {
			return nil, fmt.Errorf("check openai model status: %w", err)
		}
		return nil, fmt.Errorf("check openai models status: %w", listErr)
	}
	defer resp.Body.Close()
	if statusErr := checkResponse(resp); statusErr != nil {
		return nil, statusErr
	}
	return &llm.Status{Available: true, Model: c.model, Latency: time.Since(started), CheckedAt: checkedAt}, nil
}

func (c *Client) requestBody(req *llm.ChatRequest, stream bool) ([]byte, error) {
	if req == nil {
		return nil, errors.New("chat request is required")
	}
	model := req.Model
	if model == "" {
		model = c.model
	}
	if len(req.Messages) == 0 {
		return nil, errors.New("chat messages are required")
	}
	payload := chatRequest{
		Model:             model,
		Messages:          make([]chatMessage, 0, len(req.Messages)),
		Stream:            stream,
		Temperature:       req.Temperature,
		ParallelToolCalls: req.ParallelToolCalls,
	}
	if req.MaxOutputTokens > 0 {
		payload.MaxCompletionTokens = req.MaxOutputTokens
	}
	for _, message := range req.Messages {
		converted, err := toChatMessage(message)
		if err != nil {
			return nil, err
		}
		payload.Messages = append(payload.Messages, converted)
	}
	for _, definition := range req.Tools {
		payload.Tools = append(payload.Tools, chatTool{
			Type: "function",
			Function: chatFunctionDefinition{
				Name: definition.Name, Description: definition.Description, Parameters: definition.Parameters,
			},
		})
	}
	if req.ToolChoice != "" {
		payload.ToolChoice = string(req.ToolChoice)
	}
	if stream {
		payload.StreamOptions = &streamOptions{IncludeUsage: true}
	}
	return json.Marshal(payload)
}

func (c *Client) endpoint(resource string) string {
	base, _ := url.Parse(c.baseURL)
	trimmed := strings.TrimPrefix(resource, "/")
	base.Path = path.Join(base.Path, trimmed)
	return base.String()
}

func (c *Client) do(ctx context.Context, method, endpoint string, body []byte) (*http.Response, error) {
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	for key, values := range c.headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	return c.http.Do(req)
}

func (c *Client) addUsage(usage llm.Usage) {
	c.usageMu.Lock()
	c.usage.Add(usage)
	c.usageMu.Unlock()
}

type apiError struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    any    `json:"code"`
	} `json:"error"`
}

func checkResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	data, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	var apiErr apiError
	if json.Unmarshal(data, &apiErr) == nil && apiErr.Error.Message != "" {
		return fmt.Errorf("openai api returned %s: %s", resp.Status, apiErr.Error.Message)
	}
	return fmt.Errorf("openai api returned %s: %s", resp.Status, strings.TrimSpace(string(data)))
}

type stream struct {
	body       io.ReadCloser
	scanner    *bufio.Scanner
	queue      []*llm.StreamEvent
	response   llm.ChatResponse
	toolCalls  map[int]*toolCallBuilder
	addUsage   func(llm.Usage)
	usageAdded bool
	done       bool
}

type toolCallBuilder struct {
	id        string
	typ       string
	name      string
	arguments strings.Builder
}

func newStream(body io.ReadCloser, addUsage func(llm.Usage)) *stream {
	scanner := bufio.NewScanner(body)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	return &stream{body: body, scanner: scanner, toolCalls: make(map[int]*toolCallBuilder), addUsage: addUsage}
}

func (s *stream) Recv() (*llm.StreamEvent, error) {
	if len(s.queue) > 0 {
		event := s.queue[0]
		s.queue = s.queue[1:]
		return event, nil
	}
	if s.done {
		return nil, io.EOF
	}
	for s.scanner.Scan() {
		line := strings.TrimSpace(s.scanner.Text())
		if line == "" || strings.HasPrefix(line, ":") || !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "[DONE]" {
			s.finish()
			return s.Recv()
		}
		var chunk chatChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			return nil, fmt.Errorf("decode openai stream chunk: %w", err)
		}
		s.consume(chunk)
		if len(s.queue) > 0 {
			return s.Recv()
		}
	}
	if err := s.scanner.Err(); err != nil {
		return nil, fmt.Errorf("read openai stream: %w", err)
	}
	s.finish()
	return s.Recv()
}

func (s *stream) consume(chunk chatChunk) {
	if s.response.ID == "" {
		s.response.ID = chunk.ID
		s.response.Model = chunk.Model
		if chunk.Created > 0 {
			s.response.CreatedAt = time.Unix(chunk.Created, 0).UTC()
		}
		s.response.Message.Role = llm.RoleAssistant
	}
	if chunk.Usage != nil {
		usage := chunk.Usage.toUsage()
		s.response.Usage = usage
		s.queue = append(s.queue, &llm.StreamEvent{Type: llm.StreamEventUsage, Usage: &usage})
	}
	for _, choice := range chunk.Choices {
		if choice.Delta.Content != "" {
			s.response.Message.Content += choice.Delta.Content
			s.queue = append(s.queue, &llm.StreamEvent{Type: llm.StreamEventTextDelta, Delta: choice.Delta.Content})
		}
		for _, delta := range choice.Delta.ToolCalls {
			builder := s.toolCalls[delta.Index]
			if builder == nil {
				builder = &toolCallBuilder{}
				s.toolCalls[delta.Index] = builder
			}
			if delta.ID != "" {
				builder.id = delta.ID
			}
			if delta.Type != "" {
				builder.typ = delta.Type
			}
			if delta.Function.Name != "" {
				builder.name += delta.Function.Name
			}
			builder.arguments.WriteString(delta.Function.Arguments)
			s.queue = append(s.queue, &llm.StreamEvent{Type: llm.StreamEventToolCallDelta, ToolCall: &llm.ToolCallDelta{
				Index: delta.Index, ID: delta.ID, Type: delta.Type, Name: delta.Function.Name, Arguments: delta.Function.Arguments,
			}})
		}
		if choice.FinishReason != "" {
			s.response.FinishReason = llm.FinishReason(choice.FinishReason)
		}
	}
}

func (s *stream) finish() {
	if s.done {
		return
	}
	s.done = true
	for index := 0; index < len(s.toolCalls); index++ {
		builder := s.toolCalls[index]
		if builder == nil {
			continue
		}
		arguments := json.RawMessage(builder.arguments.String())
		if len(arguments) == 0 {
			arguments = json.RawMessage(`{}`)
		}
		s.response.Message.ToolCalls = append(s.response.Message.ToolCalls, llm.ToolCall{
			ID: builder.id, Type: builder.typ, Function: llm.FunctionCall{Name: builder.name, Arguments: arguments},
		})
	}
	if !s.usageAdded {
		s.addUsage(s.response.Usage)
		s.usageAdded = true
	}
	s.queue = append(s.queue, &llm.StreamEvent{Type: llm.StreamEventDone, Response: &s.response})
}

func (s *stream) Close() error { return s.body.Close() }
