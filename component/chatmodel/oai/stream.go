package oai

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"digital-labor/schema"
)

// DeltaChunk 是 Provider 解码流式 chunk 后的统一增量。
type DeltaChunk struct {
	ID           string
	Model        string
	Created      int64
	Reasoning    string
	Content      string
	ToolCalls    []ToolCall
	FinishReason string
	Usage        *schema.Usage
}

// Assembler 把流式增量装配成最终响应并产出事件。
type Assembler struct {
	response  schema.ChatResponse
	toolCalls map[int]*toolCallBuilder
	done      bool
}

type toolCallBuilder struct {
	id        string
	typ       string
	name      string
	arguments strings.Builder
}

func NewAssembler() *Assembler {
	assembler := &Assembler{toolCalls: make(map[int]*toolCallBuilder)}
	assembler.response.Message.Role = schema.RoleAssistant
	return assembler
}

// Push 消费一个增量并返回相应事件。
func (a *Assembler) Push(chunk DeltaChunk) []*schema.StreamEvent {
	var events []*schema.StreamEvent
	if a.response.ID == "" {
		a.response.ID = chunk.ID
		a.response.Model = chunk.Model
		if chunk.Created > 0 {
			a.response.CreatedAt = time.Unix(chunk.Created, 0).UTC()
		}
	}
	if chunk.Usage != nil {
		a.response.Usage = *chunk.Usage
		usage := *chunk.Usage
		events = append(events, &schema.StreamEvent{Type: schema.StreamEventUsage, Usage: &usage})
	}
	if chunk.Reasoning != "" {
		a.response.Message.Reasoning += chunk.Reasoning
		events = append(events, &schema.StreamEvent{Type: schema.StreamEventReasoningDelta, Delta: chunk.Reasoning})
	}
	if chunk.Content != "" {
		a.response.Message.Content += chunk.Content
		events = append(events, &schema.StreamEvent{Type: schema.StreamEventTextDelta, Delta: chunk.Content})
	}
	for _, delta := range chunk.ToolCalls {
		builder := a.toolCalls[delta.Index]
		if builder == nil {
			builder = &toolCallBuilder{}
			a.toolCalls[delta.Index] = builder
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
		events = append(events, &schema.StreamEvent{Type: schema.StreamEventToolCallDelta, ToolCall: &schema.ToolCallDelta{
			Index: delta.Index, ID: delta.ID, Type: delta.Type, Name: delta.Function.Name, Arguments: delta.Function.Arguments,
		}})
	}
	if chunk.FinishReason != "" {
		a.response.FinishReason = schema.FinishReason(chunk.FinishReason)
	}
	return events
}

// Finish 收尾并返回 Done 事件，重复调用返回 nil。
func (a *Assembler) Finish() *schema.StreamEvent {
	if a.done {
		return nil
	}
	a.done = true
	for index := 0; index < len(a.toolCalls); index++ {
		builder := a.toolCalls[index]
		if builder == nil {
			continue
		}
		arguments := json.RawMessage(builder.arguments.String())
		if len(arguments) == 0 {
			arguments = json.RawMessage(`{}`)
		}
		a.response.Message.ToolCalls = append(a.response.Message.ToolCalls, schema.ToolCall{
			ID: builder.id, Type: builder.typ, Function: schema.FunctionCall{Name: builder.name, Arguments: arguments},
		})
	}
	response := a.response
	return &schema.StreamEvent{Type: schema.StreamEventDone, Response: &response}
}

// Chat 执行非流式请求并解码标准 OpenAI 响应。
func Chat(ctx context.Context, base *Base, build func() ([]byte, error)) (*schema.ChatResponse, error) {
	body, err := build()
	if err != nil {
		return nil, err
	}
	resp, err := base.do(ctx, http.MethodPost, "chat/completions", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if err := checkResponse(resp); err != nil {
		return nil, err
	}
	var completion Completion
	if err := json.NewDecoder(resp.Body).Decode(&completion); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	response, err := completion.ToResponse()
	if err != nil {
		return nil, err
	}
	base.AddUsage(response.Usage)
	return response, nil
}

// ChatStream 执行流式请求并返回事件迭代器。请求构建或解码失败时以
// (nil, err) 产出；迭代结束时释放底层连接并累计用量。
func ChatStream(ctx context.Context, base *Base, build func() ([]byte, error), decode func([]byte) (DeltaChunk, error)) schema.Stream {
	return func(yield func(*schema.StreamEvent, error) bool) {
		body, err := build()
		if err != nil {
			yield(nil, err)
			return
		}
		resp, err := base.do(ctx, http.MethodPost, "chat/completions", body)
		if err != nil {
			yield(nil, err)
			return
		}
		defer resp.Body.Close()
		if err := checkResponse(resp); err != nil {
			yield(nil, err)
			return
		}

		assembler := NewAssembler()
		finish := func() {
			event := assembler.Finish()
			if event == nil {
				return
			}
			base.AddUsage(event.Response.Usage)
			yield(event, nil)
		}
		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, ":") || !strings.HasPrefix(line, "data:") {
				continue
			}
			data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			if data == "[DONE]" {
				finish()
				return
			}
			chunk, err := decode([]byte(data))
			if err != nil {
				yield(nil, fmt.Errorf("decode stream chunk: %w", err))
				return
			}
			for _, event := range assembler.Push(chunk) {
				if !yield(event, nil) {
					return
				}
			}
		}
		if err := scanner.Err(); err != nil {
			yield(nil, fmt.Errorf("read stream: %w", err))
			return
		}
		finish()
	}
}
