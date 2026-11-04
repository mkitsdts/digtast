package agentloop

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"digital-labor/component/chatmodel"
	"digital-labor/component/prompt"
	"digital-labor/component/tool"
	"digital-labor/schema"
)

type scriptedLLM struct {
	mu        sync.Mutex
	responses []*schema.ChatResponse
	requests  []*schema.ChatRequest
}

func (m *scriptedLLM) Chat(_ context.Context, request *schema.ChatRequest) (*schema.ChatResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.requests = append(m.requests, request)
	if len(m.responses) == 0 {
		return nil, errors.New("no scripted response")
	}
	response := m.responses[0]
	m.responses = m.responses[1:]
	return response, nil
}

func (m *scriptedLLM) ChatStream(context.Context, *schema.ChatRequest) schema.Stream {
	return schema.ErrorStream(errors.New("not implemented"))
}
func (m *scriptedLLM) GetUsage() schema.Usage { return schema.Usage{} }
func (m *scriptedLLM) CheckStatus(context.Context) (*chatmodel.Status, error) {
	return &chatmodel.Status{Available: true}, nil
}

type echoTool struct{}

func (echoTool) Definition(context.Context) (schema.ToolDefinition, error) {
	return schema.ToolDefinition{Name: "echo", Parameters: tool.ObjectSchema(map[string]any{"value": map[string]any{"type": "string"}}, "value")}, nil
}
func (echoTool) Execute(_ context.Context, arguments json.RawMessage) (*tool.Result, error) {
	var input struct {
		Value string `json:"value"`
	}
	if err := json.Unmarshal(arguments, &input); err != nil {
		return nil, err
	}
	return &tool.Result{Content: input.Value}, nil
}

type recordingHook struct {
	BaseHook
	mu     sync.Mutex
	events []string
}

func (h *recordingHook) add(event string) {
	h.mu.Lock()
	h.events = append(h.events, event)
	h.mu.Unlock()
}
func (h *recordingHook) BeforeStep(context.Context, *StepContext) error {
	h.add("before_step")
	return nil
}
func (h *recordingHook) AfterStep(context.Context, *StepContext) error {
	h.add("after_step")
	return nil
}
func (h *recordingHook) BeforeModel(context.Context, *ModelContext) error {
	h.add("before_model")
	return nil
}
func (h *recordingHook) AfterModel(context.Context, *ModelContext) error {
	h.add("after_model")
	return nil
}
func (h *recordingHook) BeforeTool(context.Context, *ToolContext) error {
	h.add("before_tool")
	return nil
}
func (h *recordingHook) AfterTool(context.Context, *ToolContext) error {
	h.add("after_tool")
	return nil
}

func TestLoopExecutesToolAndHooks(t *testing.T) {
	model := &scriptedLLM{responses: []*schema.ChatResponse{
		{Message: schema.Message{Role: schema.RoleAssistant, ToolCalls: []schema.ToolCall{{ID: "call-1", Type: "function", Function: schema.FunctionCall{Name: "echo", Arguments: json.RawMessage(`{"value":"ok"}`)}}}}, FinishReason: schema.FinishToolCalls, Usage: schema.Usage{TotalTokens: 2}},
		{Message: schema.Message{Role: schema.RoleAssistant, Content: "done"}, FinishReason: schema.FinishStop, Usage: schema.Usage{TotalTokens: 3}},
	}}
	hook := &recordingHook{}
	loop, err := New(Config{
		LLM: model, PromptLoader: prompt.LoaderFunc(func(context.Context) (string, error) { return "system prompt", nil }),
		ToolLoaders: []tool.Loader{tool.StaticLoader{echoTool{}}}, Hooks: []Hook{hook}, MaxSteps: 3,
	})
	if err != nil {
		t.Fatal(err)
	}
	result, err := loop.Run(context.Background(), Input{Messages: []schema.Message{{Role: schema.RoleUser, Content: "start"}}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Response.Message.Content != "done" || result.Steps != 2 || result.Usage.TotalTokens != 5 {
		t.Fatalf("unexpected result %#v", result)
	}
	if len(model.requests) != 2 || model.requests[0].Messages[0].Role != schema.RoleSystem {
		t.Fatalf("prompt was not loaded into model request")
	}
	secondMessages := model.requests[1].Messages
	if secondMessages[len(secondMessages)-1].Role != schema.RoleTool || secondMessages[len(secondMessages)-1].Content != "ok" {
		t.Fatalf("tool result was not appended: %#v", secondMessages)
	}
	want := []string{"before_step", "before_model", "after_model", "before_tool", "after_tool", "after_step", "before_step", "before_model", "after_model", "after_step"}
	if len(hook.events) != len(want) {
		t.Fatalf("unexpected hook events %#v", hook.events)
	}
	for index := range want {
		if hook.events[index] != want[index] {
			t.Fatalf("unexpected hook order %#v", hook.events)
		}
	}
}

// streamLLM 以迭代器方式返回预置事件，覆盖流式路径。
type streamLLM struct {
	events []*schema.StreamEvent
}

func (m *streamLLM) Chat(context.Context, *schema.ChatRequest) (*schema.ChatResponse, error) {
	return nil, errors.New("chat should not be called in stream mode")
}
func (m *streamLLM) ChatStream(_ context.Context, _ *schema.ChatRequest) schema.Stream {
	return schema.SliceStream(m.events...)
}
func (m *streamLLM) GetUsage() schema.Usage { return schema.Usage{} }
func (m *streamLLM) CheckStatus(context.Context) (*chatmodel.Status, error) {
	return &chatmodel.Status{Available: true}, nil
}

func TestLoopConsumesStreamWithReasoning(t *testing.T) {
	usage := schema.Usage{TotalTokens: 4}
	response := &schema.ChatResponse{
		Message:      schema.Message{Role: schema.RoleAssistant, Content: "答案", Reasoning: "思考"},
		FinishReason: schema.FinishStop, Usage: usage,
	}
	model := &streamLLM{events: []*schema.StreamEvent{
		{Type: schema.StreamEventReasoningDelta, Delta: "思考"},
		{Type: schema.StreamEventTextDelta, Delta: "答案"},
		{Type: schema.StreamEventUsage, Usage: &usage},
		{Type: schema.StreamEventDone, Response: response},
	}}
	loop, err := New(Config{LLM: model})
	if err != nil {
		t.Fatal(err)
	}
	var deltas []string
	result, err := loop.Run(context.Background(), Input{
		Messages: []schema.Message{{Role: schema.RoleUser, Content: "hi"}},
		Stream:   true,
		Emit: func(event Event) error {
			if event.Type == EventTextDelta || event.Type == EventReasoningDelta {
				deltas = append(deltas, event.Delta)
			}
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Response.Message.Content != "答案" || result.Response.Message.Reasoning != "思考" {
		t.Fatalf("unexpected response %#v", result.Response.Message)
	}
	if result.Usage.TotalTokens != 4 {
		t.Fatalf("unexpected usage %#v", result.Usage)
	}
	if len(deltas) != 2 || deltas[0] != "思考" || deltas[1] != "答案" {
		t.Fatalf("unexpected deltas %v", deltas)
	}
}

// chatOnlyLLM 只实现非流式接口（没有 ChatStream），流式请求应退化为 Chat。
type chatOnlyLLM struct {
	resp *schema.ChatResponse
}

func (m *chatOnlyLLM) Chat(context.Context, *schema.ChatRequest) (*schema.ChatResponse, error) {
	return m.resp, nil
}
func (m *chatOnlyLLM) GetUsage() schema.Usage { return schema.Usage{} }
func (m *chatOnlyLLM) CheckStatus(context.Context) (*chatmodel.Status, error) {
	return &chatmodel.Status{Available: true}, nil
}

func TestStreamFallsBackToChatWhenUnsupported(t *testing.T) {
	model := &chatOnlyLLM{resp: &schema.ChatResponse{
		Message: schema.Message{Role: schema.RoleAssistant, Content: "ok"}, FinishReason: schema.FinishStop,
	}}
	loop, err := New(Config{LLM: model})
	if err != nil {
		t.Fatal(err)
	}
	result, err := loop.Run(context.Background(), Input{
		Messages: []schema.Message{{Role: schema.RoleUser, Content: "hi"}},
		Stream:   true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Response.Message.Content != "ok" {
		t.Fatalf("unexpected response %#v", result.Response)
	}
}
